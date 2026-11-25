package bdd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authService "github.com/story2gherkin/modules/auth/service"
)

const evalBDDPath = "/api/v1/evaluation-runs"

var evalBDDMissingID = uuid.MustParse(
	"00000000-0000-4000-8000-000000000099",
)

type evalFeature struct {
	datasets *goldenBDDStore
	runs     *evalBDDRunRepo
	client   *evalBDDClient
	users    goldenBDDUsers
	jwt      authService.JWTService
	router   *gin.Engine

	response      *httptest.ResponseRecorder
	authorization string
	datasetID     uuid.UUID
	runID         uuid.UUID
}

func errEvalFixture(name string, got, want int) error {
	return fmt.Errorf("%s=%d, want=%d", name, got, want)
}

// ------------------------------------
// RESET FIXTURE
// ------------------------------------

func (f *evalFeature) reset() {
	f.datasets = newGoldenBDDStore()
	f.runs = newEvalBDDRunRepo()

	f.client = &evalBDDClient{
		valid:    true,
		coverage: 100,
	}

	f.users = goldenBDDUsers{
		goldenQAID: entities.RoleQA,
		goldenSAID: entities.RoleSystemAnalyst,
	}

	f.jwt = authService.NewJWTService()

	f.router = evalBDDRouter(
		f.runs,
		f.datasets,
		f.users,
		f.client,
		f.jwt,
	)

	f.response = nil
	f.authorization = ""
	f.datasetID = goldenFixtureID
	f.runID = uuid.Nil
}

// ------------------------------------
// DATASET FIXTURES
// ------------------------------------

func (f *evalFeature) datasetWithItems(count int) error {
	if count < 1 {
		return fmt.Errorf("invalid item count %d", count)
	}

	items := make([]entities.GoldenDatasetItem, 0, count)

	for i := 1; i <= count; i++ {
		items = append(items, entities.GoldenDatasetItem{
			ItemTitle:        fmt.Sprintf("Evaluation Case %d", i),
			UserStoryText:    "As a user, I want to log in so that I can access my dashboard.",
			GUIElements:      []string{"Email field", "Password field", "Login button"},
			FreeScenarioText: "The user enters valid credentials and logs in.",
			ExpectedGherkin: "Feature: User Login\n" +
				"  Scenario: Successful Login\n" +
				"    Given the user is on the login page\n" +
				"    When the user enters valid credentials\n" +
				"    Then the dashboard is displayed",
			AcceptanceCriteria: []string{
				"valid credentials are entered",
				"dashboard is displayed",
			},
		})
	}

	return f.datasets.CreateWithItems(
		context.Background(),
		&entities.GoldenDataset{
			ID:        goldenFixtureID,
			Name:      "BDD Evaluation Dataset",
			CreatedBy: goldenQAID.String(),
		},
		items,
	)
}

func (f *evalFeature) emptyDataset() {
	f.datasets.rows[goldenFixtureID] = entities.GoldenDataset{
		ID:        goldenFixtureID,
		Name:      "BDD Empty Dataset",
		CreatedBy: goldenQAID.String(),
	}
}

// ------------------------------------
// RUN FIXTURES
// ------------------------------------

func (f *evalFeature) activeRun() {
	f.runs.seed(entities.EvaluationRunRunning, 1, 0)
}

func (f *evalFeature) runningRun() {
	f.runID = f.runs.seed(entities.EvaluationRunRunning, 3, 1)
}

// ------------------------------------
// EVALUATOR BEHAVIOR
// ------------------------------------

func (f *evalFeature) evaluatorCoverage(coverage float64) {
	f.client.valid = true
	f.client.coverage = coverage
}

func (f *evalFeature) evaluatorInvalid() {
	f.client.valid = false
	f.client.coverage = 0
}

// ------------------------------------
// USER SESSIONS
// ------------------------------------

func (f *evalFeature) login(role string) error {
	var userID uuid.UUID
	var userRole entities.Role

	switch role {
	case "QA":
		userID = goldenQAID
		userRole = entities.RoleQA
	case "SA":
		userID = goldenSAID
		userRole = entities.RoleSystemAnalyst
	default:
		return fmt.Errorf("unknown role %q", role)
	}

	token := f.jwt.GenerateAccessToken(userID.String(), string(userRole))
	f.authorization = "Bearer " + token

	return nil
}

func (f *evalFeature) noLogin() {
	f.authorization = ""
}

// ------------------------------------
// REQUESTS
// ------------------------------------

