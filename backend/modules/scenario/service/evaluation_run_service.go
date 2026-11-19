package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/gorm"
)

var (
	ErrEvaluationRunForbidden = errors.New(
		"only Quality Assurance users may run evaluations",
	)

	ErrEvaluationRunActive = errors.New(
		"another evaluation run is still in progress",
	)

	ErrEvaluationRunNotFound = errors.New("evaluation run not found")

	ErrEvaluationRunEmptyDataset = errors.New(
		"golden dataset has no items to evaluate",
	)
)

// EvaluationClient is the subset of the AI-service client used by a run.
type EvaluationClient interface {
	Generate(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error)
	GenerateRuleBased(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error)
	Validate(context.Context, string, []string) (dto.ScenarioValidation, error)
}

type EvaluationRunService interface {
	Start(ctx context.Context, userID, datasetID string) (dto.EvaluationRunCreated, error)
	Progress(ctx context.Context, userID, runID string) (dto.EvaluationRunProgress, error)
}

type evaluationRunService struct {
	runs     repository.EvaluationRunRepository
	datasets repository.GoldenDatasetRepository
	users    ReviewUserRepository
	client   EvaluationClient
}

func NewEvaluationRunService(
	runs repository.EvaluationRunRepository,
	datasets repository.GoldenDatasetRepository,
	users ReviewUserRepository,
	client EvaluationClient,
) EvaluationRunService {
	return &evaluationRunService{
		runs:     runs,
		datasets: datasets,
		users:    users,
		client:   client,
	}
}

func (s *evaluationRunService) authorizeQA(ctx context.Context, userID string) (string, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return "", ErrEvaluationRunForbidden
	}

	user, err := s.users.GetUserById(ctx, nil, id.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrEvaluationRunForbidden
	}
	if err != nil {
		return "", err
	}
	if user.Role != entities.RoleQA {
		return "", ErrEvaluationRunForbidden
	}

	return user.ID.String(), nil
}

func (s *evaluationRunService) Start(
	ctx context.Context,
	userID, datasetID string,
) (dto.EvaluationRunCreated, error) {
	createdBy, err := s.authorizeQA(ctx, userID)
	if err != nil {
		return dto.EvaluationRunCreated{}, err
	}

	datasetUUID, err := uuid.Parse(datasetID)
	if err != nil {
		return dto.EvaluationRunCreated{}, ErrEvaluationRunNotFound
	}

	dataset, err := s.datasets.GetByID(ctx, datasetUUID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.EvaluationRunCreated{}, ErrEvaluationRunNotFound
	}
	if err != nil {
		return dto.EvaluationRunCreated{}, err
	}
	if len(dataset.Items) == 0 {
		return dto.EvaluationRunCreated{}, ErrEvaluationRunEmptyDataset
	}

	active, err := s.runs.HasActive(ctx)
	if err != nil {
		return dto.EvaluationRunCreated{}, err
	}
	if active {
		return dto.EvaluationRunCreated{}, ErrEvaluationRunActive
	}

	run := &entities.EvaluationRun{
		ID:              uuid.New(),
		GoldenDatasetID: &dataset.ID,
		DatasetName:     dataset.Name,
		CreatedBy:       createdBy,
		TotalCases:      len(dataset.Items),
		Status:          entities.EvaluationRunQueued,
	}

	cases := make([]entities.EvaluationRunCase, len(dataset.Items))
	for index, item := range dataset.Items {
		cases[index] = entities.EvaluationRunCase{
			ID:              uuid.New(),
			CaseID:          item.ID.String(),
			Position:        index + 1,
			Title:           item.ItemTitle,
			ExpectedGherkin: item.ExpectedGherkin,
		}
	}

	if err := s.runs.CreateWithCases(ctx, run, cases); err != nil {
		return dto.EvaluationRunCreated{}, err
	}

	items := append([]entities.GoldenDatasetItem(nil), dataset.Items...)
	go s.execute(run.ID, cases, items)

	return dto.EvaluationRunCreated{
		RunID:      run.ID.String(),
		ReportCode: fmt.Sprintf("ER-%04d", run.RunNumber),
		Status:     run.Status,
		TotalCases: run.TotalCases,
	}, nil
}

func (s *evaluationRunService) Progress(
	ctx context.Context,
	userID, runID string,
) (dto.EvaluationRunProgress, error) {
	if _, err := s.authorizeQA(ctx, userID); err != nil {
		return dto.EvaluationRunProgress{}, err
	}

	id, err := uuid.Parse(runID)
	if err != nil {
		return dto.EvaluationRunProgress{}, ErrEvaluationRunNotFound
	}

	progress, err := s.runs.GetProgress(ctx, id)
	if errors.Is(err, repository.ErrEvaluationRunNotFound) {
		return dto.EvaluationRunProgress{}, ErrEvaluationRunNotFound
	}

	return progress, err
}

