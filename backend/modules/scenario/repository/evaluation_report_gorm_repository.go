package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type evaluationReportGORMRepository struct {
	db *gorm.DB
}

func NewEvaluationReportRepository(
	db *gorm.DB,
) EvaluationReportRepository {
	return &evaluationReportGORMRepository{
		db: db,
	}
}

func (r *evaluationReportGORMRepository) DeleteFinished(
	ctx context.Context,
	runID uuid.UUID,
	creatorID string,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var run entities.EvaluationRun
		err := tx.Select("id", "status").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND created_by = ?", runID, creatorID).
			Take(&run).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEvaluationReportNotFound
		}
		if err != nil {
			return err
		}
		if run.Status != entities.EvaluationRunCompleted && run.Status != entities.EvaluationRunFailed {
			return ErrEvaluationReportActive
		}
		return tx.Delete(&run).Error
	})
}

// List reads all evaluation runs without filtering by creator.
func (r *evaluationReportGORMRepository) List(
	ctx context.Context,
	page int,
	pageSize int,
) ([]dto.EvaluationRunSummary, int64, error) {
	// Protect the offset calculation, including direct repository calls.
	if page < 1 || pageSize < 1 ||
		page-1 > int(^uint(0)>>1)/pageSize {
		return nil, 0, fmt.Errorf("invalid evaluation report pagination")
	}

	offset := (page - 1) * pageSize
	rows := []entities.EvaluationRun{}
	var total int64

	err := r.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(&entities.EvaluationRun{}).
				Count(&total).Error; err != nil {
				return err
			}

			return tx.
				Select(
					"id",
					"run_number",
					"dataset_name",
					"created_at",
					"total_cases",
					"status",
					"created_by",
				).
				Order("created_at DESC").
				Order("id DESC").
				Offset(offset).
				Limit(pageSize).
				Find(&rows).Error
		},
		&sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
			ReadOnly:  true,
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("load evaluation runs: %w", err)
	}

	runs := make([]dto.EvaluationRunSummary, 0, len(rows))

	for _, row := range rows {
		summary, err := evaluationRunToSummary(row)
		if err != nil {
			return nil, 0, err
		}

		runs = append(runs, summary)
	}

	return runs, total, nil
}

// GetByID reads one run and its ordered case snapshots.
func (r *evaluationReportGORMRepository) GetByID(
	ctx context.Context,
	runID uuid.UUID,
) (dto.EvaluationReportDetailResponse, error) {
	var empty dto.EvaluationReportDetailResponse
	var run entities.EvaluationRun
	rows := []entities.EvaluationRunCase{}

	err := r.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			if err := tx.
				Where("id = ?", runID).
				Take(&run).Error; err != nil {
				return err
			}

			return tx.
				Where("run_id = ?", runID).
				Order("position ASC").
				Order("id ASC").
				Find(&rows).Error
		},
		&sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
			ReadOnly:  true,
		},
	)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return empty, ErrEvaluationReportNotFound
	}
	if err != nil {
		return empty, fmt.Errorf("load evaluation report: %w", err)
	}

	summary, err := evaluationRunToSummary(run)
	if err != nil {
		return empty, err
	}

	metrics, err := evaluationMetricsToDTO(run.Metrics)
	if err != nil {
		return empty, err
	}

	result := dto.EvaluationReportDetailResponse{
		EvaluationRunSummary: summary,
		ErrorMessage:         run.ErrorMessage,
		Metrics:              metrics,
		Cases: make(
			[]dto.EvaluationReportCase,
			0,
			len(rows),
		),
	}

	for index, row := range rows {
		llm, err := evaluationMethodToDTO(row.LLM)
		if err != nil {
			return empty, fmt.Errorf(
				"load LLM result for case %d: %w",
				index+1,
				err,
			)
		}

		ruleBased, err := evaluationMethodToDTO(row.RuleBased)
		if err != nil {
			return empty, fmt.Errorf(
				"load rule-based result for case %d: %w",
				index+1,
				err,
			)
		}

		result.Cases = append(result.Cases, dto.EvaluationReportCase{
			CaseID:             row.CaseID,
			Title:              row.Title,
			ExpectedGherkin:    row.ExpectedGherkin,
			AcceptanceCriteria: row.AcceptanceCriteria,
			LLM:                llm,
			RuleBased:          ruleBased,
		})
	}

	return result, nil
}

