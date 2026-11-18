# Run Evaluation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** QA dapat memilih Golden Dataset, menjalankan evaluasi LLM vs Rule-Based per kasus, memantau progres, dan hasilnya otomatis tampil di Evaluation Report.

**Architecture:** ai-service menyediakan dua generator (LLM dan rule-based deterministik) di belakang kontrak `GenerateRequest`/`GenerateResponse` yang sama. Backend Go menjadi orkestrator: membuat run + snapshot kasus, menjalankan loop di goroutine, menyimpan hasil per kasus, lalu menghitung metrik. Frontend hanya memanggil `POST /evaluation-runs`, polling `GET /evaluation-runs/:run_id/progress`, dan menautkan ke report.

**Tech Stack:** Go (Gin, GORM/Postgres, samber/do), Python (FastAPI, gherkin), React/TypeScript (Vite).

**Spec:** `docs/superpowers/specs/2026-11-19-run-evaluation-design.md`

## Global Constraints

- QA-only: setiap endpoint baru menolak non-QA dengan `403` (pola `authorizeQA` di `golden_dataset_service.go`).
- Jangan mengubah kontrak Report yang sudah ada (`docs/EVALUATION_REPORT_API.md`).
- Status run tetap `QUEUED|RUNNING|COMPLETED|FAILED`; status kasus tetap `PASS|WARN|FAIL`.
- Tidak menyimpan teks Gherkin hasil generate; tidak ada similarity/winner/Overall Score.
- Tidak ada commit/push tanpa permintaan eksplisit dari user.
- Test wajib ada untuk logika baru: pytest (ai-service), `go test` (service/handler), `npm run build` (frontend).

## Amandemen spec

Spec menyebut "kolom progres baru" tidak dibangun. Progres tidak bisa dihitung andal dari hasil kasus karena satu kasus bisa gagal di kedua metode (hasil `null`). Karena itu task 3 menambah satu kolom `processed boolean` pada `evaluation_run_cases`; AutoMigrate menanganinya. Perbarui spec saat implementasi task 3.

---

### Task 1: Rule-Based Generator di ai-service

**Files:**
- Create: `ai-service/app/services/rule_based.py`
- Create: `ai-service/tests/unit/test_rule_based.py`
- Modify: `ai-service/app/api/routes/generate.py` (tambah route)

**Interfaces:**
- Consumes: `GenerateRequest`, `GenerateResponse`, `GeneratedFeature`, `Scenario` dari `app.schemas.generate`; `render_gherkin` dari `app.services.generator`.
- Produces: `generate_rule_based(request: GenerateRequest) -> GenerateResponse`, route `POST /internal/scenarios/generate-rule-based`.

- [ ] **Step 1: Tulis test yang gagal**

`ai-service/tests/unit/test_rule_based.py`:

```python
from app.schemas.generate import GenerateRequest
from app.services.rule_based import generate_rule_based
from app.services.validation import validate_gherkin


def make_request(elements: list[str]) -> GenerateRequest:
    return GenerateRequest(
        user_story="Sebagai admin, saya ingin mengelola pengguna agar data rapi.",
        gui_elements=elements,
        free_scenario="Pengguna dapat melihat daftar pengguna.",
    )


def test_covers_every_element_and_passes_validation():
    elements = [f"Tombol Aksi {i}" for i in range(1, 8)]
    response = generate_rule_based(make_request(elements))
    validation = validate_gherkin(response.gherkin, elements)

    assert response.model == "rule-based"
    assert response.prompt_version == "rule-based-v1"
    assert validation.valid is True
    assert validation.not_covered_elements == []
    assert validation.coverage_percent == 100.0


def test_is_deterministic():
    request = make_request(["Kolom Nama", "Tombol Simpan"])
    assert generate_rule_based(request).gherkin == generate_rule_based(request).gherkin


def test_handles_maximum_element_length_and_count():
    elements = [("E" * 199) + str(i) for i in range(30)]
    response = generate_rule_based(make_request(elements))
    validation = validate_gherkin(response.gherkin, elements)
    assert validation.valid is True
    assert validation.coverage_percent == 100.0
    assert all(len(line) <= 500 for line in response.gherkin.splitlines())
```

- [ ] **Step 2: Jalankan test, pastikan gagal**