func (s *evaluationRunService) execute(
	runID uuid.UUID,
	cases []entities.EvaluationRunCase,
	items []entities.GoldenDatasetItem,
) {
	ctx := context.Background()

	if err := s.runs.MarkRunning(ctx, runID); err != nil {
		s.fail(ctx, runID, "mark_running", err)
		return
	}

	metrics, err := s.runCases(ctx, cases, items)
	if err != nil {
		s.fail(ctx, runID, "save_case_result", err)
		return
	}

	if err := s.runs.Finish(ctx, runID, entities.EvaluationRunCompleted, &metrics, nil); err != nil {
		s.fail(ctx, runID, "finish", err)
	}
}

func (s *evaluationRunService) fail(ctx context.Context, runID uuid.UUID, operation string, cause error) {
	slog.Error("Evaluation run failed", "run_id", runID, "operation", operation, "error", cause)
	message := "Evaluation failed due to an internal error."
	if err := s.runs.Finish(ctx, runID, entities.EvaluationRunFailed, nil, &message); err != nil {
		slog.Error("Could not persist failed evaluation status", "run_id", runID, "error", err)
	}
}

func (s *evaluationRunService) runCases(
	ctx context.Context,
	cases []entities.EvaluationRunCase,
	items []entities.GoldenDatasetItem,
) (entities.EvaluationRunMetrics, error) {
	llmResults := make([]*entities.EvaluationCaseMethodResult, 0, len(items))
	ruleBasedResults := make([]*entities.EvaluationCaseMethodResult, 0, len(items))

	for index, item := range items {
		request := dto.GenerateRequest{
			UserStory:    item.UserStoryText,
			GuiElements:  item.GUIElements,
			FreeScenario: item.FreeScenarioText,
		}

		llm := s.evaluate(ctx, item.GUIElements, func() (dto.GenerateResponse, error) {
			return s.client.Generate(ctx, request)
		})
		ruleBased := s.evaluate(ctx, item.GUIElements, func() (dto.GenerateResponse, error) {
			return s.client.GenerateRuleBased(ctx, request)
		})

		llmResults = append(llmResults, llm)
		ruleBasedResults = append(ruleBasedResults, ruleBased)

		if err := s.runs.SaveCaseResult(ctx, cases[index].ID, llm, ruleBased); err != nil {
			return entities.EvaluationRunMetrics{}, fmt.Errorf("save case %s: %w", cases[index].ID, err)
		}
	}

	return entities.EvaluationRunMetrics{
		LLM:       summarize(llmResults, len(items)),
		RuleBased: summarize(ruleBasedResults, len(items)),
	}, nil
}

func (s *evaluationRunService) evaluate(
	ctx context.Context,
	elements []string,
	generate func() (dto.GenerateResponse, error),
) *entities.EvaluationCaseMethodResult {
	generated, err := generate()
	if err != nil {
		return nil
	}

	validation, err := s.client.Validate(ctx, generated.Gherkin, elements)
	if err != nil {
		return nil
	}

	return caseResult(validation)
}

func caseResult(validation dto.ScenarioValidation) *entities.EvaluationCaseMethodResult {
	coverage := round2(validation.CoveragePercent)
	result := &entities.EvaluationCaseMethodResult{
		Status:             entities.EvaluationCaseFail,
		GUICoveragePercent: &coverage,
	}

	if validation.Valid {
		result.Status = entities.EvaluationCaseWarn
		if coverage >= 100 {
			result.Status = entities.EvaluationCasePass
		}
	}

	return result
}

func summarize(
	results []*entities.EvaluationCaseMethodResult,
	total int,
) entities.EvaluationMethodMetrics {
	if total == 0 {
		return entities.EvaluationMethodMetrics{}
	}

	valid, coverageSum, coverageCount := 0, 0.0, 0

	for _, result := range results {
		if result == nil {
			continue
		}
		if result.Status != entities.EvaluationCaseFail {
			valid++
		}
		if result.GUICoveragePercent != nil {
			coverageSum += *result.GUICoveragePercent
			coverageCount++
		}
	}

	validation := round2(float64(valid) / float64(total) * 100)
	metrics := entities.EvaluationMethodMetrics{ValidationPercent: &validation}

	if coverageCount > 0 {
		coverage := round2(coverageSum / float64(coverageCount))
		metrics.GUICoveragePercent = &coverage
	}

	return metrics
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
