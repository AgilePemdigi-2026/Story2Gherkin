package bdd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/service"
)

const goldenBDDPath = "/api/v1/golden-datasets"

var goldenFixtureID = uuid.MustParse(
	"00000000-0000-4000-8000-000000000041",
)

type goldenFeature struct {
	store     *goldenBDDStore
	validator *goldenBDDValidator
	users     goldenBDDUsers
	jwt       authService.JWTService
	router    *gin.Engine

	response      *httptest.ResponseRecorder
	authorization string
	rawFile       []byte
	datasetID     uuid.UUID
}

// ------------------------------------
// RESET FIXTURE
// ------------------------------------

func (f *goldenFeature) reset() {
	f.store = newGoldenBDDStore()

	f.validator = &goldenBDDValidator{
		valid: true,
	}

	f.users = goldenBDDUsers{
		goldenQAID: entities.RoleQA,
		goldenSAID: entities.RoleSystemAnalyst,
	}

	f.jwt = authService.NewJWTService()

	f.router = goldenBDDRouter(
		f.store,
		f.validator,
		f.users,
		f.jwt,
	)

	f.response = nil
	f.authorization = ""
	f.rawFile = validGoldenFile(2)
	f.datasetID = goldenFixtureID
}

// ------------------------------------
// JSON FIXTURES
// ------------------------------------

func validGoldenFile(count int) []byte {
	items := make([]map[string]any, 0, count)

	for i := 1; i <= count; i++ {
		items = append(items, map[string]any{
			"item_title": fmt.Sprintf(
				"Login Scenario %d", i,
			),
			"user_story_text": "As a user, I want to log in " +
				"so that I can access my dashboard.",
			"gui_elements": []string{
				"Email field",
				"Password field",
				"Login button",
			},
			"free_scenario_text": "The user enters " +
				"valid credentials and logs in.",
			"expected_gherkin": "Feature: User Login\n" +
				"  Scenario: Successful Login\n" +
				"    Given the user is on the login page\n" +
				"    When the user enters valid credentials\n" +
				"    Then the dashboard is displayed",
		})
	}

	raw, err := json.Marshal(map[string]any{
		"items": items,
	})

	if err != nil {
		panic(err)
	}

	return raw
}

func (f *goldenFeature) validFile(count int) error {
	if count < 1 || count > 100 {
		return fmt.Errorf("unsupported fixture size %d", count)
	}

	f.rawFile = validGoldenFile(count)
	return nil
}

func (f *goldenFeature) malformedFile() {
	f.rawFile = []byte(`{"items": [invalid json]}`)
}

func (f *goldenFeature) duplicateTitles() error {
	var data struct {
		Items []map[string]any `json:"items"`
	}

	if err := json.Unmarshal(validGoldenFile(2), &data); err != nil {
		return err
	}

	data.Items[1]["item_title"] = "  LOGIN   scenario 1 "

	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}

	f.rawFile = raw
	return nil
}

func (f *goldenFeature) oversizedFile() {
	// Larger than the whole-request limit of 5 MiB.
	f.rawFile = bytes.Repeat([]byte("A"), (5<<20)+1)
}

// ------------------------------------
// USER SESSIONS
// ------------------------------------

func (f *goldenFeature) login(role string) error {
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

	token := f.jwt.GenerateAccessToken(
		userID.String(),
		string(userRole),
	)

	f.authorization = "Bearer " + token
	return nil
}

func (f *goldenFeature) noLogin() {
	f.authorization = ""
}

// ------------------------------------
// VALIDATOR BEHAVIOR
// ------------------------------------

func (f *goldenFeature) validatorValid() {
	f.validator.valid = true
	f.validator.err = nil
}

func (f *goldenFeature) validatorInvalid() {
	f.validator.valid = false
	f.validator.err = nil
}

func (f *goldenFeature) validatorUnavailable() {
	f.validator.err = service.ErrUnavailable
}

// ------------------------------------
// STORED DATA FIXTURE
// ------------------------------------

func (f *goldenFeature) existingDataset(count int) error {
	if count < 1 {
		return fmt.Errorf("invalid item count")
	}

	items := make([]entities.GoldenDatasetItem, 0, count)

	for i := 1; i <= count; i++ {
		items = append(items, entities.GoldenDatasetItem{
			ItemTitle: fmt.Sprintf("Scenario %d", i),
			UserStoryText: "As a user, I want to log in " +
				"so that I can access my dashboard.",
			GUIElements: []string{
				"Email field",
				"Login button",
			},
			FreeScenarioText: "User enters credentials.",
			ExpectedGherkin: "Feature: Login\n" +
				"  Scenario: Successful Login\n" +
				"    Given a login page\n" +
				"    When credentials are submitted\n" +
				"    Then access is granted",
		})
	}

	dataset := &entities.GoldenDataset{
		ID:        goldenFixtureID,
		Name:      "Golden Dataset Test",
		CreatedBy: goldenQAID.String(),
	}

	if err := f.store.CreateWithItems(
		context.Background(),
		dataset,
		items,
	); err != nil {
		return err
	}

	// Fixture setup is not counted as an API write.
	f.store.writes = 0

	return nil
}

