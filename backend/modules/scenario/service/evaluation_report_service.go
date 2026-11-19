package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/gorm"
)

const (
	EvaluationReportDefaultPageSize = 10
	EvaluationReportMaxPageSize     = 100
	EvaluationReportMaxPage         = 1000000
)

var (
	ErrEvaluationReportForbidden = errors.New(
		"only Quality Assurance users may access evaluation reports",
	)
	ErrEvaluationReportInvalidRequest = errors.New(
		"invalid evaluation report ID or pagination parameters",
	)
	ErrEvaluationReportNotFound = errors.New(
		"evaluation report not found",
	)
	ErrEvaluationReportNotReady = errors.New(
		"only completed evaluation reports can be downloaded",
	)
	ErrEvaluationReportActive = errors.New("evaluation is still running")
)

type EvaluationReportService interface {
	List(
		ctx context.Context,
		userID string,
		page int,
		pageSize int,
	) (dto.EvaluationRunListResponse, error)

	Detail(
		ctx context.Context,
		userID string,
		runID string,
	) (dto.EvaluationReportDetailResponse, error)

	Download(
		ctx context.Context,
		userID string,
		runID string,
	) (dto.EvaluationReportDetailResponse, error)

	Delete(ctx context.Context, userID string, runID string) error
}

type evaluationReportService struct {
	users   ReviewUserRepository
	reports repository.EvaluationReportRepository
}

func NewEvaluationReportService(
	users ReviewUserRepository,
	reports repository.EvaluationReportRepository,
) EvaluationReportService {
	return &evaluationReportService{
		users:   users,
		reports: reports,
	}
}

// authorizeQA checks the current role stored in the database.
func (s *evaluationReportService) authorizeQA(
	ctx context.Context,
	userID string,
) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return ErrEvaluationReportForbidden
	}

	user, err := s.users.GetUserById(ctx, nil, id.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrEvaluationReportForbidden
	}
	if err != nil {
		return err
	}

	if user.Role != entities.RoleQA {
		return ErrEvaluationReportForbidden
	}

	return nil
}

// List returns evaluation runs visible to all authorized QA users.
func (s *evaluationReportService) List(
	ctx context.Context,
	userID string,
	page int,
	pageSize int,
) (dto.EvaluationRunListResponse, error) {
	result := dto.EvaluationRunListResponse{
		Runs: []dto.EvaluationRunSummary{},
	}

	if err := s.authorizeQA(ctx, userID); err != nil {
		return result, err
	}

	if page < 1 || page > EvaluationReportMaxPage ||
		pageSize < 1 || pageSize > EvaluationReportMaxPageSize {
		return result, ErrEvaluationReportInvalidRequest
	}

	runs, total, err := s.reports.List(ctx, page, pageSize)
	if err != nil {
		return result, err
	}

	if runs != nil {
		result.Runs = runs
	}
	creatorID := uuid.MustParse(userID).String()
	for i := range result.Runs {
		setReportDeletePermission(&result.Runs[i], creatorID)
	}

	result.TotalPages = int(
		(total + int64(pageSize) - 1) / int64(pageSize),
	)

	return result, nil
}

// Detail reads a stored report without executing an evaluation.
func (s *evaluationReportService) Detail(
	ctx context.Context,
	userID string,
	runID string,
) (dto.EvaluationReportDetailResponse, error) {
	var empty dto.EvaluationReportDetailResponse

	if err := s.authorizeQA(ctx, userID); err != nil {
		return empty, err
	}

	id, err := uuid.Parse(runID)
	if err != nil {
		return empty, ErrEvaluationReportInvalidRequest
	}

	report, err := s.reports.GetByID(ctx, id)
	if errors.Is(err, repository.ErrEvaluationReportNotFound) {
		return empty, ErrEvaluationReportNotFound
	}
	if err != nil {
		return empty, err
	}

	// Return [] rather than null when case results are not available.
	if report.Cases == nil {
		report.Cases = []dto.EvaluationReportCase{}
	}
	setReportDeletePermission(&report.EvaluationRunSummary, uuid.MustParse(userID).String())

	return report, nil
}

func setReportDeletePermission(report *dto.EvaluationRunSummary, creatorID string) {
	report.CanDelete = report.CreatedBy == creatorID &&
		(report.Status == string(entities.EvaluationRunCompleted) || report.Status == string(entities.EvaluationRunFailed))
}

func (s *evaluationReportService) Delete(ctx context.Context, userID, runID string) error {
	if err := s.authorizeQA(ctx, userID); err != nil {
		return err
	}
	id, err := uuid.Parse(runID)
	if err != nil {
		return ErrEvaluationReportInvalidRequest
	}
	err = s.reports.DeleteFinished(ctx, id, uuid.MustParse(userID).String())
	switch {
	case errors.Is(err, repository.ErrEvaluationReportNotFound):
		return ErrEvaluationReportNotFound
	case errors.Is(err, repository.ErrEvaluationReportActive):
		return ErrEvaluationReportActive
	default:
		return err
	}
}

// Download returns a completed report for CSV generation.
func (s *evaluationReportService) Download(
	ctx context.Context,
	userID string,
	runID string,
) (dto.EvaluationReportDetailResponse, error) {
	report, err := s.Detail(ctx, userID, runID)
	if err != nil {
		return dto.EvaluationReportDetailResponse{}, err
	}

	if report.Status != "COMPLETED" {
		return dto.EvaluationReportDetailResponse{},
			ErrEvaluationReportNotReady
	}

	return report, nil
}
