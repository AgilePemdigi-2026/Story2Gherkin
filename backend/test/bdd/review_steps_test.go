package bdd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
)

type reviewFeature struct {
	store           *reviewBDDStore
	validator       *reviewBDDValidator
	users           reviewBDDUsers
	jwt             authService.JWTService
	router          *gin.Engine
	response        *httptest.ResponseRecorder
	initial         map[uuid.UUID]entities.GherkinScenario
	draft           string
	expectedVersion time.Time
	authorization   string
	actorID         uuid.UUID
}

func (f *reviewFeature) reset() {
	f.store = &reviewBDDStore{rows: map[uuid.UUID]entities.GherkinScenario{}}
	f.validator = &reviewBDDValidator{reports: map[string]dto.ScenarioValidation{reviewOriginal: reviewValidReport("User Login", 4)}}
	f.users = reviewBDDUsers{reviewOwnerID: entities.RoleSystemAnalyst, reviewOtherID: entities.RoleSystemAnalyst, reviewEmptyID: entities.RoleSystemAnalyst, reviewQAID: entities.RoleQA}
	f.jwt = authService.NewJWTService()
	f.router = reviewBDDRouter(f.store, f.validator, f.users, f.jwt)
	f.response = nil
	f.initial = nil
	f.draft = reviewOriginal
	f.expectedVersion = time.Time{}
	f.authorization = ""
	f.actorID = uuid.Nil
}

func (f *reviewFeature) captureState() {
	f.initial = map[uuid.UUID]entities.GherkinScenario{}
	for id, row := range f.store.rows {
		f.initial[id] = cloneReviewRow(row)
	}
}

func (f *reviewFeature) persisted(code string) error {
	number, err := strconv.ParseInt(strings.TrimPrefix(code, "SCN-"), 10, 64)
	if err != nil || number <= 0 {
		return fmt.Errorf("invalid fixture scenario code %q", code)
	}
	stamp := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)
	for i, owner := range []uuid.UUID{reviewOwnerID, reviewOtherID} {
		id := reviewScenarioID
		if i == 1 {
			id = uuid.MustParse("00000000-0000-4000-8000-000000000022")
		}
		requestID := uuid.New()
		f.store.rows[id] = entities.GherkinScenario{
			ID: id, ScenarioNumber: number + int64(i), GenerationRequestID: &requestID,
			FeatureName: "User Login", GherkinText: reviewOriginal, SourceType: dto.SourceLLM, Status: dto.ScenarioValid,
			CoveredGuiCount: 4, TotalGuiCount: 4, CoveragePercent: 100,
			GenerationRequest: entities.GenerationRequest{ID: requestID, UserID: owner, GuiElements: append([]string(nil), reviewElements...)},
			Timestamp:         entities.Timestamp{CreatedAt: stamp, UpdatedAt: stamp},
		}
	}
	f.expectedVersion = stamp
	f.captureState()
	return nil
}

func (f *reviewFeature) login(actor string) error {
	actors := map[string]uuid.UUID{"pemilik": reviewOwnerID, "QA": reviewQAID, "analis lain": reviewOtherID, "analis tanpa skenario": reviewEmptyID}
	id, exists := actors[actor]
	if !exists {
		return fmt.Errorf("unknown review actor %q", actor)
	}
	f.actorID = id
	f.authorization = "Bearer " + f.jwt.GenerateAccessToken(id.String(), string(f.users[id]))
	return nil
}

func (f *reviewFeature) session(mode string) error {
	switch mode {
	case "tanpa login":
		f.authorization = ""
	case "memakai token rusak":
		f.authorization = "Bearer not-a-valid-jwt"
	default:
		return fmt.Errorf("unknown session %q", mode)
	}
	return nil
}

func (f *reviewFeature) claimedAnalyst() error {
	if f.actorID == uuid.Nil {
		return fmt.Errorf("login fixture must be configured first")
	}
	f.authorization = "Bearer " + f.jwt.GenerateAccessToken(f.actorID.String(), string(entities.RoleSystemAnalyst))
	return nil
}

