package review_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
	"gorm.io/gorm"
)

const original = "Feature: Original\r\n  Scenario: Success\r\n    Given a user\r\n    When I click Login button\r\n    Then I see a dashboard\r\n"

type userStore struct{ role entities.Role }

func (u *userStore) GetUserById(context.Context, *gorm.DB, string) (entities.User, error) {
	return entities.User{Role: u.role}, nil
}

type scenarioStore struct {
	row    entities.GherkinScenario
	owner  uuid.UUID
	writes int
	err    error
}

func (s *scenarioStore) ListOwned(_ context.Context, owner uuid.UUID) ([]entities.GherkinScenario, error) {
	if s.err != nil {
		return nil, s.err
	}
	if owner != s.owner {
		return []entities.GherkinScenario{}, nil
	}
	return []entities.GherkinScenario{s.row}, nil
}
func (s *scenarioStore) GetOwned(_ context.Context, owner, id uuid.UUID) (entities.GherkinScenario, error) {
	if s.err != nil {
		return entities.GherkinScenario{}, s.err
	}
	if owner != s.owner || id != s.row.ID {
		return entities.GherkinScenario{}, gorm.ErrRecordNotFound
	}
	return s.row, nil
}
func (s *scenarioStore) SaveCorrectionOwned(_ context.Context, owner uuid.UUID, row entities.GherkinScenario, expected time.Time) (entities.GherkinScenario, error) {
	if owner != s.owner || !expected.Equal(s.row.UpdatedAt) {
		return entities.GherkinScenario{}, repository.ErrCorrectionConflict
	}
	if s.err != nil {
		return entities.GherkinScenario{}, s.err
	}
	row.UpdatedAt = s.row.UpdatedAt.Add(time.Second)
	s.row = row
	s.writes++
	return row, nil
}

type validator struct {
	result dto.ScenarioValidation
	err    error
	text   string
	calls  int
}

func (v *validator) Validate(_ context.Context, text string, elements []string) (dto.ScenarioValidation, error) {
	v.text = text
	v.calls++
	return v.result, v.err
}

func fixture() (service.ReviewService, *scenarioStore, *validator, *userStore) {
	owner, requestID := uuid.New(), uuid.New()
	store := &scenarioStore{owner: owner, row: entities.GherkinScenario{
		ID: uuid.New(), GenerationRequestID: &requestID, FeatureName: "Original", GherkinText: original,
		Status: dto.ScenarioValid, SourceType: dto.SourceLLM,
		GenerationRequest: entities.GenerationRequest{ID: requestID, UserID: owner, GuiElements: []string{"Login button", "Missing field"}},
		Timestamp:         entities.Timestamp{UpdatedAt: time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)},
	}}
	v := &validator{result: dto.ScenarioValidation{
		Valid: true, FeatureName: "Corrected", Checks: dto.StructureChecks{Feature: true, Scenario: true, Given: true, When: true, Then: true},
		Issues: []dto.ValidationIssue{}, UsedElements: []string{"Login button"}, NotCoveredElements: []string{"Missing field"}, CoveragePercent: 50,
	}}
	u := &userStore{role: entities.RoleSystemAnalyst}
	return service.NewReviewService(v, u, store), store, v, u
}

func TestValidCorrectionPersistsOriginalDraftAndPartialCoverage(t *testing.T) {
	svc, store, v, _ := fixture()
	text := strings.Replace(original, "Original", "Corrected", 1)
	result, err := svc.Save(context.Background(), store.owner.String(), store.row.ID.String(), dto.SaveCorrectionRequest{Gherkin: text, UpdatedAt: store.row.UpdatedAt})
	if err != nil {
		t.Fatal(err)
	}
	if store.writes != 1 || store.row.GherkinText != text || v.text != text {
		t.Fatal("validated draft must be saved byte-for-byte")
	}
	if result.SourceType != "CORRECTED" || result.Status != "VALID" || store.row.CoveragePercent != 50 {
		t.Fatalf("unexpected correction: %+v", result)
	}
	if result.UpdatedAt.Equal(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("updated_at did not advance")
	}
}

func TestInvalidCorrectionAndUnavailableValidatorDoNotWrite(t *testing.T) {
	for _, unavailable := range []bool{false, true} {
		svc, store, v, _ := fixture()
		v.result.Valid = false
		v.result.Issues = []dto.ValidationIssue{{Line: 3, Column: 1, Message: "Given is required"}}
		if unavailable {
			v.err = service.ErrUnavailable
		}
		_, err := svc.Save(context.Background(), store.owner.String(), store.row.ID.String(), dto.SaveCorrectionRequest{Gherkin: "invalid", UpdatedAt: store.row.UpdatedAt})
		if err == nil || store.writes != 0 || store.row.GherkinText != original || store.row.SourceType != dto.SourceLLM {
			t.Fatal("invalid/unavailable validation mutated saved data")
		}
	}
}