// ------------------------------------
// HTTP REQUEST HELPERS
// ------------------------------------

func (f *goldenFeature) sendRequest(
	method string,
	path string,
	body *bytes.Buffer,
	contentType string,
) {
	var request *http.Request

	if body != nil {
		request = httptest.NewRequest(method, path, body)
	} else {
		request = httptest.NewRequest(
			method, path, http.NoBody,
		)
	}

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	if f.authorization != "" {
		request.Header.Set(
			"Authorization",
			f.authorization,
		)
	}

	f.response = httptest.NewRecorder()
	f.router.ServeHTTP(f.response, request)
}

func (f *goldenFeature) upload() error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField(
		"name", "BDD Golden Dataset Test",
	); err != nil {
		return err
	}

	if err := writer.WriteField(
		"description", "Golden Dataset BDD fixture",
	); err != nil {
		return err
	}

	part, err := writer.CreateFormFile(
		"file", "golden-test.json",
	)
	if err != nil {
		return err
	}

	if _, err := part.Write(f.rawFile); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	f.sendRequest(
		http.MethodPost,
		goldenBDDPath,
		&body,
		writer.FormDataContentType(),
	)

	return nil
}

func (f *goldenFeature) list() {
	f.sendRequest(
		http.MethodGet,
		goldenBDDPath,
		nil,
		"",
	)
}

func (f *goldenFeature) detail() {
	f.sendRequest(
		http.MethodGet,
		goldenBDDPath+"/"+f.datasetID.String(),
		nil,
		"",
	)
}

func (f *goldenFeature) missingDetail() {
	f.sendRequest(
		http.MethodGet,
		goldenBDDPath+
			"/00000000-0000-4000-8000-000000000099",
		nil,
		"",
	)
}

func (f *goldenFeature) operation(operation string) error {
	switch operation {
	case "daftar":
		f.list()
	case "detail":
		f.detail()
	default:
		return fmt.Errorf("unknown operation %q", operation)
	}

	return nil
}

// ------------------------------------
// ASSERTIONS
// ------------------------------------