Run: `cd ai-service && .venv/bin/pytest tests/unit/test_rule_based.py -q`
Expected: FAIL `ModuleNotFoundError: No module named 'app.services.rule_based'`

- [ ] **Step 3: Implementasi**

`ai-service/app/services/rule_based.py`:

```python
import time

from app.schemas.generate import GenerateRequest, GenerateResponse, GeneratedFeature, Scenario
from app.services.generator import render_gherkin

PROMPT_VERSION = "rule-based-v1"
ELEMENTS_PER_SCENARIO = 6
ELEMENTS_PER_STEP = 2


def _feature_name(user_story: str) -> str:
    first_line = user_story.strip().splitlines()[0].strip()
    if first_line.endswith("."):
        first_line = first_line[:-1]
    return first_line[:200] or "User story"


def _chunks(items: list[str], size: int) -> list[list[str]]:
    return [items[i:i + size] for i in range(0, len(items), size)]


def _steps_for(chunk: list[str]) -> list[str]:
    return [
        "the user uses " + " and ".join(f'"{element}"' for element in group)
        for group in _chunks(chunk, ELEMENTS_PER_STEP)
    ]


def generate_rule_based(request: GenerateRequest) -> GenerateResponse:
    started = time.monotonic()
    scenarios = []
    offset = 0
    for index, chunk in enumerate(_chunks(request.gui_elements, ELEMENTS_PER_SCENARIO), start=1):
        scenarios.append(Scenario(
            name=f"Rule-based check {index}",
            kind="positive",
            gui_element_ids=list(range(offset + 1, offset + len(chunk) + 1)),
            given=["the user is on the page described by the user story"],
            when=_steps_for(chunk),
            then=[f"the system shows the expected result: {request.free_scenario[:200]}"],
        ))
        offset += len(chunk)
    feature = GeneratedFeature(feature_name=_feature_name(request.user_story), scenarios=scenarios)
    return GenerateResponse(
        feature=feature,
        gherkin=render_gherkin(feature),
        model="rule-based",
        prompt_version=PROMPT_VERSION,
        duration_ms=int((time.monotonic() - started) * 1000),
    )
```

Catatan: `gui_element_ids` hanya metadata skema; `render_gherkin` tidak memakainya. `ELEMENTS_PER_STEP=2` menjaga baris ≤ ~440 karakter (elemen maksimum 200 karakter).

Tambah route di `ai-service/app/api/routes/generate.py`:

```python
from app.services.rule_based import generate_rule_based


@router.post("/internal/scenarios/generate-rule-based", response_model=GenerateResponse)
async def generate_rule_based_route(request_body: GenerateRequest):
    return generate_rule_based(request_body)
```

- [ ] **Step 4: Jalankan test, pastikan lulus**

Run: `cd ai-service && .venv/bin/pytest tests/unit/test_rule_based.py -q`
Expected: 3 passed

- [ ] **Step 5: Pastikan route terpasang dan tetap butuh service key**

Run: `cd ai-service && .venv/bin/pytest -q`
Expected: seluruh suite lulus (route otomatis memakai dependency `require_service_key` karena didaftarkan lewat router yang sama di `app/main.py`).

---

### Task 2: Client Go memanggil generator rule-based

**Files:**
- Modify: `backend/modules/scenario/service/llm_client.go`
- Test: `backend/modules/scenario/service/llm_client_rule_based_test.go`

**Interfaces:**
- Produces: `func (c *LLMClient) GenerateRuleBased(ctx context.Context, input dto.GenerateRequest) (dto.GenerateResponse, error)` — URL = `strings.TrimSuffix(c.endpoint, "/generate") + "/generate-rule-based"`, header `X-AI-Service-Key` sama seperti `Generate`.

- [ ] **Step 1: Tulis test yang gagal**

`backend/modules/scenario/service/llm_client_rule_based_test.go`:

```go
package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/story2gherkin/modules/scenario/dto"
)

func TestLLMClientGenerateRuleBased(t *testing.T) {
	var gotPath, gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotKey = r.URL.Path, r.Header.Get("X-AI-Service-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"feature":{"feature_name":"F","scenarios":[],"warnings":[]},"gherkin":"Feature: F\n","model":"rule-based","prompt_version":"rule-based-v1","duration_ms":1}`))
	}))
	defer server.Close()

	client, err := NewLLMClient(server.URL+"/internal/scenarios/generate", "secret-key-at-least-32-characters-long")
	if err != nil {
		t.Fatalf("NewLLMClient: %v", err)
	}

	res, err := client.GenerateRuleBased(context.Background(), dto.GenerateRequest{
		UserStory:    "story long enough",
		GuiElements:  []string{"Tombol"},
		FreeScenario: "scenario long enough",
	})
	if err != nil {
		t.Fatalf("GenerateRuleBased: %v", err)
	}
	if gotPath != "/internal/scenarios/generate-rule-based" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotKey == "" || res.Gherkin == "" {
		t.Fatalf("missing key or gherkin")
	}
}
```

- [ ] **Step 2: Jalankan test, pastikan gagal**

Run: `cd backend && go test ./modules/scenario/service -run TestLLMClientGenerateRuleBased -v`
Expected: FAIL `client.GenerateRuleBased undefined`

- [ ] **Step 3: Implementasi**

Ekstrak badan `Generate` menjadi helper `postGenerate(ctx, url, input)` lalu tambahkan:

```go
// GenerateRuleBased requests the deterministic generator.
func (c *LLMClient) GenerateRuleBased(
	ctx context.Context,
	input dto.GenerateRequest,
) (dto.GenerateResponse, error) {
	return c.postGenerate(ctx, c.ruleBasedEndpoint, input)
}
```

di mana `ruleBasedEndpoint` diisi di `NewLLMClient`:

```go
ruleBased: strings.TrimSuffix(endpoint, "/generate") + "/generate-rule-based",
```

- [ ] **Step 4: Jalankan test**

Run: `cd backend && go test ./modules/scenario/service -run TestLLMClient -v`
Expected: PASS (semua test client).

---

### Task 3: Entity + repository run evaluasi

**Files:**
- Modify: `backend/database/entities/evaluation_run_case_entity.go` (tambah `Processed bool`)
- Create: `backend/modules/scenario/repository/evaluation_run_repository.go`
- Modify: `docs/superpowers/specs/2026-11-19-run-evaluation-design.md` (amandemen kolom `processed`)

**Interfaces:**
- Produces:

```go
type EvaluationRunRepository interface {
	HasActive(ctx context.Context) (bool, error)
	CreateWithCases(ctx context.Context, run *entities.EvaluationRun, cases []entities.EvaluationRunCase) error
	MarkRunning(ctx context.Context, runID uuid.UUID) error
	SaveCaseResult(ctx context.Context, caseID uuid.UUID, llm, ruleBased *entities.EvaluationCaseMethodResult) error
	GetProgress(ctx context.Context, runID uuid.UUID) (dto.EvaluationRunProgress, error)
	Finish(ctx context.Context, runID uuid.UUID, status entities.EvaluationRunStatus, metrics *entities.EvaluationRunMetrics, message *string) error
}
```

- `NewEvaluationRunRepository(db *gorm.DB) EvaluationRunRepository`.

- [ ] **Step 1: Tambah field entity**

```go
	// Processed flips once both generators finished, even when both failed.
	Processed bool `gorm:"not null;default:false"`
```

- [ ] **Step 2: Implementasi repository**

`HasActive`: `db.Model(&entities.EvaluationRun{}).Where("status = ?", entities.EvaluationRunRunning).Count(&n)` → `n > 0`.

`CreateWithCases`: satu transaksi —

```go
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
	if err := tx.Create(run).Error; err != nil {
		return err
	}
	for i := range cases {
		cases[i].RunID = run.ID
	}
	if err := tx.Create(&cases).Error; err != nil {
		return err
	}
	return tx.Model(&entities.GoldenDataset{}).
		Where("id = ? AND used_at IS NULL", run.GoldenDatasetID).
		Update("used_at", time.Now()).Error
})
```

`SaveCaseResult`:

```go
return r.db.WithContext(ctx).Model(&entities.EvaluationRunCase{}).
	Where("id = ?", caseID).
	Updates(map[string]any{"llm": llm, "rule_based": ruleBased, "processed": true}).Error