func (f *evalFeature) sendRequest(method, path string, body []byte) {
	var request *http.Request

	if body == nil {
		request = httptest.NewRequest(method, path, http.NoBody)
	} else {
		request = httptest.NewRequest(method, path, bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
	}

	if f.authorization != "" {
		request.Header.Set("Authorization", f.authorization)
	}

	f.response = httptest.NewRecorder()
	f.router.ServeHTTP(f.response, request)
}

func (f *evalFeature) startRun(datasetID uuid.UUID) error {
	body, err := json.Marshal(map[string]string{
		"golden_dataset_id": datasetID.String(),
	})
	if err != nil {
		return err
	}

	f.sendRequest(http.MethodPost, evalBDDPath, body)

	var envelope struct {
		Status bool `json:"status"`
		Data   struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(f.response.Body.Bytes(), &envelope); err != nil {
		return err
	}

	if envelope.Data.RunID != "" {
		f.runID, err = uuid.Parse(envelope.Data.RunID)
	}

	return err
}

func (f *evalFeature) startRunOnDataset() error {
	return f.startRun(f.datasetID)
}

func (f *evalFeature) startRunOnMissingDataset() error {
	return f.startRun(evalBDDMissingID)
}

func (f *evalFeature) requestProgress(runID uuid.UUID) {
	f.sendRequest(
		http.MethodGet,
		evalBDDPath+"/"+runID.String()+"/progress",
		nil,
	)
}

func (f *evalFeature) requestStoredProgress() {
	f.requestProgress(f.runID)
}

func (f *evalFeature) requestMissingProgress() {
	f.requestProgress(evalBDDMissingID)
}

// ------------------------------------
// ASSERTIONS
// ------------------------------------

func (f *evalFeature) status(want int) error {
	if f.response == nil {
		return fmt.Errorf("HTTP request has not been sent")
	}

	if f.response.Code != want {
		return fmt.Errorf(
			"HTTP status=%d, want=%d; response=%s",
			f.response.Code,
			want,
			f.response.Body.String(),
		)
	}

	return nil
}

func (f *evalFeature) responseTotalCases(want int) error {
	var envelope struct {
		Data struct {
			TotalCases int `json:"total_cases"`
		} `json:"data"`
	}

	if err := json.Unmarshal(f.response.Body.Bytes(), &envelope); err != nil {
		return err
	}

	if envelope.Data.TotalCases != want {
		return fmt.Errorf(
			"total_cases=%d, want=%d",
			envelope.Data.TotalCases,
			want,
		)
	}

	return nil
}

// waitFinished waits for the background worker to end the run.
func (f *evalFeature) waitFinished() (entities.EvaluationRun, error) {
	deadline := time.Now().Add(2 * time.Second)

	for {
		run, exists := f.runs.snapshot(f.runID)

		if exists &&
			(run.Status == entities.EvaluationRunCompleted ||
				run.Status == entities.EvaluationRunFailed) {
			return run, nil
		}

		if time.Now().After(deadline) {
			if !exists {
				return entities.EvaluationRun{}, fmt.Errorf("run is not stored")
			}

			return entities.EvaluationRun{}, fmt.Errorf(
				"run status=%q, want COMPLETED or FAILED",
				run.Status,
			)
		}

		time.Sleep(5 * time.Millisecond)
	}
}

func (f *evalFeature) finished(status string, processed int) error {
	run, err := f.waitFinished()
	if err != nil {
		return err
	}

	if string(run.Status) != status {
		return fmt.Errorf("run status=%q, want=%q", run.Status, status)
	}

	progress, err := f.runs.GetProgress(context.Background(), run.ID)
	if err != nil {
		return err
	}

	if progress.ProcessedCases != processed {
		return fmt.Errorf(
			"processed_cases=%d, want=%d",
			progress.ProcessedCases,
			processed,
		)
	}

	return nil
}

func (f *evalFeature) metricsAvailable() error {
	run, exists := f.runs.snapshot(f.runID)

	if !exists || run.Metrics == nil {
		return fmt.Errorf("run metrics are not stored")
	}

	if run.Metrics.LLM.ValidationPercent == nil ||
		run.Metrics.RuleBased.ValidationPercent == nil ||
		run.Metrics.LLM.AcceptanceCriteriaPrecisionPercent == nil ||
		run.Metrics.LLM.AcceptanceCriteriaRecallPercent == nil ||
		run.Metrics.LLM.AcceptanceCriteriaF1Percent == nil ||
		run.Metrics.RuleBased.AcceptanceCriteriaPrecisionPercent == nil ||
		run.Metrics.RuleBased.AcceptanceCriteriaRecallPercent == nil ||
		run.Metrics.RuleBased.AcceptanceCriteriaF1Percent == nil {
		return fmt.Errorf("validation_percent is missing for one method")
	}

	return nil
}

func (f *evalFeature) acceptanceF1(want float64) error {
	run, exists := f.runs.snapshot(f.runID)
	if !exists || run.Metrics == nil ||
		run.Metrics.LLM.AcceptanceCriteriaF1Percent == nil ||
		run.Metrics.RuleBased.AcceptanceCriteriaF1Percent == nil {
		return fmt.Errorf("acceptance criteria F1 is unavailable")
	}

	if *run.Metrics.LLM.AcceptanceCriteriaF1Percent != want ||
		*run.Metrics.RuleBased.AcceptanceCriteriaF1Percent != want {
		return fmt.Errorf(
			"acceptance criteria F1=%v/%v, want=%v",
			*run.Metrics.LLM.AcceptanceCriteriaF1Percent,
			*run.Metrics.RuleBased.AcceptanceCriteriaF1Percent,
			want,
		)
	}

	return nil
}

func (f *evalFeature) caseStatus(want, method string) error {
	result, err := f.runs.caseResult(f.runID, method)
	if err != nil {
		return err
	}

	if result == nil {
		return fmt.Errorf("case result for %q is not stored", method)
	}

	if string(result.Status) != want {
		return fmt.Errorf(
			"case status=%q, want=%q",
			result.Status,
			want,
		)
	}

	return nil
}

func (f *evalFeature) noRuns() error {
	if got := f.runs.runCount(); got != 0 {
		return errEvalFixture("stored runs", got, 0)
	}

	return nil
}

func (f *evalFeature) noNewRuns() error {
	if got := f.runs.createdCount(); got != 0 {
		return errEvalFixture("created runs", got, 0)
	}

	return nil
}

func (f *evalFeature) progressCounts(processed, total int) error {
	var envelope struct {
		Data struct {
			TotalCases     int `json:"total_cases"`
			ProcessedCases int `json:"processed_cases"`
		} `json:"data"`
	}

	if err := json.Unmarshal(f.response.Body.Bytes(), &envelope); err != nil {
		return err
	}

	if envelope.Data.ProcessedCases != processed ||
		envelope.Data.TotalCases != total {
		return fmt.Errorf(
			"progress=%d/%d, want=%d/%d",
			envelope.Data.ProcessedCases,
			envelope.Data.TotalCases,
			processed,
			total,
		)
	}

	return nil
}

// ------------------------------------
// SCENARIO REGISTRATION
// ------------------------------------

func InitializeEvaluationScenario(ctx *godog.ScenarioContext) {
	f := &evalFeature{}

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		f.reset()
		return ctx, nil
	})

	ctx.Step(`^pengguna evaluasi login sebagai "([^"]+)"$`, f.login)
	ctx.Step(`^pengguna evaluasi tidak login$`, f.noLogin)

	ctx.Step(`^terdapat Golden Dataset evaluasi berisi (\d+) item$`, f.datasetWithItems)
	ctx.Step(`^terdapat Golden Dataset evaluasi tanpa item$`, f.emptyDataset)

	ctx.Step(`^terdapat evaluasi aktif yang tersimpan$`, f.activeRun)
	ctx.Step(
		`^terdapat evaluasi yang sedang berjalan dengan (\d+) kasus dan (\d+) kasus diproses$`,
		func(total, processed int) error {
			if total < 1 || processed < 0 || processed > total {
				return fmt.Errorf("invalid progress fixture %d/%d", processed, total)
			}

			f.runningRun()
			return nil
		},
	)

	ctx.Step(
		`^evaluator menghasilkan Gherkin valid dengan cakupan (\d+) persen$`,
		func(coverage float64) { f.evaluatorCoverage(coverage) },
	)
	ctx.Step(`^evaluator menghasilkan Gherkin tidak valid$`, f.evaluatorInvalid)

	ctx.Step(`^pengguna menjalankan evaluasi$`, f.startRunOnDataset)
	ctx.Step(
		`^pengguna menjalankan evaluasi pada dataset yang tidak tersedia$`,
		f.startRunOnMissingDataset,
	)
	ctx.Step(`^pengguna meminta progress evaluasi$`, f.requestStoredProgress)
	ctx.Step(
		`^pengguna meminta progress evaluasi yang tidak tersedia$`,
		f.requestMissingProgress,
	)

	ctx.Step(`^status respons evaluasi harus (\d+)$`, f.status)
	ctx.Step(`^respons evaluasi berisi (\d+) kasus$`, f.responseTotalCases)
	ctx.Step(
		`^evaluasi berstatus "([^"]+)" dengan (\d+) kasus diproses$`,
		f.finished,
	)
	ctx.Step(`^metrik evaluasi menyimpan dua metode$`, f.metricsAvailable)
	ctx.Step(`^metrik acceptance criteria F1 bernilai (\d+) persen$`, f.acceptanceF1)
	ctx.Step(
		`^kasus evaluasi berstatus "([^"]+)" untuk metode "([^"]+)"$`,
		f.caseStatus,
	)
	ctx.Step(`^tidak ada evaluasi yang tersimpan$`, f.noRuns)
	ctx.Step(`^tidak ada evaluasi baru yang tersimpan$`, f.noNewRuns)
	ctx.Step(
		`^progress evaluasi menunjukkan (\d+) dari (\d+) kasus$`,
		f.progressCounts,
	)
}

// ------------------------------------
// STRICT BDD TEST SUITE
// ------------------------------------

func TestRunEvaluationFeatures(t *testing.T) {
	t.Setenv(
		"JWT_SECRET",
		"run-evaluation-bdd-only-signing-key-not-production",
	)

	gin.SetMode(gin.TestMode)

	suite := godog.TestSuite{
		Name:                "run_evaluation",
		ScenarioInitializer: InitializeEvaluationScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/run_evaluation.feature"},
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Run Evaluation BDD suite failed")
	}
}
