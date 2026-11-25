package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/gorm"
)

type caseSave struct {
	llm, ruleBased *entities.EvaluationCaseMethodResult
}

type fakeRunRepo struct {
	active   bool
	saved    []caseSave
	finished entities.EvaluationRunStatus
	metrics  *entities.EvaluationRunMetrics
	message  *string
	saveErr  error
}

func (f *fakeRunRepo) HasActive(context.Context) (bool, error) { return f.active, nil }

func (f *fakeRunRepo) CreateWithCases(context.Context, *entities.EvaluationRun, []entities.EvaluationRunCase) error {
	return nil
}

func (f *fakeRunRepo) MarkRunning(context.Context, uuid.UUID) error { return nil }

func (f *fakeRunRepo) SaveCaseResult(_ context.Context, _ uuid.UUID, llm, ruleBased *entities.EvaluationCaseMethodResult) error {
	f.saved = append(f.saved, caseSave{llm: llm, ruleBased: ruleBased})
	return f.saveErr
}

func (f *fakeRunRepo) GetProgress(context.Context, uuid.UUID) (dto.EvaluationRunProgress, error) {
	return dto.EvaluationRunProgress{}, nil
}

func (f *fakeRunRepo) Finish(_ context.Context, _ uuid.UUID, status entities.EvaluationRunStatus, metrics *entities.EvaluationRunMetrics, message *string) error {
	f.finished, f.metrics, f.message = status, metrics, message
	return nil
}

type fakeEvalClient struct {
	generate func(dto.GenerateRequest) (dto.GenerateResponse, error)
	rule     func(dto.GenerateRequest) (dto.GenerateResponse, error)
	validate func(string) dto.ScenarioValidation
}

func (f *fakeEvalClient) Generate(_ context.Context, request dto.GenerateRequest) (dto.GenerateResponse, error) {
	return f.generate(request)
}

func (f *fakeEvalClient) GenerateRuleBased(_ context.Context, request dto.GenerateRequest) (dto.GenerateResponse, error) {
	if f.rule == nil {
		return dto.GenerateResponse{Gherkin: "rule"}, nil
	}
	return f.rule(request)
}

func (f *fakeEvalClient) Validate(_ context.Context, text string, _ []string) (dto.ScenarioValidation, error) {
	return f.validate(text), nil
}

type fakeDatasetRepo struct {
	dataset *entities.GoldenDataset
}