func (f *reviewFeature) validDraft(coverage int) error {
	if coverage != 0 && coverage != 75 && coverage != 100 {
		return fmt.Errorf("unsupported coverage fixture %d", coverage)
	}
	f.draft = "Feature: Updated User Login\n  Scenario: Successful login\n    Given a user is ready\n    When the user submits credentials\n    Then access is granted\n"
	if coverage >= 75 {
		f.draft = "Feature: Updated User Login\n  Scenario: Successful login\n    Given the Email field and Password field are filled\n    When the user clicks the Login button\n    Then access is granted\n"
	}
	if coverage == 100 {
		f.draft = strings.Replace(f.draft, "access is granted", "the Dashboard heading is displayed", 1)
	}
	f.validator.reports[f.draft] = reviewValidReport("Updated User Login", coverage/25)
	return nil
}

func (f *reviewFeature) invalidDraft(doc *godog.DocString) error {
	f.draft = doc.Content
	available := false
	f.validator.reports[f.draft] = dto.ScenarioValidation{
		Valid: false, FeatureName: "User Login", Checks: dto.StructureChecks{Feature: true, Scenario: true, When: true, Then: true},
		Issues:       []dto.ValidationIssue{{Line: 3, Column: 5, Message: "Given belum memiliki deskripsi langkah."}},
		UsedElements: []string{}, NotCoveredElements: []string{}, CoverageAvailable: &available,
	}
	return nil
}

func (f *reviewFeature) failure(mode string) error {
	switch mode {
	case "tidak tersedia":
		f.validator.err = service.ErrUnavailable
	case "timeout":
		f.validator.err = service.ErrTimeout
	case "respons rusak":
		f.validator.err = service.ErrBadGateway
	default:
		return fmt.Errorf("unknown validator failure %q", mode)
	}
	return nil
}

func (f *reviewFeature) staleVersion() { f.expectedVersion = f.expectedVersion.Add(-time.Second) }

func (f *reviewFeature) invalidStored() {
	row := f.store.rows[reviewScenarioID]
	row.Status = dto.ScenarioInvalid
	f.store.rows[reviewScenarioID] = row
	f.captureState()
}

func (f *reviewFeature) request(method, path string, payload any) error {
	var body []byte
	if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if f.authorization != "" {
		req.Header.Set("Authorization", f.authorization)
	}
	f.response = httptest.NewRecorder()
	f.router.ServeHTTP(f.response, req)
	return nil
}

func (f *reviewFeature) operation(operation string) error {
	path := "/api/v1/scenarios"
	switch operation {
	case "daftar":
		return f.request(http.MethodGet, path, nil)
	case "detail":
		return f.request(http.MethodGet, path+"/"+reviewScenarioID.String(), nil)
	case "simpan":
		return f.request(http.MethodPut, path+"/"+reviewScenarioID.String(), dto.SaveCorrectionRequest{Gherkin: f.draft, UpdatedAt: f.expectedVersion})
	case "download":
		return f.request(http.MethodGet, path+"/"+reviewScenarioID.String()+"/download", nil)
	default:
		return fmt.Errorf("unknown review operation %q", operation)
	}
}

func (f *reviewFeature) missingVersion() error {
	return f.request(http.MethodPut, "/api/v1/scenarios/"+reviewScenarioID.String(), map[string]string{"gherkin": f.draft})
}

func (f *reviewFeature) status(code int) error {
	if f.response == nil {
		return fmt.Errorf("review request has not been sent")
	}
	if f.response.Code != code {
		return fmt.Errorf("review status=%d, want %d, body=%s", f.response.Code, code, f.response.Body.String())
	}
	return nil
}