```

`GetProgress`:

```go
var row struct {
	Status      entities.EvaluationRunStatus
	TotalCases  int
	Processed   int
	ErrorMessage *string
}
err := r.db.WithContext(ctx).Model(&entities.EvaluationRun{}).
	Select("evaluation_runs.status, evaluation_runs.total_cases, evaluation_runs.error_message, (SELECT count(*) FROM evaluation_run_cases WHERE run_id = evaluation_runs.id AND processed) AS processed").
	Where("evaluation_runs.id = ?", runID).
	Scan(&row).Error
```

`ErrEvaluationRunNotFound` dikembalikan bila `RowsAffected == 0` (gunakan `Take`/`First` agar `gorm.ErrRecordNotFound` bisa dipetakan).

`Finish`: update `status`, `metrics`, `error_message`.

- [ ] **Step 3: Kompilasi**

Run: `cd backend && go build ./...`
Expected: sukses.

---

### Task 4: Service orkestrasi + metrik

**Files:**
- Create: `backend/modules/scenario/service/evaluation_run_service.go`
- Create: `backend/modules/scenario/service/evaluation_run_service_test.go`

**Interfaces:**
- Consumes: `EvaluationRunRepository` (Task 3), `LLMClient.Generate/Validate/GenerateRuleBased`, `GoldenDatasetRepository.GetByID`, `ReviewUserRepository.GetUserById`.
- Produces:

```go
type EvaluationRunService interface {
	Start(ctx context.Context, userID, datasetID string) (dto.EvaluationRunCreated, error)
	Progress(ctx context.Context, userID, runID string) (dto.EvaluationRunProgress, error)
}

func NewEvaluationRunService(
	runs repository.EvaluationRunRepository,
	datasets repository.GoldenDatasetRepository,
	users ReviewUserRepository,
	client EvaluationClient,
) EvaluationRunService

type EvaluationClient interface {
	Generate(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error)
	GenerateRuleBased(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error)
	Validate(context.Context, string, []string) (dto.ScenarioValidation, error)
}
```

- [ ] **Step 1: Tulis test yang gagal**

`backend/modules/scenario/service/evaluation_run_service_test.go` (package `service`, fake repo + fake client):

```go
func TestRunCasesComputesStatusesAndMetrics(t *testing.T) {
	cases := []entities.EvaluationRunCase{{ID: uuid.New(), Position: 1}, {ID: uuid.New(), Position: 2}}
	repo := &fakeRunRepo{cases: cases}
	client := &fakeEvalClient{
		generate: map[string]dto.GenerateResponse{"case-1": {Gherkin: "valid"}, "case-2": {Gherkin: "invalid"}},
		rule:     dto.GenerateResponse{Gherkin: "valid"},
		validate: func(text string, _ []string) dto.ScenarioValidation {
			if text == "valid" {
				return dto.ScenarioValidation{Valid: true, CoveragePercent: 100}
			}
			return dto.ScenarioValidation{Valid: false, CoveragePercent: 50}
		},
	}
	svc := &evaluationRunService{runs: repo, client: client}

	svc.runCases(context.Background(), uuid.New(), []entities.GoldenDatasetItem{
		{ItemTitle: "A", UserStoryText: "story long enough", GUIElements: []string{"X"}, FreeScenarioText: "scenario long enough"},
		{ItemTitle: "B", UserStoryText: "story long enough", GUIElements: []string{"Y"}, FreeScenarioText: "scenario long enough"},
	})

	if repo.finishedStatus != entities.EvaluationRunCompleted {
		t.Fatalf("status = %s", repo.finishedStatus)
	}
	if got := *repo.metrics.LLM.ValidationPercent; got != 50 {
		t.Fatalf("llm validity = %v", got)
	}
	if got := *repo.metrics.RuleBased.ValidationPercent; got != 100 {
		t.Fatalf("rule validity = %v", got)
	}
	if repo.saved[0].LLM.Status != entities.EvaluationCasePass || repo.saved[0].LLM.GUICoveragePercent == nil {
		t.Fatalf("case 1 result wrong: %+v", repo.saved[0].LLM)
	}
	if repo.saved[1].LLM.Status != entities.EvaluationCaseFail {
		t.Fatalf("case 2 should be FAIL")
	}
}