func (f *fakeDatasetRepo) CreateWithItems(context.Context, *entities.GoldenDataset, []entities.GoldenDatasetItem) error {
	return nil
}
func (f *fakeDatasetRepo) UpdateWithItems(context.Context, uuid.UUID, string, *string, []entities.GoldenDatasetItem) error {
	return nil
}
func (f *fakeDatasetRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (f *fakeDatasetRepo) List(context.Context, string, int, int) ([]entities.GoldenDataset, int64, error) {
	return nil, 0, nil
}
func (f *fakeDatasetRepo) GetByID(context.Context, uuid.UUID) (*entities.GoldenDataset, error) {
	if f.dataset == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return f.dataset, nil
}

type fakeUsers struct {
	user entities.User
	err  error
}

func (f fakeUsers) GetUserById(context.Context, *gorm.DB, string) (entities.User, error) {
	return f.user, f.err
}

func validClient(valid map[string]bool) *fakeEvalClient {
	return &fakeEvalClient{
		generate: func(request dto.GenerateRequest) (dto.GenerateResponse, error) {
			if valid[request.UserStory] {
				return dto.GenerateResponse{Gherkin: request.UserStory}, nil
			}
			return dto.GenerateResponse{Gherkin: request.UserStory}, nil
		},
		validate: func(text string) dto.ScenarioValidation {
			if valid[text] {
				return dto.ScenarioValidation{Valid: true, CoveragePercent: 100}
			}
			return dto.ScenarioValidation{Valid: false, CoveragePercent: 50}
		},
	}
}

func testItems() []entities.GoldenDatasetItem {
	return []entities.GoldenDatasetItem{
		{ItemTitle: "A", UserStoryText: "story-a-long", GUIElements: []string{"X"}, FreeScenarioText: "free-a-long"},
		{ItemTitle: "B", UserStoryText: "story-b-long", GUIElements: []string{"Y"}, FreeScenarioText: "free-b-long"},
	}
}

func testCases() []entities.EvaluationRunCase {
	return []entities.EvaluationRunCase{{ID: uuid.New()}, {ID: uuid.New()}}
}

func TestRunCasesComputesStatusesAndMetrics(t *testing.T) {
	repo := &fakeRunRepo{}
	client := validClient(map[string]bool{"story-a-long": true, "rule": true})
	svc := &evaluationRunService{runs: repo, client: client}

	metrics, err := svc.runCases(context.Background(), testCases(), testItems())
	if err != nil {
		t.Fatal(err)
	}

	if *metrics.LLM.ValidationPercent != 50 {
		t.Fatalf("llm validity = %v", *metrics.LLM.ValidationPercent)
	}
	if *metrics.RuleBased.ValidationPercent != 100 {
		t.Fatalf("rule validity = %v", *metrics.RuleBased.ValidationPercent)
	}
	if repo.saved[0].llm.Status != entities.EvaluationCasePass {
		t.Fatalf("case 1 llm status = %s", repo.saved[0].llm.Status)
	}
	if repo.saved[1].llm.Status != entities.EvaluationCaseFail {
		t.Fatalf("case 2 llm status = %s", repo.saved[1].llm.Status)
	}
}

func TestWarnWhenCoverageBelowFull(t *testing.T) {
	repo := &fakeRunRepo{}
	client := &fakeEvalClient{
		generate: func(dto.GenerateRequest) (dto.GenerateResponse, error) {
			return dto.GenerateResponse{Gherkin: "g"}, nil
		},
		validate: func(string) dto.ScenarioValidation {
			return dto.ScenarioValidation{Valid: true, CoveragePercent: 66.666}
		},
	}
	svc := &evaluationRunService{runs: repo, client: client}

	if _, err := svc.runCases(context.Background(), testCases()[:1], testItems()[:1]); err != nil {
		t.Fatal(err)
	}

	if repo.saved[0].llm.Status != entities.EvaluationCaseWarn {
		t.Fatalf("status = %s", repo.saved[0].llm.Status)
	}
	if got := *repo.saved[0].llm.GUICoveragePercent; got != 66.67 {
		t.Fatalf("coverage = %v", got)
	}
}

func TestRunCasesSummarizesAcceptanceCriteria(t *testing.T) {
	repo := &fakeRunRepo{}
	gherkin := "Feature: Login\n" +
		"  Scenario: Successful login\n" +
		"    Given the login page is open\n" +
		"    When valid credentials are entered\n" +
		"    Then the dashboard is displayed\n"
	client := &fakeEvalClient{
		generate: func(dto.GenerateRequest) (dto.GenerateResponse, error) {
			return dto.GenerateResponse{Gherkin: gherkin}, nil
		},
		rule: func(dto.GenerateRequest) (dto.GenerateResponse, error) {
			return dto.GenerateResponse{Gherkin: gherkin}, nil
		},
		validate: func(string) dto.ScenarioValidation {
			return dto.ScenarioValidation{Valid: true, CoveragePercent: 100}
		},
	}
	items := []entities.GoldenDatasetItem{{
		ItemTitle:        "Login",
		UserStoryText:    "login-story",
		GUIElements:      []string{"Login button"},
		FreeScenarioText: "valid login behavior",
		AcceptanceCriteria: []string{
			"valid credentials are entered",
			"dashboard is displayed",
		},
	}}

	metrics, err := (&evaluationRunService{runs: repo, client: client}).runCases(
		context.Background(),
		[]entities.EvaluationRunCase{{ID: uuid.New()}},
		items,
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, method := range []entities.EvaluationMethodMetrics{
		metrics.LLM,
		metrics.RuleBased,
	} {
		if method.AcceptanceCriteriaPrecisionPercent == nil ||
			method.AcceptanceCriteriaRecallPercent == nil ||
			method.AcceptanceCriteriaF1Percent == nil ||
			*method.AcceptanceCriteriaPrecisionPercent != 100 ||
			*method.AcceptanceCriteriaRecallPercent != 100 ||
			*method.AcceptanceCriteriaF1Percent != 100 {
			t.Fatalf("unexpected acceptance criteria metrics: %+v", method)
		}
	}
}

func TestScoreAcceptanceCriteriaPenalizesUnmatchedBehavior(t *testing.T) {
	score := scoreAcceptanceCriteria(
		"Feature: Login\n"+
			"  Scenario: Successful login\n"+
			"    Given the login page is open\n"+
			"    When valid credentials are entered\n"+
			"    Then the dashboard is displayed\n"+
			"    And a promotional banner is shown\n",
		[]string{
			"valid credentials are entered",
			"dashboard is displayed",
		},
	)

	if score.precision() != 66.67 || score.recall() != 100 || score.f1() != 80 {
		t.Fatalf("score = precision %v, recall %v, F1 %v", score.precision(), score.recall(), score.f1())
	}
}

func TestCaseFailureDoesNotStopRun(t *testing.T) {
	repo := &fakeRunRepo{}
	client := &fakeEvalClient{
		generate: func(request dto.GenerateRequest) (dto.GenerateResponse, error) {
			if request.UserStory == "story-a-long" {
				return dto.GenerateResponse{}, errors.New("llm down")
			}
			return dto.GenerateResponse{Gherkin: "ok"}, nil
		},
		validate: func(string) dto.ScenarioValidation {
			return dto.ScenarioValidation{Valid: true, CoveragePercent: 100}
		},
	}
	svc := &evaluationRunService{runs: repo, client: client}

	svc.execute(uuid.New(), testCases(), testItems())

	if len(repo.saved) != 2 || repo.saved[0].llm != nil {
		t.Fatalf("failed case must be saved with nil LLM: %+v", repo.saved)
	}
	if repo.finished != entities.EvaluationRunCompleted {
		t.Fatalf("status = %s", repo.finished)
	}
	if *repo.metrics.LLM.ValidationPercent != 50 {
		t.Fatalf("llm validity = %v", *repo.metrics.LLM.ValidationPercent)
	}
}

func TestStartRejectsWhenActiveRunExists(t *testing.T) {
	dataset := &entities.GoldenDataset{ID: uuid.New(), Name: "D", Items: testItems()}
	svc := &evaluationRunService{
		runs:     &fakeRunRepo{active: true},
		datasets: &fakeDatasetRepo{dataset: dataset},
		users:    fakeUsers{user: entities.User{ID: uuid.New(), Role: entities.RoleQA}},
	}

	_, err := svc.Start(context.Background(), uuid.NewString(), dataset.ID.String())

	if !errors.Is(err, ErrEvaluationRunActive) {
		t.Fatalf("err = %v", err)
	}
}

func TestExecuteFailsWhenCaseResultCannotBeSaved(t *testing.T) {
	repo := &fakeRunRepo{saveErr: errors.New("case persistence unavailable")}
	svc := &evaluationRunService{
		runs:   repo,
		client: validClient(map[string]bool{"story-a-long": true, "rule": true}),
	}

	svc.execute(uuid.New(), testCases(), testItems())

	if repo.finished != entities.EvaluationRunFailed {
		t.Fatalf("status = %s, want FAILED when a case cannot be persisted", repo.finished)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("processed %d cases, want to stop after first persistence failure", len(repo.saved))
	}
	if repo.metrics != nil || repo.message == nil {
		t.Fatal("a failed run must have a public error and no completed metrics")
	}
}

func TestStartRejectsNonQA(t *testing.T) {
	svc := &evaluationRunService{users: fakeUsers{user: entities.User{Role: entities.RoleSystemAnalyst}}}

	_, err := svc.Start(context.Background(), uuid.NewString(), uuid.NewString())

	if !errors.Is(err, ErrEvaluationRunForbidden) {
		t.Fatalf("err = %v", err)
	}
}

var _ repository.EvaluationRunRepository = (*fakeRunRepo)(nil)
