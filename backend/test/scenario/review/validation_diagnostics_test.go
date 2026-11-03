package review_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
)

func TestValidationClientPreservesUnavailableCoverageAndRejectsContradictions(t *testing.T) {
	for _, test := range []struct {
		name      string
		valid     bool
		percent   float64
		uncovered []string
		wantError bool
	}{
		{"parser failure", false, 0, []string{}, false},
		{"valid document without coverage", true, 0, []string{}, true},
		{"unavailable coverage with percentage", false, 50, []string{}, true},
		{"unavailable coverage with uncovered list", false, 0, []string{"Login button"}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			available := false
			report := dto.ScenarioValidation{
				Valid: test.valid, FeatureName: "User login",
				Checks:       dto.StructureChecks{Feature: true, Scenario: true},
				Issues:       []dto.ValidationIssue{{Line: 4, Column: 5, Message: "Given belum memiliki deskripsi langkah."}},
				UsedElements: []string{}, NotCoveredElements: test.uncovered,
				CoveragePercent: test.percent, CoverageAvailable: &available,
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(report)
			}))
			defer server.Close()
			client, err := service.NewLLMClient(server.URL, strings.Repeat("k", 32))
			if err != nil {
				t.Fatal(err)
			}
			result, err := client.Validate(context.Background(), "invalid original", []string{"Login button"})
			if test.wantError {
				if !errors.Is(err, service.ErrBadGateway) {
					t.Fatalf("got %v", err)
				}
				return
			}
			if err != nil || result.CoverageAvailable == nil || *result.CoverageAvailable ||
				!result.Checks.Feature || !result.Checks.Scenario || len(result.Issues) != 1 {
				t.Fatalf("lost diagnostics: %+v, %v", result, err)
			}
		})
	}
}

func TestHTTPInvalidSaveKeepsDiagnosticsAndSavedVersion(t *testing.T) {
	svc, store, validator, _ := fixture()
	available := false
	validator.result.Valid = false
	validator.result.Checks = dto.StructureChecks{Feature: true, Scenario: true}
	validator.result.Issues = []dto.ValidationIssue{{Line: 4, Column: 5, Message: "Given belum memiliki deskripsi langkah."}}
	validator.result.CoverageAvailable = &available
	validator.result.CoveragePercent = 0
	validator.result.UsedElements = []string{}
	validator.result.NotCoveredElements = []string{}
	body, _ := json.Marshal(dto.SaveCorrectionRequest{Gherkin: "draft without descriptions", UpdatedAt: store.row.UpdatedAt})
	req := httptest.NewRequest("PUT", "/scenarios/"+store.row.ID.String(), strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	routerFor(svc, store.owner.String()).ServeHTTP(response, req)
	if response.Code != 422 || !strings.Contains(response.Body.String(), `"coverage_available":false`) ||
		!strings.Contains(response.Body.String(), "Given belum memiliki deskripsi langkah.") ||
		store.writes != 0 || store.row.GherkinText != original {
		t.Fatalf("unexpected response or changed saved text: %d %s", response.Code, response.Body.String())
	}
}
