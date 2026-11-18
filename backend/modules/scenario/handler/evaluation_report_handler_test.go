package handler_test

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
	"gorm.io/gorm"
)

const (
	evaluationHTTPBase   = "/api/v1/evaluation-runs"
	evaluationHTTPUserID = "469075dd-2376-4b21-aebc-a842f571173a"
	evaluationHTTPRunID  = "6b308226-0b1f-4efb-92c6-9b4066132b74"
)

func TestEvaluationReportHTTPRequiresAuthentication(t *testing.T) {
	headers := []struct {
		name  string
		value string
	}{
		{"missing token", ""},
		{"invalid token", "Bearer invalid-token"},
		{"wrong scheme", "Basic invalid-token"},
	}

	for _, path := range evaluationHTTPPaths() {
		for _, header := range headers {
			t.Run(path+"/"+header.name, func(t *testing.T) {
				f := newEvaluationHTTPFixture(t)

				rec := f.get(path, header.value)
				assertEvaluationHTTPError(t, rec, http.StatusUnauthorized)

				if f.users.calls != 0 ||
					f.reports.listCalls != 0 ||
					f.reports.detailCalls != 0 {
					t.Fatal("unauthenticated requests must not reach storage")
				}
			})
		}
	}
}

func TestEvaluationReportHTTPUsesDatabaseRole(t *testing.T) {
	for _, path := range evaluationHTTPPaths() {
		t.Run(path, func(t *testing.T) {
			f := newEvaluationHTTPFixture(t)
			f.reports.report.ReportCode = "ER-0025"
			// The token claims QA, but the current database role is different.
			f.users.role = entities.RoleSystemAnalyst

			rec := f.get(path, f.authorization)
			assertEvaluationHTTPError(t, rec, http.StatusForbidden)

			if f.users.calls != 1 {
				t.Fatal("the current database role must be checked")
			}

			if f.reports.listCalls != 0 || f.reports.detailCalls != 0 {
				t.Fatal("forbidden requests must not reach report storage")
			}
		})
	}
}

func TestEvaluationReportHTTPList(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		empty     bool
		total     int64
		wantPage  int
		wantSize  int
		wantPages int
	}{
		{"defaults", "", false, 1, 1, 10, 1},
		{"custom pagination", "?page=2&page_size=1", false, 3, 2, 1, 3},
		{"empty list", "", true, 0, 1, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEvaluationHTTPFixture(t)
			f.reports.total = tt.total

			if tt.empty {
				f.reports.runs = nil
			}

			rec := f.get(evaluationHTTPBase+tt.query, f.authorization)
			data := evaluationHTTPData[dto.EvaluationRunListResponse](t, rec)

			if data.Runs == nil {
				t.Fatal("runs must be an array, not null")
			}

			if data.TotalPages != tt.wantPages {
				t.Fatalf(
					"total_pages = %d, want %d",
					data.TotalPages,
					tt.wantPages,
				)
			}

			wantCount := 1
			if tt.empty {
				wantCount = 0
			}

			if len(data.Runs) != wantCount {
				t.Fatalf("run count = %d, want %d", len(data.Runs), wantCount)
			}

			if !tt.empty && data.Runs[0].RunID != evaluationHTTPRunID {
				t.Fatal("the stored run was not returned")
			}

			if f.reports.page != tt.wantPage ||
				f.reports.pageSize != tt.wantSize {
				t.Fatal("incorrect pagination parameters reached the repository")
			}
		})
	}
}

func TestEvaluationReportHTTPDetailPreservesJSONValues(t *testing.T) {
	for _, status := range []string{"QUEUED", "RUNNING", "COMPLETED", "FAILED"} {
		t.Run(status, func(t *testing.T) {
			f := newEvaluationHTTPFixture(t)
			f.reports.report.Status = status

			if status == "COMPLETED" {
				zero := 0.0
				f.reports.report.Metrics = &dto.EvaluationReportMetrics{
					LLM: dto.EvaluationMethodSummary{
						ValidationPercent: &zero,
					},
				}
			} else {
				f.reports.report.Cases = nil
				f.reports.report.Metrics = nil
			}

			if status == "FAILED" {
				message := "Evaluation could not be completed."
				f.reports.report.ErrorMessage = &message
			}

			rec := f.get(
				evaluationHTTPBase+"/"+evaluationHTTPRunID,
				f.authorization,
			)

			got := evaluationHTTPData[dto.EvaluationReportDetailResponse](t, rec)

			want := f.reports.report
			if want.Cases == nil {
				want.Cases = []dto.EvaluationReportCase{}
			}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("report = %#v, want %#v", got, want)
			}

			// Verify that nullable JSON fields are present, not omitted.
			var envelope struct {
				Data map[string]json.RawMessage `json:"data"`
			}

			if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
				t.Fatal(err)
			}

			for _, field := range []string{"metrics", "cases", "error_message"} {
				if _, exists := envelope.Data[field]; !exists {
					t.Errorf("required field %q is missing", field)
				}
			}
		})
	}
}

