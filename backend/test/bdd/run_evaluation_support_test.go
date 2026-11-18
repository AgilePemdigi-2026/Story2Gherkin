package bdd_test

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
)

// ------------------------------------
// FAKE EVALUATION RUN REPOSITORY
// ------------------------------------

type evalBDDRunRepo struct {
	mu        sync.Mutex
	runs      map[uuid.UUID]*entities.EvaluationRun
	caseRun   map[uuid.UUID]uuid.UUID
	llm       map[uuid.UUID]*entities.EvaluationCaseMethodResult
	ruleBased map[uuid.UUID]*entities.EvaluationCaseMethodResult
	processed map[uuid.UUID]bool
	created   int
}

func newEvalBDDRunRepo() *evalBDDRunRepo {
	return &evalBDDRunRepo{
		runs:      make(map[uuid.UUID]*entities.EvaluationRun),
		caseRun:   make(map[uuid.UUID]uuid.UUID),
		llm:       make(map[uuid.UUID]*entities.EvaluationCaseMethodResult),
		ruleBased: make(map[uuid.UUID]*entities.EvaluationCaseMethodResult),
		processed: make(map[uuid.UUID]bool),
	}
}

func (r *evalBDDRunRepo) HasActive(context.Context) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, run := range r.runs {
		if run.Status == entities.EvaluationRunQueued || run.Status == entities.EvaluationRunRunning {
			return true, nil
		}
	}

	return false, nil
}

func (r *evalBDDRunRepo) CreateWithCases(
	_ context.Context,
	run *entities.EvaluationRun,
	cases []entities.EvaluationRunCase,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.created++
	run.RunNumber = int64(r.created)
	run.Status = entities.EvaluationRunQueued

	r.runs[run.ID] = run

	for _, item := range cases {
		r.caseRun[item.ID] = run.ID
	}

	return nil
}

func (r *evalBDDRunRepo) MarkRunning(_ context.Context, runID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if run, exists := r.runs[runID]; exists {
		run.Status = entities.EvaluationRunRunning
	}

	return nil
}

func (r *evalBDDRunRepo) SaveCaseResult(
	_ context.Context,
	caseID uuid.UUID,
	llm, ruleBased *entities.EvaluationCaseMethodResult,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.llm[caseID] = llm
	r.ruleBased[caseID] = ruleBased
	r.processed[caseID] = true

	return nil
}

func (r *evalBDDRunRepo) GetProgress(
	_ context.Context,
	runID uuid.UUID,
) (dto.EvaluationRunProgress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, exists := r.runs[runID]
	if !exists {
		return dto.EvaluationRunProgress{}, repository.ErrEvaluationRunNotFound
	}

	processed := 0

	for caseID, owner := range r.caseRun {
		if owner == runID && r.processed[caseID] {
			processed++
		}
	}

	return dto.EvaluationRunProgress{
		RunID:          runID.String(),
		Status:         run.Status,
		TotalCases:     run.TotalCases,
		ProcessedCases: processed,
		ErrorMessage:   run.ErrorMessage,
	}, nil
}

func (r *evalBDDRunRepo) Finish(
	_ context.Context,
	runID uuid.UUID,
	status entities.EvaluationRunStatus,
	metrics *entities.EvaluationRunMetrics,
	message *string,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if run, exists := r.runs[runID]; exists {
		run.Status = status
		run.Metrics = metrics
		run.ErrorMessage = message
	}

	return nil
}

// snapshot returns a copy of a stored run for assertions.
func (r *evalBDDRunRepo) snapshot(runID uuid.UUID) (entities.EvaluationRun, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, exists := r.runs[runID]

	if !exists {
		return entities.EvaluationRun{}, false
	}

	return *run, true
}

func (r *evalBDDRunRepo) onlyRun() (entities.EvaluationRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.runs) != 1 {
		return entities.EvaluationRun{}, errEvalFixture("run count", len(r.runs), 1)
	}

	for _, run := range r.runs {
		return *run, nil
	}

	return entities.EvaluationRun{}, errEvalFixture("run count", 0, 1)
}

func (r *evalBDDRunRepo) caseResult(
	runID uuid.UUID,
	method string,
) (*entities.EvaluationCaseMethodResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for caseID, owner := range r.caseRun {
		if owner != runID {
			continue
		}

		if method == "rule_based" {
			return r.ruleBased[caseID], nil
		}

		return r.llm[caseID], nil
	}

	return nil, errEvalFixture("stored cases", 0, 1)
}

func (r *evalBDDRunRepo) runCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.runs)
}

func (r *evalBDDRunRepo) createdCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.created
}

// ------------------------------------
// FAKE EVALUATION CLIENT
// ------------------------------------

type evalBDDClient struct {
	mu       sync.Mutex
	valid    bool
	coverage float64
	calls    int
}

func (c *evalBDDClient) generate() (dto.GenerateResponse, error) {
	c.mu.Lock()
	c.calls++
	c.mu.Unlock()

	return dto.GenerateResponse{
		Gherkin: "Feature: BDD Evaluation\n" +
			"  Scenario: Evaluated Flow\n" +
			"    Given a login page\n" +
			"    When valid credentials are entered\n" +
			"    Then the dashboard is displayed",
		Feature: dto.Feature{
			FeatureName: "BDD Evaluation",
			Scenarios: []dto.Scenario{
				{Name: "Evaluated Flow", Kind: "positive"},
			},
		},
	}, nil
}

func (c *evalBDDClient) Generate(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
	return c.generate()
}

func (c *evalBDDClient) GenerateRuleBased(
	context.Context,
	dto.GenerateRequest,
) (dto.GenerateResponse, error) {
	return c.generate()
}

func (c *evalBDDClient) Validate(
	_ context.Context,
	_ string,
	elements []string,
) (dto.ScenarioValidation, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return dto.ScenarioValidation{
		Valid:              c.valid,
		FeatureName:        "BDD Evaluation",
		UsedElements:       append([]string(nil), elements...),
		NotCoveredElements: []string{},
		CoveragePercent:    c.coverage,
	}, nil
}

// ------------------------------------
// ROUTER
// ------------------------------------

func evalBDDRouter(
	runs repository.EvaluationRunRepository,
	datasets repository.GoldenDatasetRepository,
	users service.ReviewUserRepository,
	client service.EvaluationClient,
	jwt authService.JWTService,
) *gin.Engine {
	runService := service.NewEvaluationRunService(runs, datasets, users, client)
	runHandler := handler.NewEvaluationRunHandler(runService)

	router := gin.New()
	scenario.RegisterEvaluationRunRoutes(router.Group("/api/v1"), jwt, runHandler)

	return router
}

// seed stores a run with the given total/processed counts.
func (r *evalBDDRunRepo) seed(
	status entities.EvaluationRunStatus,
	total, processed int,
) uuid.UUID {
	r.mu.Lock()
	defer r.mu.Unlock()

	runID := uuid.New()

	run := &entities.EvaluationRun{
		ID:          runID,
		DatasetName: "BDD Evaluation Dataset",
		TotalCases:  total,
		Status:      status,
	}

	r.runs[runID] = run

	for index := 0; index < total; index++ {
		caseID := uuid.New()
		r.caseRun[caseID] = runID

		if index < processed {
			r.processed[caseID] = true
		}
	}

	return runID
}
