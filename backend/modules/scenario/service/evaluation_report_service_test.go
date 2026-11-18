package service_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
	"gorm.io/gorm"
)

const (
	evaluationReportTestUserID = "469075dd-2376-4b21-aebc-a842f571173a"
	evaluationReportTestRunID  = "6b308226-0b1f-4efb-92c6-9b4066132b74"
)

func TestEvaluationReportServiceRejectsUnauthorizedAccess(t *testing.T) {
	lookupError := errors.New("user storage unavailable")

	tests := []struct {
		name      string
		userID    string
		role      entities.Role
		lookupErr error
		wantErr   error
	}{
		{
			"system analyst",
			evaluationReportTestUserID,
			entities.RoleSystemAnalyst,
			nil,
			service.ErrEvaluationReportForbidden,
		},
		{
			"unknown role",
			evaluationReportTestUserID,
			entities.Role("admin"),
			nil,
			service.ErrEvaluationReportForbidden,
		},
		{
			"missing user",
			evaluationReportTestUserID,
			entities.RoleQA,
			gorm.ErrRecordNotFound,
			service.ErrEvaluationReportForbidden,
		},
		{
			"invalid user ID",
			"invalid-user-id",
			entities.RoleQA,
			nil,
			service.ErrEvaluationReportForbidden,
		},
		{
			"user lookup failure",
			evaluationReportTestUserID,
			entities.RoleQA,
			lookupError,
			lookupError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, action := range evaluationReportTestActions() {
				t.Run(action.name, func(t *testing.T) {
					svc, users, reports := newEvaluationReportTestService()
					users.user.Role = tt.role
					users.err = tt.lookupErr

					err := action.call(svc, tt.userID)

					if !errors.Is(err, tt.wantErr) {
						t.Fatalf("error = %v, want %v", err, tt.wantErr)
					}

					if reports.listCalls != 0 || reports.detailCalls != 0 {
						t.Fatal("report storage must not be accessed without authorization")
					}
				})
			}
		})
	}
}

func TestEvaluationReportServiceAllowsEveryQAAccount(t *testing.T) {
	userIDs := []string{
		evaluationReportTestUserID,
		"dd45f5df-aa4d-401c-bce6-3ba63280cf01",
	}

	for _, userID := range userIDs {
		t.Run(userID, func(t *testing.T) {
			svc, users, reports := newEvaluationReportTestService()
			users.user.ID = uuid.MustParse(userID)

			reports.report = dto.EvaluationReportDetailResponse{
				EvaluationRunSummary: dto.EvaluationRunSummary{
					RunID:       evaluationReportTestRunID,
					DatasetName: "Shared evaluation",
					Status:      "COMPLETED",
				},
			}
			reports.runs = []dto.EvaluationRunSummary{
				reports.report.EvaluationRunSummary,
			}
			reports.total = 1

			for _, action := range evaluationReportTestActions() {
				if err := action.call(svc, userID); err != nil {
					t.Fatalf("%s failed for QA account: %v", action.name, err)
				}
			}

			if users.lastUserID != userID {
				t.Fatalf("authorized user = %q, want %q", users.lastUserID, userID)
			}

			if reports.lastRunID.String() != evaluationReportTestRunID {
				t.Fatal("QA accounts must be able to read the same shared run")
			}
		})
	}
}