func (f *goldenFeature) status(want int) error {
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

func (f *goldenFeature) storedItems(want int) error {
	if len(f.store.rows) != 1 {
		return fmt.Errorf(
			"stored datasets=%d, want=1",
			len(f.store.rows),
		)
	}

	for _, dataset := range f.store.rows {
		if len(dataset.Items) != want {
			return fmt.Errorf(
				"stored items=%d, want=%d",
				len(dataset.Items),
				want,
			)
		}

		if dataset.ItemCount != want {
			return fmt.Errorf(
				"item_count=%d, want=%d",
				dataset.ItemCount,
				want,
			)
		}

		if f.store.writes != 1 {
			return fmt.Errorf(
				"writes=%d, want=1",
				f.store.writes,
			)
		}
	}

	return nil
}

func (f *goldenFeature) noNewDataset() error {
	if len(f.store.rows) != 0 {
		return fmt.Errorf(
			"unexpected stored datasets: %d",
			len(f.store.rows),
		)
	}

	if f.store.writes != 0 {
		return fmt.Errorf(
			"unexpected writes: %d",
			f.store.writes,
		)
	}

	return nil
}

func (f *goldenFeature) validatorCalls(want int) error {
	if f.validator.calls != want {
		return fmt.Errorf(
			"validator calls=%d, want=%d",
			f.validator.calls,
			want,
		)
	}

	return nil
}

// Decode the standard API envelope.
func (f *goldenFeature) responseData() (json.RawMessage, error) {
	if f.response == nil {
		return nil, fmt.Errorf("response is unavailable")
	}

	var envelope struct {
		Status bool            `json:"status"`
		Data   json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(
		f.response.Body.Bytes(),
		&envelope,
	); err != nil {
		return nil, err
	}

	if !envelope.Status {
		return nil, fmt.Errorf(
			"API did not report success: %s",
			f.response.Body.String(),
		)
	}

	return envelope.Data, nil
}

func (f *goldenFeature) listCount(want int) error {
	raw, err := f.responseData()
	if err != nil {
		return err
	}

	var result struct {
		Total    int64             `json:"total"`
		Datasets []json.RawMessage `json:"datasets"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return err
	}

	if len(result.Datasets) != want ||
		result.Total != int64(want) {
		return fmt.Errorf(
			"list items=%d, total=%d, want=%d",
			len(result.Datasets),
			result.Total,
			want,
		)
	}

	return nil
}

func (f *goldenFeature) detailItems(want int) error {
	raw, err := f.responseData()
	if err != nil {
		return err
	}

	var result struct {
		ID    string `json:"id"`
		Items []struct {
			ItemTitle       string   `json:"item_title"`
			GUIElements     []string `json:"gui_elements"`
			ExpectedGherkin string   `json:"expected_gherkin"`
		} `json:"items"`
	}

	if err := json.Unmarshal(raw, &result); err != nil {
		return err
	}

	if result.ID != f.datasetID.String() {
		return fmt.Errorf(
			"unexpected dataset ID: %s",
			result.ID,
		)
	}

	if len(result.Items) != want {
		return fmt.Errorf(
			"detail items=%d, want=%d",
			len(result.Items),
			want,
		)
	}

	for _, item := range result.Items {
		if strings.TrimSpace(item.ItemTitle) == "" ||
			len(item.GUIElements) == 0 ||
			strings.TrimSpace(item.ExpectedGherkin) == "" {
			return fmt.Errorf(
				"detail contains incomplete item",
			)
		}
	}

	return nil
}

// ------------------------------------
// GODog STEP REGISTRATION
// ------------------------------------

func InitializeGoldenDatasetScenario(
	ctx *godog.ScenarioContext,
) {
	f := &goldenFeature{}

	ctx.Before(func(
		ctx context.Context,
		_ *godog.Scenario,
	) (context.Context, error) {
		f.reset()
		return ctx, nil
	})

	ctx.Step(
		`^pengguna Golden Dataset login sebagai "([^"]+)"$`,
		f.login,
	)

	ctx.Step(
		`^pengguna Golden Dataset tidak login$`,
		f.noLogin,
	)

	ctx.Step(
		`^file Golden Dataset berisi (\d+) item valid$`,
		f.validFile,
	)

	ctx.Step(
		`^file Golden Dataset berisi JSON yang rusak$`,
		f.malformedFile,
	)

	ctx.Step(
		`^file Golden Dataset memiliki judul item duplikat$`,
		f.duplicateTitles,
	)

	ctx.Step(
		`^file Golden Dataset melebihi batas 5 MB$`,
		f.oversizedFile,
	)

	ctx.Step(
		`^validator Golden Dataset menyatakan semua Gherkin valid$`,
		f.validatorValid,
	)

	ctx.Step(
		`^validator Golden Dataset menyatakan Gherkin invalid$`,
		f.validatorInvalid,
	)

	ctx.Step(
		`^layanan validator Golden Dataset tidak tersedia$`,
		f.validatorUnavailable,
	)

	ctx.Step(
		`^terdapat 1 Golden Dataset yang sudah tersimpan$`,
		func() error {
			return f.existingDataset(2)
		},
	)

	ctx.Step(
		`^pengguna mengunggah Golden Dataset$`,
		f.upload,
	)

	ctx.Step(
		`^pengguna meminta daftar Golden Dataset$`,
		f.list,
	)

	ctx.Step(
		`^pengguna meminta detail Golden Dataset$`,
		f.detail,
	)

	ctx.Step(
		`^pengguna meminta detail Golden Dataset yang tidak tersedia$`,
		f.missingDetail,
	)

	ctx.Step(
		`^pengguna menjalankan operasi Golden Dataset "([^"]+)"$`,
		f.operation,
	)

	ctx.Step(
		`^status respons Golden Dataset harus (\d+)$`,
		f.status,
	)

	ctx.Step(
		`^Golden Dataset tersimpan dengan (\d+) item$`,
		f.storedItems,
	)

	ctx.Step(
		`^tidak ada Golden Dataset baru yang tersimpan$`,
		f.noNewDataset,
	)

	ctx.Step(
		`^validator Golden Dataset dipanggil (\d+) kali$`,
		f.validatorCalls,
	)

	ctx.Step(
		`^daftar Golden Dataset berisi (\d+) dataset$`,
		f.listCount,
	)

	ctx.Step(
		`^detail Golden Dataset memiliki (\d+) item$`,
		f.detailItems,
	)
}

// ------------------------------------
// STRICT BDD TEST SUITE
// ------------------------------------

func TestGoldenDatasetFeatures(t *testing.T) {
	t.Setenv(
		"JWT_SECRET",
		"golden-dataset-bdd-only-signing-key-not-production",
	)

	gin.SetMode(gin.TestMode)

	suite := godog.TestSuite{
		Name:                "golden_dataset",
		ScenarioInitializer: InitializeGoldenDatasetScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/golden_dataset.feature"},
			TestingT: t,
			Strict:   true,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("Golden Dataset BDD suite failed")
	}
}