func TestCaseFailureDoesNotStopRun(t *testing.T) {
	repo := &fakeRunRepo{cases: []entities.EvaluationRunCase{{ID: uuid.New()}, {ID: uuid.New()}}}
	client := &fakeEvalClient{
		generateErr:  map[string]error{"case-1": errors.New("llm down")},
		generate:     map[string]dto.GenerateResponse{"case-2": {Gherkin: "valid"}},
		rule:         dto.GenerateResponse{Gherkin: "valid"},
		validate:     func(string, []string) dto.ScenarioValidation { return dto.ScenarioValidation{Valid: true, CoveragePercent: 100} },
	}
	svc := &evaluationRunService{runs: repo, client: client}

	svc.runCases(context.Background(), uuid.New(), twoItems())

	if len(repo.saved) != 2 || repo.saved[0].LLM != nil {
		t.Fatalf("failed case must be saved with nil LLM: %+v", repo.saved)
	}
	if repo.finishedStatus != entities.EvaluationRunCompleted {
		t.Fatalf("status = %s", repo.finishedStatus)
	}
}

func TestStartRejectsWhenActiveRunExists(t *testing.T) {
	svc := &evaluationRunService{runs: &fakeRunRepo{active: true}}
	_, err := svc.Start(context.Background(), validQAID, uuid.NewString())
	if !errors.Is(err, ErrEvaluationRunActive) {
		t.Fatalf("err = %v", err)
	}
}
```

- [ ] **Step 2: Jalankan test, pastikan gagal**

Run: `cd backend && go test ./modules/scenario/service -run TestRunCases -v`
Expected: FAIL `undefined: newEvaluationRunner`

- [ ] **Step 3: Implementasi**

Aturan inti (fungsi murni, diuji langsung):

```go
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

