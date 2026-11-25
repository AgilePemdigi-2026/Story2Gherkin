package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"unicode"

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
			ID:                 uuid.New(),
			CaseID:             item.ID.String(),
			Position:           index + 1,
			Title:              item.ItemTitle,
			ExpectedGherkin:    item.ExpectedGherkin,
			AcceptanceCriteria: append([]string(nil), item.AcceptanceCriteria...),
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

		llm := s.evaluate(ctx, item.GUIElements, item.AcceptanceCriteria, func() (dto.GenerateResponse, error) {
			return s.client.Generate(ctx, request)
		})
		ruleBased := s.evaluate(ctx, item.GUIElements, item.AcceptanceCriteria, func() (dto.GenerateResponse, error) {
			return s.client.GenerateRuleBased(ctx, request)
		})

		llmResults = append(llmResults, llm)
		ruleBasedResults = append(ruleBasedResults, ruleBased)

		if err := s.runs.SaveCaseResult(ctx, cases[index].ID, llm, ruleBased); err != nil {
			return entities.EvaluationRunMetrics{}, fmt.Errorf("save case %s: %w", cases[index].ID, err)
		}
	}

	return entities.EvaluationRunMetrics{
		LLM:       summarize(llmResults, len(items), acceptanceCriteriaTotal(items)),
		RuleBased: summarize(ruleBasedResults, len(items), acceptanceCriteriaTotal(items)),
	}, nil
}

func (s *evaluationRunService) evaluate(
	ctx context.Context,
	elements []string,
	criteria []string,
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

	return caseResult(validation, generated.Gherkin, criteria)
}

func caseResult(
	validation dto.ScenarioValidation,
	gherkin string,
	criteria []string,
) *entities.EvaluationCaseMethodResult {
	coverage := round2(validation.CoveragePercent)
	result := &entities.EvaluationCaseMethodResult{
		Status:             entities.EvaluationCaseFail,
		GUICoveragePercent: &coverage,
	}

	if len(criteria) > 0 {
		score := scoreAcceptanceCriteria(gherkin, criteria)
		if !validation.Valid {
			score = acceptanceCriteriaScore{total: len(criteria)}
		}
		result.AcceptanceCriteriaMatched = score.matched
		result.AcceptanceCriteriaTotal = score.total
		result.GeneratedBehaviorMatched = score.generatedMatched
		result.GeneratedBehaviorTotal = score.generatedTotal
		result.AcceptanceCriteriaPrecisionPercent = float64Pointer(score.precision())
		result.AcceptanceCriteriaRecallPercent = float64Pointer(score.recall())
		result.AcceptanceCriteriaF1Percent = float64Pointer(score.f1())
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
	criteriaTotal int,
) entities.EvaluationMethodMetrics {
	if total == 0 {
		return entities.EvaluationMethodMetrics{}
	}

	valid, coverageSum, coverageCount := 0, 0.0, 0
	matchedCriteria, matchedBehaviors, generatedBehaviors := 0, 0, 0

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
		matchedCriteria += result.AcceptanceCriteriaMatched
		matchedBehaviors += result.GeneratedBehaviorMatched
		generatedBehaviors += result.GeneratedBehaviorTotal
	}

	validation := round2(float64(valid) / float64(total) * 100)
	metrics := entities.EvaluationMethodMetrics{ValidationPercent: &validation}

	if coverageCount > 0 {
		coverage := round2(coverageSum / float64(coverageCount))
		metrics.GUICoveragePercent = &coverage
	}
	if criteriaTotal > 0 {
		precision, recall := 0.0, round2(float64(matchedCriteria)/float64(criteriaTotal)*100)
		if generatedBehaviors > 0 {
			precision = round2(float64(matchedBehaviors) / float64(generatedBehaviors) * 100)
		}
		f1 := 0.0
		if precision+recall > 0 {
			f1 = round2(2 * precision * recall / (precision + recall))
		}
		metrics.AcceptanceCriteriaPrecisionPercent = &precision
		metrics.AcceptanceCriteriaRecallPercent = &recall
		metrics.AcceptanceCriteriaF1Percent = &f1
	}

	return metrics
}

func acceptanceCriteriaTotal(items []entities.GoldenDatasetItem) int {
	total := 0
	for _, item := range items {
		total += len(item.AcceptanceCriteria)
	}
	return total
}

type acceptanceCriteriaScore struct {
	matched          int
	total            int
	generatedMatched int
	generatedTotal   int
}

func (s acceptanceCriteriaScore) precision() float64 {
	if s.generatedTotal == 0 {
		return 0
	}
	return round2(float64(s.generatedMatched) / float64(s.generatedTotal) * 100)
}

func (s acceptanceCriteriaScore) recall() float64 {
	if s.total == 0 {
		return 0
	}
	return round2(float64(s.matched) / float64(s.total) * 100)
}

func (s acceptanceCriteriaScore) f1() float64 {
	precision, recall := s.precision(), s.recall()
	if precision+recall == 0 {
		return 0
	}
	return round2(2 * precision * recall / (precision + recall))
}

func float64Pointer(value float64) *float64 {
	value = round2(value)
	return &value
}

func scoreAcceptanceCriteria(gherkin string, criteria []string) acceptanceCriteriaScore {
	steps := gherkinBehaviorSteps(gherkin)
	score := acceptanceCriteriaScore{total: len(criteria), generatedTotal: len(steps)}

	for _, step := range steps {
		matchesCriterion := false
		for _, criterion := range criteria {
			if acceptanceCriterionMatches(criterion, step) {
				matchesCriterion = true
			}
		}
		if matchesCriterion {
			score.generatedMatched++
		}
	}

	for _, criterion := range criteria {
		for _, step := range steps {
			if acceptanceCriterionMatches(criterion, step) {
				score.matched++
				break
			}
		}
	}

	return score
}

func gherkinBehaviorSteps(gherkin string) []string {
	steps := []string{}
	phase := ""
	for _, line := range strings.Split(gherkin, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		keyword := strings.ToLower(strings.TrimSuffix(fields[0], ":"))
		text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), fields[0]))
		switch keyword {
		case "given":
			phase = "given"
		case "when", "then":
			phase = keyword
			steps = append(steps, text)
		case "and", "but":
			if phase == "when" || phase == "then" {
				steps = append(steps, text)
			}
		}
	}
	return steps
}

var acceptanceStopWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "are": {}, "as": {}, "at": {}, "be": {}, "by": {},
	"dengan": {}, "di": {}, "dapat": {}, "for": {}, "from": {}, "in": {}, "is": {},
	"ke": {}, "of": {}, "on": {}, "pada": {}, "sistem": {}, "system": {}, "the": {},
	"to": {}, "user": {}, "pengguna": {}, "yang": {},
}

func acceptanceCriterionMatches(criterion, step string) bool {
	target, candidate := meaningfulTerms(criterion), meaningfulTerms(step)
	if len(target) == 0 || len(candidate) == 0 {
		return false
	}

	matches := 0
	for term := range target {
		if _, exists := candidate[term]; exists {
			matches++
		}
	}

	// shortcut: lexical matching does not identify synonyms, upgrade to QA-reviewed criterion aliases when the dataset grows.
	return float64(matches)/float64(len(target)) >= 0.6
}

func meaningfulTerms(text string) map[string]struct{} {
	terms := make(map[string]struct{})
	for _, token := range strings.FieldsFunc(strings.ToLower(text), func(char rune) bool {
		return !unicode.IsLetter(char) && !unicode.IsNumber(char)
	}) {
		if _, ignored := acceptanceStopWords[token]; !ignored {
			terms[token] = struct{}{}
		}
	}
	return terms
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
