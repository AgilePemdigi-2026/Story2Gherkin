package bdd_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/dto"
)

const historyBDDPath = "/api/v1/history"

type historyFeature struct {
	store         *historyBDDStore
	users         historyBDDUsers
	jwt           authService.JWTService
	router        *gin.Engine
	response      *httptest.ResponseRecorder
	authorization string
	actorID       uuid.UUID
	detailID      uuid.UUID
}

// ------------------------------------
// RESET FIXTURE
// ------------------------------------

func (f *historyFeature) reset() {
	f.store = newHistoryBDDStore()

	f.users = historyBDDUsers{
		historyOwnerID: entities.RoleSystemAnalyst,
		historyOtherID: entities.RoleSystemAnalyst,
		historyQAID:    entities.RoleQA,
	}

	f.jwt = authService.NewJWTService()
	f.router = historyBDDRouter(
		f.store,
		f.users,
		f.jwt,
	)

	f.response = nil
	f.authorization = ""
	f.actorID = uuid.Nil
	f.detailID = uuid.Nil
}

// ------------------------------------
// DATA FIXTURES
// ------------------------------------

func (f *historyFeature) seedBoth() {
	f.store.seedSuccess()
	f.store.seedFailed()
}

func (f *historyFeature) seedTwoSuccess() {
	f.store.seedSuccess()
	f.store.seedSecondSuccess()
}

func (f *historyFeature) seedSuccess() {
	f.store.seedSuccess()
}

func (f *historyFeature) seedFailed() {
	f.store.seedFailed()
}

func (f *historyFeature) seedOther() {
	f.store.seedOther()
}

func (f *historyFeature) noHistory() error {
	if len(f.store.requests) != 0 ||
		len(f.store.scenarios) != 0 {
		return fmt.Errorf("history fixture is not empty")
	}

	return nil
}

// ------------------------------------
// LOGIN AND AUTHENTICATION
// ------------------------------------

func (f *historyFeature) login(actor string) error {
	switch actor {
	case "pemilik":
		f.actorID = historyOwnerID

	case "SA lain":
		f.actorID = historyOtherID

	case "QA":
		f.actorID = historyQAID

	default:
		return fmt.Errorf(
			"unknown History actor: %q",
			actor,
		)
	}

	role := f.users[f.actorID]

	f.authorization = "Bearer " +
		f.jwt.GenerateAccessToken(
			f.actorID.String(),
			string(role),
		)

	return nil
}

func (f *historyFeature) noLogin() {
	f.actorID = uuid.Nil
	f.authorization = ""
}

func (f *historyFeature) invalidToken() {
	f.actorID = uuid.Nil
	f.authorization = "Bearer invalid-jwt-token"
}

func (f *historyFeature) claimedAnalyst() error {
	if f.actorID == uuid.Nil {
		return fmt.Errorf(
			"user must be logged in first",
		)
	}

	// JWT claims System Analyst, but the stored
	// user role remains QA.
	f.authorization = "Bearer " +
		f.jwt.GenerateAccessToken(
			f.actorID.String(),
			string(entities.RoleSystemAnalyst),
		)

	return nil
}

// ------------------------------------
// HTTP REQUEST
// ------------------------------------

func (f *historyFeature) request(path string) {
	req := httptest.NewRequest(
		http.MethodGet,
		path,
		nil,
	)

	if f.authorization != "" {
		req.Header.Set(
			"Authorization",
			f.authorization,
		)
	}

	f.response = httptest.NewRecorder()
	f.router.ServeHTTP(f.response, req)
}

func (f *historyFeature) list() {
	f.request(historyBDDPath)
}

func (f *historyFeature) listPage(
	page int,
	pageSize int,
) {
	f.request(
		fmt.Sprintf(
			"%s?page=%d&page_size=%d",
			historyBDDPath,
			page,
			pageSize,
		),
	)
}

func (f *historyFeature) invalidPagination(
	parameter string,
) {
	f.request(historyBDDPath + "?" + parameter)
}

func (f *historyFeature) detailSuccess() {
	f.detailID = historySuccessID

	f.request(
		historyBDDPath + "/" +
			historySuccessID.String(),
	)
}

func (f *historyFeature) detailFailed() {
	f.detailID = historyFailedID

	f.request(
		historyBDDPath + "/" +
			historyFailedID.String(),
	)
}

func (f *historyFeature) detailOther() {
	f.detailID = historyOtherReqID

	f.request(
		historyBDDPath + "/" +
			historyOtherReqID.String(),
	)
}

func (f *historyFeature) detailMissing() {
	f.detailID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000099",
	)

	f.request(
		historyBDDPath + "/" +
			f.detailID.String(),
	)
}