func TestEvaluationReportHTTPInvalidParametersAndMissingRuns(t *testing.T) {
	missingID := "178663c3-d19b-4b8a-97f2-9e1073d58363"

	tests := []struct {
		suffix string
		status int
	}{
		{"?page=abc", http.StatusBadRequest},
		{"?page=0", http.StatusBadRequest},
		{"?page=-1", http.StatusBadRequest},
		{"?page=1000001", http.StatusBadRequest},
		{"?page_size=abc", http.StatusBadRequest},
		{"?page_size=0", http.StatusBadRequest},
		{"?page_size=101", http.StatusBadRequest},
		{"/invalid-id", http.StatusBadRequest},
		{"/invalid-id/download", http.StatusBadRequest},
		{"/" + missingID, http.StatusNotFound},
		{"/" + missingID + "/download", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.suffix, func(t *testing.T) {
			f := newEvaluationHTTPFixture(t)

			rec := f.get(evaluationHTTPBase+tt.suffix, f.authorization)
			assertEvaluationHTTPError(t, rec, tt.status)
		})
	}
}

func TestEvaluationReportHTTPDownload(t *testing.T) {
	f := newEvaluationHTTPFixture(t)

	rec := f.get(
		evaluationHTTPBase+"/"+evaluationHTTPRunID+"/download",
		f.authorization,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}

	wantDisposition := `attachment; filename="ER-0001.csv"`

	if got := rec.Header().Get("Content-Disposition"); got != wantDisposition {
		t.Fatalf("Content-Disposition = %q, want %q", got, wantDisposition)
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("report downloads must use Cache-Control: no-store")
	}

	reader := csv.NewReader(rec.Body)
	reader.FieldsPerRecord = 7

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	item := f.reports.report.Cases[0]

	want := [][]string{
		{
			"case_id",
			"title",
			"expected_gherkin",
			"llm_status",
			"llm_gui_coverage_percent",
			"rule_based_status",
			"rule_based_gui_coverage_percent",
		},
		{
			item.CaseID,
			item.Title,
			item.ExpectedGherkin,
			"FAIL",
			"0",
			"",
			"",
		},
	}

	if !reflect.DeepEqual(records, want) {
		t.Fatalf("CSV records = %#v, want %#v", records, want)
	}
}

func TestEvaluationReportHTTPRejectsUnfinishedDownloads(t *testing.T) {
	for _, status := range []string{"QUEUED", "RUNNING", "FAILED"} {
		t.Run(status, func(t *testing.T) {
			f := newEvaluationHTTPFixture(t)
			f.reports.report.Status = status

			rec := f.get(
				evaluationHTTPBase+"/"+evaluationHTTPRunID+"/download",
				f.authorization,
			)

			assertEvaluationHTTPError(t, rec, http.StatusConflict)

			if rec.Header().Get("Content-Disposition") != "" {
				t.Fatal("a rejected download must not include an attachment header")
			}
		})
	}
}

func TestEvaluationReportHTTPHidesStorageErrors(t *testing.T) {
	const privateDetail = "PRIVATE_STORAGE_FAILURE"

	for _, path := range evaluationHTTPPaths() {
		t.Run(path, func(t *testing.T) {
			f := newEvaluationHTTPFixture(t)
			f.reports.err = errors.New(privateDetail)

			rec := f.get(path, f.authorization)
			assertEvaluationHTTPError(t, rec, http.StatusInternalServerError)

			if strings.Contains(rec.Body.String(), privateDetail) {
				t.Fatal("internal storage details must not appear in the response")
			}

			if rec.Header().Get("Content-Disposition") != "" {
				t.Fatal("a failed request must not return an attachment header")
			}
		})
	}
}

func TestEvaluationReportHTTPRejectsInvalidCSVBeforeSendingDownload(t *testing.T) {
	f := newEvaluationHTTPFixture(t)
	invalidCoverage := 101.0
	f.reports.report.Cases[0].LLM.GUICoveragePercent = &invalidCoverage

	rec := f.get(
		evaluationHTTPBase+"/"+evaluationHTTPRunID+"/download",
		f.authorization,
	)

	assertEvaluationHTTPError(t, rec, http.StatusInternalServerError)

	if rec.Header().Get("Content-Disposition") != "" {
		t.Fatal("CSV generation errors must not produce an attachment")
	}
}

// Test fixtures use the actual service and JWT middleware.

type evaluationHTTPUsers struct {
	role  entities.Role
	calls int
}

func (s *evaluationHTTPUsers) GetUserById(
	_ context.Context,
	_ *gorm.DB,
	_ string,
) (entities.User, error) {
	s.calls++
	return entities.User{Role: s.role}, nil
}

type evaluationHTTPRepository struct {
	runs        []dto.EvaluationRunSummary
	total       int64
	report      dto.EvaluationReportDetailResponse
	err         error
	listCalls   int
	detailCalls int
	page        int
	pageSize    int
}