func (f *reviewFeature) data(target any) error {
	if f.response == nil {
		return fmt.Errorf("review request has not been sent")
	}
	var envelope struct {
		Status bool            `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(f.response.Body.Bytes(), &envelope); err != nil {
		return err
	}
	if !envelope.Status {
		return fmt.Errorf("expected success envelope, got %s", f.response.Body.String())
	}
	return json.Unmarshal(envelope.Data, target)
}

func (f *reviewFeature) ownList(code string) error {
	var rows []dto.ScenarioSummary
	if err := f.data(&rows); err != nil {
		return err
	}
	if len(rows) != 1 || rows[0].ID != reviewScenarioID.String() || rows[0].ScenarioCode != code {
		return fmt.Errorf("wrong owned scenario list: %+v", rows)
	}
	return nil
}

func (f *reviewFeature) emptyList() error {
	var rows []dto.ScenarioSummary
	if err := f.data(&rows); err != nil {
		return err
	}
	if rows == nil || len(rows) != 0 {
		return fmt.Errorf("expected empty JSON array, got %+v", rows)
	}
	return nil
}

func (f *reviewFeature) detailCode(code string) error {
	var row dto.ReviewScenario
	if err := f.data(&row); err != nil {
		return err
	}
	if row.ID != reviewScenarioID.String() || row.ScenarioCode != code {
		return fmt.Errorf("wrong scenario UUID/code: %+v", row.ScenarioSummary)
	}
	return nil
}

func (f *reviewFeature) storedDraft() error {
	row := f.store.rows[reviewScenarioID]
	if row.GherkinText != f.draft || row.SourceType != dto.SourceCorrected || row.Status != dto.ScenarioValid || !row.UpdatedAt.After(f.initial[reviewScenarioID].UpdatedAt) {
		return fmt.Errorf("correction was not persisted byte-for-byte with valid metadata")
	}
	if f.validator.lastText != f.draft || !reflect.DeepEqual(f.validator.lastElements, reviewElements) {
		return fmt.Errorf("validator did not receive exact draft and original GUI elements")
	}
	return nil
}

func (f *reviewFeature) storedCoverage(want int) error {
	row := f.store.rows[reviewScenarioID]
	if row.CoveragePercent != float64(want) || row.CoveredGuiCount != want/25 || row.TotalGuiCount != 4 {
		return fmt.Errorf("wrong persisted coverage: %+v", row)
	}
	return nil
}

func (f *reviewFeature) detailDraft() error {
	var row dto.ReviewScenario
	if err := f.data(&row); err != nil {
		return err
	}
	if row.Gherkin != f.draft || row.FeatureName != "Updated User Login" || row.SourceType != "CORRECTED" || !row.Validation.Valid {
		return fmt.Errorf("reloaded detail lost saved correction: %+v", row)
	}
	return nil
}

func (f *reviewFeature) validationIssue(line int) error {
	var envelope struct {
		Status bool `json:"status"`
		Data   struct {
			Validation dto.ScenarioValidation `json:"validation"`
		} `json:"data"`
	}
	if err := json.Unmarshal(f.response.Body.Bytes(), &envelope); err != nil {
		return err
	}
	v := envelope.Data.Validation
	if envelope.Status || v.Valid || !v.Checks.Feature || !v.Checks.Scenario || v.Checks.Given || len(v.Issues) != 1 || v.Issues[0].Line != line || !strings.Contains(v.Issues[0].Message, "Given belum memiliki deskripsi") || v.CoverageAvailable == nil || *v.CoverageAvailable {
		return fmt.Errorf("invalid response lost specific diagnostics: %s", f.response.Body.String())
	}
	return nil
}

func (f *reviewFeature) unchanged() error {
	if !reflect.DeepEqual(f.store.rows, f.initial) {
		return fmt.Errorf("rejected request changed stored scenario data")
	}
	return nil
}

func (f *reviewFeature) writes(want int) error {
	if f.store.writes != want {
		return fmt.Errorf("review writes=%d, want %d", f.store.writes, want)
	}
	return nil
}

func (f *reviewFeature) validatorCalls(want int) error {
	if f.validator.calls != want {
		return fmt.Errorf("validator calls=%d, want %d", f.validator.calls, want)
	}
	return nil
}

func (f *reviewFeature) noLeak() error {
	var envelope struct {
		Status bool            `json:"status"`
		Data   json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(f.response.Body.Bytes(), &envelope); err != nil {
		return err
	}
	hasData := len(envelope.Data) > 0 && string(envelope.Data) != "null"
	if envelope.Status || hasData || strings.Contains(f.response.Body.String(), "User Login") || strings.Contains(f.response.Body.String(), reviewScenarioID.String()) {
		return fmt.Errorf("unauthorized response leaked scenario information")
	}
	return nil
}

func (f *reviewFeature) exactDownload() error {
	row := f.store.rows[reviewScenarioID]
	if f.response.Body.String() != row.GherkinText || f.validator.lastText != row.GherkinText {
		return fmt.Errorf("download did not validate and return exact stored bytes")
	}
	media, parameters, err := mime.ParseMediaType(f.response.Header().Get("Content-Disposition"))
	name := parameters["filename"]
	if err != nil || media != "attachment" || !strings.HasSuffix(name, ".feature") || strings.ContainsAny(name, "\r\n/\\") {
		return fmt.Errorf("unsafe download filename: %q", name)
	}
	if f.response.Header().Get("Content-Type") != "text/plain; charset=utf-8" || f.response.Header().Get("Cache-Control") != "no-store" || f.response.Header().Get("X-Content-Type-Options") != "nosniff" {
		return fmt.Errorf("incorrect download response headers")
	}
	return nil
}

func InitializeReviewScenario(ctx *godog.ScenarioContext) {
	f := &reviewFeature{}
	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) { f.reset(); return ctx, nil })
	ctx.Step(`^ada skenario review tersimpan dengan kode "(SCN-\d+)"$`, f.persisted)
	ctx.Step(`^pengguna review login sebagai "([^"]+)"$`, f.login)
	ctx.Step(`^pengguna review "([^"]+)"$`, f.session)
	ctx.Step(`^token review mengklaim role System Analyst$`, f.claimedAnalyst)
	ctx.Step(`^validator menyatakan draft review valid dengan coverage (\d+) persen$`, f.validDraft)
	ctx.Step(`^validator menolak draft review berikut karena Given kosong:$`, f.invalidDraft)
	ctx.Step(`^layanan validator review mengalami "([^"]+)"$`, f.failure)
	ctx.Step(`^pengguna membawa versi review yang sudah kedaluwarsa$`, f.staleVersion)
	ctx.Step(`^versi review tersimpan berstatus INVALID$`, f.invalidStored)
	ctx.Step(`^pengguna menjalankan operasi review "([^"]+)"$`, f.operation)
	ctx.Step(`^pengguna mengirim koreksi review tanpa updated_at$`, f.missingVersion)
	ctx.Step(`^status respons review harus (\d+)$`, f.status)
	ctx.Step(`^daftar review hanya memuat skenario miliknya dengan kode "([^"]+)"$`, f.ownList)
	ctx.Step(`^daftar review kosong$`, f.emptyList)
	ctx.Step(`^respons detail review memuat kode "([^"]+)" dan UUID semula$`, f.detailCode)
	ctx.Step(`^draft review tersimpan persis sebagai CORRECTED dan VALID$`, f.storedDraft)
	ctx.Step(`^coverage review tersimpan sebesar (\d+) persen$`, f.storedCoverage)
	ctx.Step(`^respons detail review berisi koreksi yang sudah disimpan$`, f.detailDraft)
	ctx.Step(`^respons review menjelaskan kesalahan deskripsi Given pada baris (\d+)$`, f.validationIssue)
	ctx.Step(`^data review tersimpan tidak berubah$`, f.unchanged)
	ctx.Step(`^penyimpanan review dilakukan (\d+) kali$`, f.writes)
	ctx.Step(`^validator review dipanggil (\d+) kali$`, f.validatorCalls)
	ctx.Step(`^respons review tidak membocorkan isi skenario$`, f.noLeak)
	ctx.Step(`^file feature berisi versi tersimpan secara persis$`, f.exactDownload)
}

func TestReviewFeatures(t *testing.T) {
	t.Setenv("JWT_SECRET", "review-bdd-only-signing-key-not-an-application-secret")
	gin.SetMode(gin.TestMode)
	suite := godog.TestSuite{
		Name:                "review_scenario",
		ScenarioInitializer: InitializeReviewScenario,
		Options:             &godog.Options{Format: "pretty", Paths: []string{"features/scenario_review.feature"}, TestingT: t, Strict: true},
	}
	if suite.Run() != 0 {
		t.Fatal("Review Scenario BDD suite failed")
	}
}