func TestRoleAndOwnershipAreEnforcedForEveryReviewOperation(t *testing.T) {
	for _, qa := range []bool{false, true} {
		svc, store, v, u := fixture()
		user := uuid.NewString()
		if qa {
			user = store.owner.String()
			u.role = entities.RoleQA
		}
		ctx, id := context.Background(), store.row.ID.String()
		_, detailErr := svc.Detail(ctx, user, id)
		_, saveErr := svc.Save(ctx, user, id, dto.SaveCorrectionRequest{Gherkin: original, UpdatedAt: store.row.UpdatedAt})
		_, downloadErr := svc.Download(ctx, user, id)
		for _, err := range []error{detailErr, saveErr, downloadErr} {
			want := service.ErrScenarioNotFound
			if qa {
				want = service.ErrReviewForbidden
			}
			if !errors.Is(err, want) {
				t.Fatalf("got %v, expected %v", err, want)
			}
		}
		items, err := svc.List(ctx, user)
		if qa && !errors.Is(err, service.ErrReviewForbidden) {
			t.Fatal(err)
		}
		if !qa && (err != nil || len(items) != 0) {
			t.Fatal("other user's list leaked data")
		}
		if v.calls != 0 || store.writes != 0 {
			t.Fatal("denied operation reached validator or update")
		}
	}
}

func TestStaleCorrectionCannotOverwriteSavedVersion(t *testing.T) {
	svc, store, v, _ := fixture()
	_, err := svc.Save(context.Background(), store.owner.String(), store.row.ID.String(), dto.SaveCorrectionRequest{Gherkin: "draft", UpdatedAt: store.row.UpdatedAt.Add(-time.Second)})
	if !errors.Is(err, repository.ErrCorrectionConflict) || store.writes != 0 || v.calls != 0 {
		t.Fatal("stale revision was not rejected")
	}
}

func TestDownloadRejectsInvalidStoredVersion(t *testing.T) {
	svc, store, v, _ := fixture()
	v.result.Valid = false
	_, err := svc.Download(context.Background(), store.owner.String(), store.row.ID.String())
	var invalid *service.InvalidCorrectionError
	if !errors.As(err, &invalid) {
		t.Fatalf("got %v", err)
	}
}

func routerFor(svc service.ReviewService, user string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if user != "" {
			c.Set("user_id", user)
		}
		c.Next()
	})
	h := handler.NewReviewHandler(svc)
	r.GET("/scenarios", h.List)
	r.GET("/scenarios/:id", h.Detail)
	r.PUT("/scenarios/:id", h.Save)
	r.GET("/scenarios/:id/download", h.Download)
	return r
}

func TestHTTPDownloadUsesSavedTextAndSafeFilename(t *testing.T) {
	svc, store, _, _ := fixture()
	store.row.FeatureName = "../../CON\r\nX-Injected: yes"
	r := routerFor(svc, store.owner.String())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/scenarios/"+store.row.ID.String()+"/download", nil))
	if rec.Code != 200 || rec.Body.String() != original {
		t.Fatalf("download = %d %q", rec.Code, rec.Body.String())
	}
	header := rec.Header().Get("Content-Disposition")
	if strings.ContainsAny(header, "\r\n/\\") || !strings.Contains(header, "scenario-") || !strings.Contains(header, ".feature") {
		t.Fatal(header)
	}
}

func TestHTTPInvalidSaveReturnsLineDetailsWithoutWriting(t *testing.T) {
	svc, store, v, _ := fixture()
	v.result.Valid = false
	v.result.Issues = []dto.ValidationIssue{{Line: 4, Column: 5, Message: "Then must follow When"}}
	body, _ := json.Marshal(dto.SaveCorrectionRequest{Gherkin: "bad draft", UpdatedAt: store.row.UpdatedAt})
	req := httptest.NewRequest("PUT", "/scenarios/"+store.row.ID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	routerFor(svc, store.owner.String()).ServeHTTP(rec, req)
	if rec.Code != 422 || !strings.Contains(rec.Body.String(), "Save failed. Invalid Gherkin structure.") || !strings.Contains(rec.Body.String(), `"line":4`) || store.writes != 0 {
		t.Fatalf("invalid response %d %s", rec.Code, rec.Body.String())
	}
}

func TestHTTPAuthorization(t *testing.T) {
	for _, user := range []string{"", "qa"} {
		svc, store, _, u := fixture()
		id := user
		status := 401
		if user == "qa" {
			id = store.owner.String()
			u.role = entities.RoleQA
			status = 403
		}
		rec := httptest.NewRecorder()
		routerFor(svc, id).ServeHTTP(rec, httptest.NewRequest("GET", "/scenarios", nil))
		if rec.Code != status {
			t.Fatalf("got %d, expected %d", rec.Code, status)
		}
	}
}

func TestValidationClientCallsPrivateParserAndChecksResponse(t *testing.T) {
	_, _, v, _ := fixture()
	for _, malformed := range []bool{false, true} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/internal/scenarios/validate" || r.Header.Get("X-AI-Service-Key") != strings.Repeat("k", 32) {
				t.Error("incorrect service request")
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["gherkin"] != original {
				t.Error("wrong text sent to parser")
			}
			if malformed {
				_, _ = w.Write([]byte(`{"valid":true}`))
				return
			}
			_ = json.NewEncoder(w).Encode(v.result)
		}))
		client, err := service.NewLLMClient(server.URL, strings.Repeat("k", 32))
		if err != nil {
			t.Fatal(err)
		}
		result, err := client.Validate(context.Background(), original, []string{"Login button", "Missing field"})
		server.Close()
		if malformed {
			if !errors.Is(err, service.ErrBadGateway) {
				t.Fatal("malformed report accepted")
			}
		} else if err != nil || !result.Valid || result.CoveragePercent != 50 {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	}
}