func (s *evaluationHTTPRepository) List(
	_ context.Context,
	page int,
	pageSize int,
) ([]dto.EvaluationRunSummary, int64, error) {
	s.listCalls++
	s.page = page
	s.pageSize = pageSize
	return s.runs, s.total, s.err
}

func (s *evaluationHTTPRepository) GetByID(
	_ context.Context,
	runID uuid.UUID,
) (dto.EvaluationReportDetailResponse, error) {
	s.detailCalls++

	if s.err != nil {
		return dto.EvaluationReportDetailResponse{}, s.err
	}

	if runID.String() != s.report.RunID {
		return dto.EvaluationReportDetailResponse{},
			repository.ErrEvaluationReportNotFound
	}

	return s.report, nil
}

type evaluationHTTPFixture struct {
	router        *gin.Engine
	users         *evaluationHTTPUsers
	reports       *evaluationHTTPRepository
	authorization string
}

func newEvaluationHTTPFixture(t *testing.T) *evaluationHTTPFixture {
	t.Helper()

	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "evaluation-report-http-tests-only-secret")

	users := &evaluationHTTPUsers{role: entities.RoleQA}
	zero := 0.0

	report := dto.EvaluationReportDetailResponse{
		EvaluationRunSummary: dto.EvaluationRunSummary{
			RunID:       evaluationHTTPRunID,
			ReportCode:  "ER-0001",
			DatasetName: "Golden Dataset snapshot",
			CreatedAt:   time.Date(2026, 11, 17, 2, 0, 0, 0, time.UTC),
			TotalCases:  1,
			Status:      "COMPLETED",
		},
		Cases: []dto.EvaluationReportCase{
			{
				CaseID: "GC-001",
				Title:  "Login, \"admin\"",
				ExpectedGherkin: "Feature: Login\n" +
					"  Scenario: Access dashboard\n" +
					"    Given the login page is open\n" +
					"    When the user submits valid credentials\n" +
					"    Then the dashboard appears",
				LLM: &dto.EvaluationMethodResult{
					Status:             "FAIL",
					GUICoveragePercent: &zero,
				},
			},
		},
	}

	reports := &evaluationHTTPRepository{
		runs:   []dto.EvaluationRunSummary{report.EvaluationRunSummary},
		total:  1,
		report: report,
	}

	svc := service.NewEvaluationReportService(users, reports)
	h := handler.NewEvaluationReportHandler(svc)
	jwt := authService.NewJWTService()

	router := gin.New()

	scenario.RegisterEvaluationReportRoutes(
		router.Group("/api/v1"),
		jwt,
		h,
	)

	return &evaluationHTTPFixture{
		router:  router,
		users:   users,
		reports: reports,
		authorization: "Bearer " + jwt.GenerateAccessToken(
			evaluationHTTPUserID,
			string(entities.RoleQA),
		),
	}
}

func (f *evaluationHTTPFixture) get(
	path string,
	authorization string,
) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)

	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}

	rec := httptest.NewRecorder()
	f.router.ServeHTTP(rec, req)
	return rec
}

func evaluationHTTPPaths() []string {
	return []string{
		evaluationHTTPBase,
		evaluationHTTPBase + "/" + evaluationHTTPRunID,
		evaluationHTTPBase + "/" + evaluationHTTPRunID + "/download",
	}
}

func evaluationHTTPData[T any](
	t *testing.T,
	rec *httptest.ResponseRecorder,
) T {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Fatal("expected a JSON response")
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("report responses must use Cache-Control: no-store")
	}

	var envelope struct {
		Status *bool `json:"status"`
		Data   T     `json:"data"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}

	if envelope.Status == nil || !*envelope.Status {
		t.Fatal("expected status: true")
	}

	return envelope.Data
}

func assertEvaluationHTTPError(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	wantStatus int,
) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf(
			"status = %d, want %d; body = %s",
			rec.Code,
			wantStatus,
			rec.Body.String(),
		)
	}

	var envelope struct {
		Status *bool  `json:"status"`
		Error  string `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("expected a JSON error response: %v", err)
	}

	if envelope.Status == nil || *envelope.Status || envelope.Error == "" {
		t.Fatalf("invalid error response: %s", rec.Body.String())
	}
}

func TestEvaluationReportHTTPRejectsInvalidReportCodes(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"missing code", ""},
		{"too few digits", "ER-123"},
		{"non-numeric code", "ER-ABCD"},
		{"path characters", "../ER-0001"},
		{"header characters", "ER-0001\r\nX-Injected: yes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newEvaluationHTTPFixture(t)
			f.reports.report.ReportCode = tt.code

			rec := f.get(
				evaluationHTTPBase+"/"+evaluationHTTPRunID+"/download",
				f.authorization,
			)

			assertEvaluationHTTPError(
				t,
				rec,
				http.StatusInternalServerError,
			)

			if rec.Header().Get("Content-Disposition") != "" {
				t.Fatal("an invalid report code must not produce an attachment")
			}
		})
	}
}