func TestEvaluationReportServiceListPagination(t *testing.T) {
	sample := []dto.EvaluationRunSummary{
		{
			RunID:       evaluationReportTestRunID,
			DatasetName: "Stored dataset",
			Status:      "COMPLETED",
		},
	}

	tests := []struct {
		name      string
		page      int
		pageSize  int
		total     int64
		runs      []dto.EvaluationRunSummary
		wantPages int
		wantErr   bool
	}{
		{"empty list", 1, 10, 0, nil, 0, false},
		{"partial last page", 3, 10, 21, sample, 3, false},
		{"empty page beyond results", 4, 10, 21, nil, 3, false},
		{"maximum page size", 1, 100, 101, sample, 2, false},
		{"zero page", 0, 10, 0, nil, 0, true},
		{"negative page", -1, 10, 0, nil, 0, true},
		{"page above limit", service.EvaluationReportMaxPage + 1, 10, 0, nil, 0, true},
		{"zero page size", 1, 0, 0, nil, 0, true},
		{"negative page size", 1, -1, 0, nil, 0, true},
		{"page size above limit", 1, service.EvaluationReportMaxPageSize + 1, 0, nil, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, reports := newEvaluationReportTestService()
			reports.runs = tt.runs
			reports.total = tt.total

			got, err := svc.List(
				context.Background(),
				evaluationReportTestUserID,
				tt.page,
				tt.pageSize,
			)

			if tt.wantErr {
				if !errors.Is(err, service.ErrEvaluationReportInvalidRequest) {
					t.Fatalf("error = %v, want ErrEvaluationReportInvalidRequest", err)
				}
				if reports.listCalls != 0 {
					t.Fatal("invalid pagination must not reach report storage")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if got.Runs == nil {
				t.Fatal("runs must be an array, not null")
			}

			wantRuns := tt.runs
			if wantRuns == nil {
				wantRuns = []dto.EvaluationRunSummary{}
			}

			if !reflect.DeepEqual(got.Runs, wantRuns) {
				t.Fatalf("runs = %#v, want %#v", got.Runs, wantRuns)
			}

			if got.TotalPages != tt.wantPages {
				t.Fatalf("total_pages = %d, want %d", got.TotalPages, tt.wantPages)
			}

			if reports.lastPage != tt.page || reports.lastPageSize != tt.pageSize {
				t.Fatal("pagination parameters were not forwarded correctly")
			}
		})
	}
}

func TestEvaluationReportServiceDetailPreservesStoredResults(t *testing.T) {
	for _, status := range []string{"QUEUED", "RUNNING", "COMPLETED", "FAILED"} {
		t.Run(status, func(t *testing.T) {
			svc, _, reports := newEvaluationReportTestService()

			stored := dto.EvaluationReportDetailResponse{
				EvaluationRunSummary: dto.EvaluationRunSummary{
					RunID:       evaluationReportTestRunID,
					DatasetName: "Dataset name captured at run creation",
					TotalCases:  1,
					Status:      status,
				},
			}

			if status == "COMPLETED" {
				zero := 0.0

				stored.Metrics = &dto.EvaluationReportMetrics{
					LLM: dto.EvaluationMethodSummary{
						ValidationPercent: &zero,
					},
				}

				stored.Cases = []dto.EvaluationReportCase{
					{
						CaseID:          "GC-001",
						Title:           "Stored reference case",
						ExpectedGherkin: "Feature: Original snapshot\n  Scenario: Original case",
						LLM: &dto.EvaluationMethodResult{
							Status:             "FAIL",
							GUICoveragePercent: &zero,
						},
					},
				}
			}

			if status == "FAILED" {
				message := "Evaluation could not be completed."
				stored.ErrorMessage = &message
			}

			reports.report = stored

			got, err := svc.Detail(
				context.Background(),
				evaluationReportTestUserID,
				evaluationReportTestRunID,
			)
			if err != nil {
				t.Fatal(err)
			}

			want := stored
			if want.Cases == nil {
				want.Cases = []dto.EvaluationReportCase{}
			}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("report = %#v, want %#v", got, want)
			}

			if reports.detailCalls != 1 {
				t.Fatalf("detail calls = %d, want 1", reports.detailCalls)
			}

			if reports.lastRunID.String() != evaluationReportTestRunID {
				t.Fatal("the requested run ID was not forwarded correctly")
			}
		})
	}
}

func TestEvaluationReportServiceDetailErrors(t *testing.T) {
	storageError := errors.New("report storage unavailable")

	tests := []struct {
		name      string
		runID     string
		repoErr   error
		wantErr   error
		wantCalls int
	}{
		{
			"invalid run ID",
			"invalid-run-id",
			nil,
			service.ErrEvaluationReportInvalidRequest,
			0,
		},
		{
			"missing run",
			evaluationReportTestRunID,
			fmt.Errorf("lookup failed: %w", repository.ErrEvaluationReportNotFound),
			service.ErrEvaluationReportNotFound,
			1,
		},
		{
			"storage failure",
			evaluationReportTestRunID,
			storageError,
			storageError,
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _, reports := newEvaluationReportTestService()
			reports.err = tt.repoErr

			_, err := svc.Detail(
				context.Background(),
				evaluationReportTestUserID,
				tt.runID,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}

			if reports.detailCalls != tt.wantCalls {
				t.Fatalf("detail calls = %d, want %d", reports.detailCalls, tt.wantCalls)
			}
		})
	}
}