func evaluationRunToSummary(
	run entities.EvaluationRun,
) (dto.EvaluationRunSummary, error) {
	var empty dto.EvaluationRunSummary

	if run.ID == uuid.Nil || run.RunNumber <= 0 || run.TotalCases < 0 {
		return empty, fmt.Errorf("invalid stored evaluation run metadata")
	}

	switch run.Status {
	case entities.EvaluationRunQueued,
		entities.EvaluationRunRunning,
		entities.EvaluationRunCompleted,
		entities.EvaluationRunFailed:
	default:
		return empty, fmt.Errorf("invalid stored evaluation run status")
	}

	return dto.EvaluationRunSummary{
		RunID:       run.ID.String(),
		ReportCode:  fmt.Sprintf("ER-%04d", run.RunNumber),
		DatasetName: run.DatasetName,
		CreatedAt:   run.CreatedAt,
		TotalCases:  run.TotalCases,
		Status:      string(run.Status),
		CreatedBy:   run.CreatedBy,
	}, nil
}

func evaluationMetricsToDTO(
	metrics *entities.EvaluationRunMetrics,
) (*dto.EvaluationReportMetrics, error) {
	if metrics == nil {
		return nil, nil
	}

	values := []*float64{
		metrics.LLM.ValidationPercent,
		metrics.LLM.GUICoveragePercent,
		metrics.LLM.AcceptanceCriteriaPrecisionPercent,
		metrics.LLM.AcceptanceCriteriaRecallPercent,
		metrics.LLM.AcceptanceCriteriaF1Percent,
		metrics.RuleBased.ValidationPercent,
		metrics.RuleBased.GUICoveragePercent,
		metrics.RuleBased.AcceptanceCriteriaPrecisionPercent,
		metrics.RuleBased.AcceptanceCriteriaRecallPercent,
		metrics.RuleBased.AcceptanceCriteriaF1Percent,
	}

	for _, value := range values {
		if !validEvaluationPercent(value) {
			return nil, fmt.Errorf("invalid stored evaluation summary percentage")
		}
	}

	return &dto.EvaluationReportMetrics{
		LLM: dto.EvaluationMethodSummary{
			ValidationPercent:                  metrics.LLM.ValidationPercent,
			GUICoveragePercent:                 metrics.LLM.GUICoveragePercent,
			AcceptanceCriteriaPrecisionPercent: metrics.LLM.AcceptanceCriteriaPrecisionPercent,
			AcceptanceCriteriaRecallPercent:    metrics.LLM.AcceptanceCriteriaRecallPercent,
			AcceptanceCriteriaF1Percent:        metrics.LLM.AcceptanceCriteriaF1Percent,
		},
		RuleBased: dto.EvaluationMethodSummary{
			ValidationPercent:                  metrics.RuleBased.ValidationPercent,
			GUICoveragePercent:                 metrics.RuleBased.GUICoveragePercent,
			AcceptanceCriteriaPrecisionPercent: metrics.RuleBased.AcceptanceCriteriaPrecisionPercent,
			AcceptanceCriteriaRecallPercent:    metrics.RuleBased.AcceptanceCriteriaRecallPercent,
			AcceptanceCriteriaF1Percent:        metrics.RuleBased.AcceptanceCriteriaF1Percent,
		},
	}, nil
}

func evaluationMethodToDTO(
	result *entities.EvaluationCaseMethodResult,
) (*dto.EvaluationMethodResult, error) {
	if result == nil {
		return nil, nil
	}

	switch result.Status {
	case entities.EvaluationCasePass,
		entities.EvaluationCaseWarn,
		entities.EvaluationCaseFail:
	default:
		return nil, fmt.Errorf("invalid stored evaluation case status")
	}

	if !validEvaluationPercent(result.GUICoveragePercent) {
		return nil, fmt.Errorf("invalid stored evaluation case coverage")
	}

	return &dto.EvaluationMethodResult{
		Status:                             string(result.Status),
		GUICoveragePercent:                 result.GUICoveragePercent,
		AcceptanceCriteriaPrecisionPercent: result.AcceptanceCriteriaPrecisionPercent,
		AcceptanceCriteriaRecallPercent:    result.AcceptanceCriteriaRecallPercent,
		AcceptanceCriteriaF1Percent:        result.AcceptanceCriteriaF1Percent,
	}, nil
}

func validEvaluationPercent(value *float64) bool {
	if value == nil {
		return true
	}

	return !math.IsNaN(*value) &&
		!math.IsInf(*value, 0) &&
		*value >= 0 &&
		*value <= 100
}