func summarize(results []*entities.EvaluationCaseMethodResult, total int) entities.EvaluationMethodMetrics {
	if total == 0 {
		return entities.EvaluationMethodMetrics{}
	}
	valid, coverageSum, coverageCount := 0, 0.0, 0
	for _, r := range results {
		if r == nil {
			continue
		}
		if r.Status != entities.EvaluationCaseFail {
			valid++
		}
		if r.GUICoveragePercent != nil {
			coverageSum += *r.GUICoveragePercent
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
```

`runCases(ctx, runID, items)`:

```go
for index, item := range items {
	request := dto.GenerateRequest{UserStory: item.UserStoryText, GuiElements: item.GUIElements, FreeScenario: item.FreeScenarioText}
	var llm, ruleBased *entities.EvaluationCaseMethodResult
	if res, err := s.client.Generate(ctx, request); err == nil {
		if v, err := s.client.Validate(ctx, res.Gherkin, item.GUIElements); err == nil {
			llm = caseResult(v)
		}
	}
	if res, err := s.client.GenerateRuleBased(ctx, request); err == nil {
		if v, err := s.client.Validate(ctx, res.Gherkin, item.GUIElements); err == nil {
			ruleBased = caseResult(v)
		}
	}
	results = append(results, llm, ruleBased)
	_ = s.runs.SaveCaseResult(ctx, caseIDs[index], llm, ruleBased)
}
```

`Start` urutan: authorize QA → parse `datasetID` → `datasets.GetByID` (404 bila kosong item → `ErrEvaluationRunEmptyDataset`) → `runs.HasActive` (409) → bangun run (`DatasetName`, `CreatedBy`, `TotalCases`, `Status: QUEUED`) + kasus (pre-generate `uuid.New()` untuk `run.ID` agar `RunID` bisa diisi) → `CreateWithCases` → `go s.execute(run.ID, items)` → return `{run_id, report_code: fmt.Sprintf("ER-%04d", run.RunNumber), status, total_cases}`.

`execute` memakai `context.Background()`, `MarkRunning`, `runCases`, lalu `Finish(COMPLETED, &metrics, nil)`; bila terjadi error DB fatal → `Finish(FAILED, nil, &msg)` dengan pesan publik `"Evaluation failed due to an internal error."`.

- [ ] **Step 4: Jalankan test**

Run: `cd backend && go test ./modules/scenario/service -v`
Expected: seluruh test service PASS.

---

### Task 5: DTO, handler, route, DI

**Files:**
- Create: `backend/modules/scenario/dto/evaluation_run_dto.go`
- Create: `backend/modules/scenario/handler/evaluation_run_handler.go`
- Create: `backend/modules/scenario/handler/evaluation_run_handler_test.go`
- Create: `backend/modules/scenario/evaluation_run_routes.go`
- Modify: `backend/modules/scenario/routes.go` (panggil `RegisterEvaluationRunRoutes`)
- Modify: `backend/providers/core.go` (wiring)

**Interfaces:**
- `POST /api/v1/evaluation-runs` body `{"golden_dataset_id":"<uuid>"}` → `201 {status:true,message,data:{run_id,report_code,status,total_cases}}`.
- `GET /api/v1/evaluation-runs/:run_id/progress` → `200 {data:{run_id,status,total_cases,processed_cases,error_message}}`.
- Pemetaan error: `ErrEvaluationRunForbidden` → 403, `ErrEvaluationRunActive` → 409, `ErrEvaluationRunNotFound`/`ErrEvaluationRunEmptyDataset` → 404/400, else 500.

- [ ] **Step 1: DTO**

```go
type EvaluationRunCreateRequest struct {
	GoldenDatasetID string `json:"golden_dataset_id" binding:"required,uuid"`
}

type EvaluationRunCreated struct {
	RunID      string                      `json:"run_id"`
	ReportCode string                      `json:"report_code"`
	Status     entities.EvaluationRunStatus `json:"status"`
	TotalCases int                         `json:"total_cases"`
}

type EvaluationRunProgress struct {
	RunID          string                      `json:"run_id"`
	Status         entities.EvaluationRunStatus `json:"status"`
	TotalCases     int                         `json:"total_cases"`
	ProcessedCases int                         `json:"processed_cases"`
	ErrorMessage   *string                     `json:"error_message"`
}
```

- [ ] **Step 2: Test handler (fake service)**

```go
func TestCreateMapsActiveRunToConflict(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("user_id", uuid.NewString())
	c.Request = httptest.NewRequest(http.MethodPost, "/evaluation-runs",
		strings.NewReader(`{"golden_dataset_id":"`+uuid.NewString()+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := NewEvaluationRunHandler(fakeRunService{startErr: service.ErrEvaluationRunActive})

	h.Create(c)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

func TestCreateReturnsCreatedRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("user_id", uuid.NewString())
	c.Request = httptest.NewRequest(http.MethodPost, "/evaluation-runs",
		strings.NewReader(`{"golden_dataset_id":"`+uuid.NewString()+`"}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h := NewEvaluationRunHandler(fakeRunService{created: dto.EvaluationRunCreated{
		RunID: "abc", ReportCode: "ER-0001", Status: entities.EvaluationRunQueued, TotalCases: 3,
	}})

	h.Create(c)

	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"run_id":"abc"`) {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

type fakeRunService struct {
	created  dto.EvaluationRunCreated
	startErr error
}

func (f fakeRunService) Start(context.Context, string, string) (dto.EvaluationRunCreated, error) {
	return f.created, f.startErr
}

func (f fakeRunService) Progress(context.Context, string, string) (dto.EvaluationRunProgress, error) {
	return dto.EvaluationRunProgress{}, nil
}
```

- [ ] **Step 3: Implementasi handler**

Ikuti pola `golden_dataset_handler.go`: `c.Header("Cache-Control","no-store")`, `userID := c.GetString("user_id")` (401 bila kosong), `ShouldBindJSON`, panggil service, bungkus dengan `utils.BuildResponseSuccess/Failed`.

- [ ] **Step 4: Route + DI**

`evaluation_run_routes.go`:

```go
func RegisterEvaluationRunRoutes(server *gin.RouterGroup, jwtService authService.JWTService, runs *handler.EvaluationRunHandler) {
	routes := server.Group("/evaluation-runs", middlewares.Authenticate(jwtService))
	routes.POST("", runs.Create)
	routes.GET("/:run_id/progress", runs.Progress)
}
```

`routes.go`: tambah `runs := do.MustInvoke[*handler.EvaluationRunHandler](injector)` + `RegisterEvaluationRunRoutes(server, jwt, runs)`.

`providers/core.go`:

```go
do.Provide(injector, func(i *do.Injector) (*scenHandler.EvaluationRunHandler, error) {
	baseURL := os.Getenv("AI_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8001"
	}
	client, err := scenService.NewLLMClient(baseURL, os.Getenv("AI_SERVICE_KEY"))
	if err != nil {
		return nil, err
	}
	svc := scenService.NewEvaluationRunService(
		scenRepo.NewEvaluationRunRepository(db),
		scenRepo.NewGoldenDatasetRepository(db),
		userRepository,
		client,
	)
	return scenHandler.NewEvaluationRunHandler(svc), nil
})
```

- [ ] **Step 5: Build + test**

Run: `cd backend && go build ./... && go test ./modules/scenario/... ./test/...`
Expected: PASS.

---

### Task 6: Halaman Run Evaluation

**Files:**
- Create: `frontend/src/services/evaluationRunApi.ts`
- Modify: `frontend/src/pages/EvaluationPage.tsx`
- Modify: `frontend/src/pages/GenerateScenarioPage.css` hanya bila perlu kelas baru (utamakan kelas yang sudah ada: `generate-page`, `generate-layout`, `generate-main`, `documentation-card`)

**Interfaces:**
- Consumes: `GET /golden-datasets` (existing `goldenDatasetApi.list`), `POST /evaluation-runs`, `GET /evaluation-runs/:id/progress`.
- Produces: `createEvaluationRun(datasetId): Promise<EvaluationRunCreated>`, `getEvaluationRunProgress(runId): Promise<EvaluationRunProgress>`.

- [ ] **Step 1: API client**

```ts
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8888/api/v1'

export interface EvaluationRunCreated {
  run_id: string
  report_code: string
  status: 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED'
  total_cases: number
}

export interface EvaluationRunProgress {
  run_id: string
  status: 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED'
  total_cases: number
  processed_cases: number
  error_message: string | null
}
```

Kedua fungsi mengirim `Authorization: Bearer ${token}` dari `authSession` dengan pola yang sama seperti `evaluationReportApi.ts`, dan melempar `Error` berisi pesan dari envelope `{status,message,error}`.

- [ ] **Step 2: Halaman**

Struktur mengikuti mockup:
- dropdown dataset dari `goldenDatasetApi.list('', 1, 100)`; label `${name} — ${item_count} Cases`;
- kartu Informasi Dataset (Jumlah Kasus, Tanggal Dibuat, Deskripsi) dari item terpilih;
- kartu Evaluation Metrics (dua label: Validation Metric, GUI Coverage) + daftar proses;
- dua kartu metode + `VS`;
- tombol `Jalankan Evaluasi` (disabled bila tidak ada dataset, dataset tanpa item, atau sedang berjalan).

Perilaku: klik → `createEvaluationRun` → poll `getEvaluationRunProgress` tiap 1500 ms → tampilkan `processed_cases/total_cases`; saat `COMPLETED` tampilkan "Evaluation completed successfully" + `Link` ke `/evaluation/reports/${run_id}`; saat `FAILED` tampilkan `error_message`. Hentikan polling saat komponen unmount (`clearInterval`).

- [ ] **Step 3: Build**

Run: `cd frontend && npm run build`
Expected: sukses tanpa error TypeScript.

---

### Task 7: Dokumentasi + smoke test manual

**Files:**
- Modify: `docs/EVALUATION_REPORT_API.md` (bagian baru "Run Evaluation")

- [ ] **Step 1: Dokumentasikan endpoint dan aturan**

Tambahkan: body `POST`, respons `201`, `409` saat ada run `RUNNING`, `progress` endpoint, serta aturan `PASS/WARN/FAIL` dan rumus metrik persis seperti spec.

- [ ] **Step 2: Smoke test end-to-end**

```bash
docker compose up -d postgres backend
cd ai-service && .venv/bin/uvicorn app.main:app --host 0.0.0.0 --port 8001 &
# backend container butuh AI_SERVICE_URL=http://172.23.0.1:8001
curl -s -X POST localhost:8008/api/v1/evaluation-runs -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d '{"golden_dataset_id":"'"$DATASET_ID"'"}'
curl -s localhost:8008/api/v1/evaluation-runs/$RUN_ID/progress -H "Authorization: Bearer $TOKEN"
fetch http://localhost:5173/evaluation  # login QA: qa.lokal@test.local
```

Expected: `201` + run dibuat, progress naik ke `total_cases`, status `COMPLETED`, report baru muncul di `/evaluation/reports`, dataset terpilih ber-`used_at`.

- [ ] **Step 3: Matikan ai-service**

Run: `fuser -k 8001/tcp`
Expected: proses uvicorn berhenti.