func TestEvaluationReportServiceListStorageError(t *testing.T) {
	svc, _, reports := newEvaluationReportTestService()
	storageError := errors.New("report storage unavailable")
	reports.err = storageError

	_, err := svc.List(
		context.Background(),
		evaluationReportTestUserID,
		1,
		10,
	)

	if !errors.Is(err, storageError) {
		t.Fatalf("error = %v, want storage error", err)
	}
}

func TestEvaluationReportServiceDownloadRequiresCompletedRun(t *testing.T) {
	for _, status := range []string{"QUEUED", "RUNNING", "FAILED", "COMPLETED"} {
		t.Run(status, func(t *testing.T) {
			svc, _, reports := newEvaluationReportTestService()

			reports.report = dto.EvaluationReportDetailResponse{
				EvaluationRunSummary: dto.EvaluationRunSummary{
					RunID:      evaluationReportTestRunID,
					Status:     status,
					TotalCases: 1,
				},
				Cases: []dto.EvaluationReportCase{
					{
						CaseID:          "GC-001",
						ExpectedGherkin: "Feature: Stored reference",
						LLM: &dto.EvaluationMethodResult{
							Status: "FAIL",
						},
					},
				},
			}

			got, err := svc.Download(
				context.Background(),
				evaluationReportTestUserID,
				evaluationReportTestRunID,
			)

			if status != "COMPLETED" {
				if !errors.Is(err, service.ErrEvaluationReportNotReady) {
					t.Fatalf("error = %v, want ErrEvaluationReportNotReady", err)
				}
				if got.RunID != "" || len(got.Cases) != 0 {
					t.Fatal("an unfinished download must not return report data")
				}
				return
			}

			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, reports.report) {
				t.Fatal("completed reports must retain cases with FAIL results")
			}
		})
	}
}

// Test doubles below do not connect to a database.

type evaluationReportUsersStub struct {
	user       entities.User
	err        error
	lastUserID string
}

func (s *evaluationReportUsersStub) GetUserById(
	_ context.Context,
	_ *gorm.DB,
	userID string,
) (entities.User, error) {
	s.lastUserID = userID
	return s.user, s.err
}

type evaluationReportRepositoryStub struct {
	runs         []dto.EvaluationRunSummary
	total        int64
	report       dto.EvaluationReportDetailResponse
	err          error
	listCalls    int
	detailCalls  int
	lastPage     int
	lastPageSize int
	lastRunID    uuid.UUID
}

func (s *evaluationReportRepositoryStub) List(
	_ context.Context,
	page int,
	pageSize int,
) ([]dto.EvaluationRunSummary, int64, error) {
	s.listCalls++
	s.lastPage = page
	s.lastPageSize = pageSize
	return s.runs, s.total, s.err
}

func (s *evaluationReportRepositoryStub) GetByID(
	_ context.Context,
	runID uuid.UUID,
) (dto.EvaluationReportDetailResponse, error) {
	s.detailCalls++
	s.lastRunID = runID
	return s.report, s.err
}

func newEvaluationReportTestService() (
	service.EvaluationReportService,
	*evaluationReportUsersStub,
	*evaluationReportRepositoryStub,
) {
	users := &evaluationReportUsersStub{
		user: entities.User{
			ID:   uuid.MustParse(evaluationReportTestUserID),
			Role: entities.RoleQA,
		},
	}
	reports := &evaluationReportRepositoryStub{}

	svc := service.NewEvaluationReportService(users, reports)
	return svc, users, reports
}

type evaluationReportTestAction struct {
	name string
	call func(service.EvaluationReportService, string) error
}

func evaluationReportTestActions() []evaluationReportTestAction {
	return []evaluationReportTestAction{
		{
			name: "list",
			call: func(s service.EvaluationReportService, userID string) error {
				_, err := s.List(context.Background(), userID, 1, 10)
				return err
			},
		},
		{
			name: "detail",
			call: func(s service.EvaluationReportService, userID string) error {
				_, err := s.Detail(
					context.Background(), userID, evaluationReportTestRunID,
				)
				return err
			},
		},
		{
			name: "download",
			call: func(s service.EvaluationReportService, userID string) error {
				_, err := s.Download(
					context.Background(), userID, evaluationReportTestRunID,
				)
				return err
			},
		},
	}
}