func (f *historyFeature) detailInvalidUUID() {
	f.detailID = uuid.Nil

	f.request(
		historyBDDPath + "/invalid-uuid",
	)
}

func (f *historyFeature) operation(
	operation string,
) error {
	switch operation {
	case "daftar":
		f.list()

	case "detail":
		f.detailSuccess()

	default:
		return fmt.Errorf(
			"unknown History operation: %q",
			operation,
		)
	}

	return nil
}

// ------------------------------------
// RESPONSE HELPERS
// ------------------------------------

func (f *historyFeature) status(
	expected int,
) error {
	if f.response == nil {
		return fmt.Errorf(
			"HTTP request has not been sent",
		)
	}

	if f.response.Code != expected {
		return fmt.Errorf(
			"HTTP status=%d, want=%d; body=%s",
			f.response.Code,
			expected,
			f.response.Body.String(),
		)
	}

	return nil
}

func (f *historyFeature) decode(
	target any,
) error {
	if f.response == nil {
		return fmt.Errorf("response is unavailable")
	}

	var envelope struct {
		Status bool            `json:"status"`
		Data   json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(
		f.response.Body.Bytes(),
		&envelope,
	); err != nil {
		return err
	}

	if !envelope.Status {
		return fmt.Errorf(
			"expected successful API response: %s",
			f.response.Body.String(),
		)
	}

	if len(envelope.Data) == 0 ||
		string(envelope.Data) == "null" {
		return fmt.Errorf("response data is empty")
	}

	return json.Unmarshal(
		envelope.Data,
		target,
	)
}

func (f *historyFeature) listData() (
	dto.HistoryListResponse,
	error,
) {
	var result dto.HistoryListResponse

	err := f.decode(&result)

	return result, err
}

func (f *historyFeature) detailData() (
	dto.HistoryDetail,
	error,
) {
	var result dto.HistoryDetail

	err := f.decode(&result)

	return result, err
}

// ------------------------------------
// LIST ASSERTIONS
// ------------------------------------

func (f *historyFeature) ownList(
	expected int,
) error {
	result, err := f.listData()
	if err != nil {
		return err
	}

	if len(result.History) != expected ||
		result.Total != int64(expected) {
		return fmt.Errorf(
			"history count=%d, total=%d, want=%d",
			len(result.History),
			result.Total,
			expected,
		)
	}

	expectedIDs := map[string]bool{
		historySuccessID.String(): false,
	}

	for _, row := range result.History {
		if _, exists := expectedIDs[row.ID]; !exists {
			return fmt.Errorf(
				"unexpected History ID in list: %s",
				row.ID,
			)
		}

		expectedIDs[row.ID] = true
	}

	for id, found := range expectedIDs {
		if !found {
			return fmt.Errorf(
				"owned History %s is missing",
				id,
			)
		}
	}

	return nil
}

func (f *historyFeature) noOtherHistory() error {
	result, err := f.listData()
	if err != nil {
		return err
	}

	for _, row := range result.History {
		if row.ID == historyOtherReqID.String() ||
			row.UserStory == historyOtherStory {
			return fmt.Errorf(
				"History belonging to another user leaked",
			)
		}
	}

	return nil
}

func (f *historyFeature) emptyList() error {
	result, err := f.listData()
	if err != nil {
		return err
	}

	if result.History == nil ||
		len(result.History) != 0 {
		return fmt.Errorf(
			"expected an empty History array",
		)
	}

	return nil
}

func (f *historyFeature) pagination(
	expectedTotal int,
	expectedPages int,
) error {
	result, err := f.listData()
	if err != nil {
		return err
	}

	if result.Total != int64(expectedTotal) ||
		result.TotalPages != expectedPages {
		return fmt.Errorf(
			"pagination total=%d pages=%d, want=%d/%d",
			result.Total,
			result.TotalPages,
			expectedTotal,
			expectedPages,
		)
	}

	return nil
}

func (f *historyFeature) listCount(
	expected int,
) error {
	result, err := f.listData()
	if err != nil {
		return err
	}

	if len(result.History) != expected {
		return fmt.Errorf(
			"history list size=%d, want=%d",
			len(result.History),
			expected,
		)
	}

	return nil
}

// ------------------------------------
// DETAIL ASSERTIONS
// ------------------------------------

func (f *historyFeature) expectedInput() (
	string,
	[]string,
	string,
	error,
) {
	switch f.detailID {
	case historySuccessID:
		return historySuccessStory,
			historySuccessElements,
			historySuccessFree,
			nil

	case historyFailedID:
		return historyFailedStory,
			historyFailedElements,
			historyFailedFree,
			nil

	default:
		return "", nil, "",
			fmt.Errorf(
				"unsupported History detail fixture",
			)
	}
}

func (f *historyFeature) originalStory() error {
	result, err := f.detailData()
	if err != nil {
		return err
	}

	expected, _, _, err := f.expectedInput()
	if err != nil {
		return err
	}

	if result.ID != f.detailID.String() ||
		result.UserStory != expected {
		return fmt.Errorf(
			"History detail lost original User Story",
		)
	}

	return nil
}

func (f *historyFeature) originalElements() error {
	result, err := f.detailData()
	if err != nil {
		return err
	}

	_, expected, _, err := f.expectedInput()
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(
		result.GUIElements,
		expected,
	) {
		return fmt.Errorf(
			"GUI Elements mismatch: got=%v want=%v",
			result.GUIElements,
			expected,
		)
	}

	return nil
}

func (f *historyFeature) originalFreeScenario() error {
	result, err := f.detailData()
	if err != nil {
		return err
	}

	_, _, expected, err := f.expectedInput()
	if err != nil {
		return err
	}

	if result.FreeScenario != expected {
		return fmt.Errorf(
			"Free Scenario differs from original input",
		)
	}

	return nil
}

func (f *historyFeature) generatedGherkin() error {
	result, err := f.detailData()
	if err != nil {
		return err
	}

	if result.Scenario == nil {
		return fmt.Errorf(
			"SUCCESS History has no generated scenario",
		)
	}

	scenario := result.Scenario

	if scenario.ID != historyScenarioID.String() ||
		scenario.ScenarioCode != "SCN-0007" ||
		scenario.FeatureName != "User Login" ||
		scenario.Gherkin != historySuccessGherkin ||
		scenario.SourceType != "LLM" ||
		scenario.Status != "VALID" ||
		scenario.CoveragePercent != 100 {
		return fmt.Errorf(
			"Generated Gherkin metadata or text is incorrect: %+v",
			scenario,
		)
	}

	return nil
}

func (f *historyFeature) errorMessage() error {
	result, err := f.detailData()
	if err != nil {
		return err
	}

	if result.ErrorMessage != historyFailureMessage {
		return fmt.Errorf(
			"error message=%q, want=%q",
			result.ErrorMessage,
			historyFailureMessage,
		)
	}

	return nil
}

func (f *historyFeature) nullScenario() error {
	result, err := f.detailData()
	if err != nil {
		return err
	}

	if result.Scenario != nil {
		return fmt.Errorf(
			"FAILED History should not contain a scenario",
		)
	}

	return nil
}

func (f *historyFeature) requestStatus(
	expected string,
) error {
	result, err := f.detailData()
	if err != nil {
		return err
	}

	if result.RequestStatus != expected {
		return fmt.Errorf(
			"request status=%q, want=%q",
			result.RequestStatus,
			expected,
		)
	}

	return nil
}

// ------------------------------------
// SECURITY ASSERTION
// ------------------------------------

func (f *historyFeature) noLeak() error {
	if f.response == nil {
		return fmt.Errorf("response is unavailable")
	}

	var envelope struct {
		Status bool            `json:"status"`
		Data   json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(
		f.response.Body.Bytes(),
		&envelope,
	); err != nil {
		return err
	}

	hasData := len(envelope.Data) > 0 &&
		string(envelope.Data) != "null"

	body := f.response.Body.String()

	if envelope.Status ||
		hasData ||
		strings.Contains(body, historyOtherStory) ||
		strings.Contains(
			body,
			historyOtherReqID.String(),
		) {
		return fmt.Errorf(
			"unauthorized response leaked History data",
		)
	}

	return nil
}

// ------------------------------------
// GODog STEP REGISTRATION
// ------------------------------------

func (f *historyFeature) noFailedHistory() error {
	result, err := f.listData()
	if err != nil {
		return err
	}

	for _, item := range result.History {
		if item.ID == historyFailedID.String() ||
			item.UserStory == historyFailedStory ||
			item.RequestStatus == "FAILED" {
			return fmt.Errorf(
				"FAILED generation request should not appear in History",
			)
		}
	}

	return nil
}

func InitializeHistoryScenario(
	ctx *godog.ScenarioContext,
) {
	f := &historyFeature{}

	ctx.Before(func(
		ctx context.Context,
		_ *godog.Scenario,
	) (context.Context, error) {
		f.reset()
		return ctx, nil
	})

	// Fixtures
	ctx.Step(
		`^daftar History tidak memuat request FAILED$`,
		f.noFailedHistory,
	)
	ctx.Step(
		`^tersedia History SUCCESS dan FAILED milik SA$`,
		f.seedBoth,
	)
	ctx.Step(
		`^tersedia 2 History milik SA$`,
		f.seedTwoSuccess,
	)
	ctx.Step(
		`^tersedia History SUCCESS milik SA$`,
		f.seedSuccess,
	)
	ctx.Step(
		`^tersedia History FAILED milik SA tanpa skenario$`,
		f.seedFailed,
	)
	ctx.Step(
		`^tersedia History milik SA lain$`,
		f.seedOther,
	)
	ctx.Step(
		`^tidak ada History tersimpan$`,
		f.noHistory,
	)

	// Authentication
	ctx.Step(
		`^pengguna History login sebagai "([^"]+)"$`,
		f.login,
	)
	ctx.Step(
		`^pengguna History tidak login$`,
		f.noLogin,
	)
	ctx.Step(
		`^pengguna History menggunakan token rusak$`,
		f.invalidToken,
	)
	ctx.Step(
		`^JWT History mengklaim role System Analyst$`,
		f.claimedAnalyst,
	)

	// HTTP requests
	ctx.Step(
		`^pengguna meminta daftar History$`,
		f.list,
	)
	ctx.Step(
		`^pengguna meminta daftar History halaman (\d+) dengan ukuran (\d+)$`,
		f.listPage,
	)
	ctx.Step(
		`^pengguna meminta daftar History dengan parameter "([^"]+)"$`,
		f.invalidPagination,
	)
	ctx.Step(
		`^pengguna meminta detail History SUCCESS$`,
		f.detailSuccess,
	)
	ctx.Step(
		`^pengguna meminta detail History FAILED$`,
		f.detailFailed,
	)
	ctx.Step(
		`^pengguna meminta detail History milik SA lain$`,
		f.detailOther,
	)
	ctx.Step(
		`^pengguna meminta detail History yang tidak tersedia$`,
		f.detailMissing,
	)
	ctx.Step(
		`^pengguna meminta detail History dengan UUID tidak valid$`,
		f.detailInvalidUUID,
	)
	ctx.Step(
		`^pengguna menjalankan operasi History "([^"]+)"$`,
		f.operation,
	)

	// Assertions
	ctx.Step(
		`^status respons History harus (\d+)$`,
		f.status,
	)
	ctx.Step(
		`^daftar History hanya berisi (\d+) riwayat milik pengguna$`,
		f.ownList,
	)
	ctx.Step(
		`^daftar History tidak memuat riwayat pengguna lain$`,
		f.noOtherHistory,
	)
	ctx.Step(
		`^daftar History kosong$`,
		f.emptyList,
	)
	ctx.Step(
		`^pagination History memiliki total (\d+) dan (\d+) halaman$`,
		f.pagination,
	)
	ctx.Step(
		`^daftar History berisi (\d+) item$`,
		f.listCount,
	)
	ctx.Step(
		`^detail History memuat User Story asli$`,
		f.originalStory,
	)
	ctx.Step(
		`^detail History memuat GUI Elements asli$`,
		f.originalElements,
	)
	ctx.Step(
		`^detail History memuat Free Scenario asli$`,
		f.originalFreeScenario,
	)
	ctx.Step(
		`^detail History memuat Generated Gherkin$`,
		f.generatedGherkin,
	)
	ctx.Step(
		`^detail History memuat pesan kegagalan Generate$`,
		f.errorMessage,
	)
	ctx.Step(
		`^hasil skenario History bernilai null$`,
		f.nullScenario,
	)
	ctx.Step(
		`^status History adalah "([^"]+)"$`,
		f.requestStatus,
	)
	ctx.Step(
		`^respons History tidak membocorkan data pengguna lain$`,
		f.noLeak,
	)
	ctx.Step(
		`^respons History tidak membocorkan data FAILED$`,
		f.noFailedLeak,
	)
}

func (f *historyFeature) noFailedLeak() error {
	if f.response == nil {
		return fmt.Errorf("response is unavailable")
	}

	body := f.response.Body.String()

	if strings.Contains(body, historyFailedStory) ||
		strings.Contains(body, historyFailureMessage) ||
		strings.Contains(body, historyFailedID.String()) {
		return fmt.Errorf(
			"FAILED generation data leaked through History detail",
		)
	}

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(
		f.response.Body.Bytes(),
		&envelope,
	); err != nil {
		return err
	}

	if len(envelope.Data) > 0 &&
		string(envelope.Data) != "null" {
		return fmt.Errorf(
			"FAILED History detail should not return data",
		)
	}

	return nil
}

// ------------------------------------
// STRICT BDD TEST SUITE
// ------------------------------------

func TestHistoryFeatures(t *testing.T) {
	t.Setenv(
		"JWT_SECRET",
		"history-bdd-only-signing-key-not-production",
	)

	gin.SetMode(gin.TestMode)

	suite := godog.TestSuite{
		Name:                "generation_history",
		ScenarioInitializer: InitializeHistoryScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/generation_history.feature"},
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Generation History BDD suite failed")
	}
}
