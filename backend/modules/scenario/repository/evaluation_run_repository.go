package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"gorm.io/gorm"
)

// ErrEvaluationRunNotFound indicates that the requested run does not exist.
var ErrEvaluationRunNotFound = errors.New("evaluation run not found")

// EvaluationRunRepository owns the write side of evaluation runs.
type EvaluationRunRepository interface {
	// HasActive reports whether a run is queued or running.
	HasActive(ctx context.Context) (bool, error)

	// CreateWithCases stores a run, its case snapshots, and locks the dataset.
	CreateWithCases(
		ctx context.Context,
		run *entities.EvaluationRun,
		cases []entities.EvaluationRunCase,
	) error

	MarkRunning(ctx context.Context, runID uuid.UUID) error

	SaveCaseResult(
		ctx context.Context,
		caseID uuid.UUID,
		llm, ruleBased *entities.EvaluationCaseMethodResult,
	) error

	// GetProgress returns the run status and processed case count.
	GetProgress(
		ctx context.Context,
		runID uuid.UUID,
	) (dto.EvaluationRunProgress, error)

	Finish(
		ctx context.Context,
		runID uuid.UUID,
		status entities.EvaluationRunStatus,
		metrics *entities.EvaluationRunMetrics,
		message *string,
	) error
}

type evaluationRunRepository struct {
	db *gorm.DB
}

func NewEvaluationRunRepository(db *gorm.DB) EvaluationRunRepository {
	return &evaluationRunRepository{db: db}
}

func (r *evaluationRunRepository) HasActive(ctx context.Context) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.EvaluationRun{}).
		Where("status IN ?", []entities.EvaluationRunStatus{
			entities.EvaluationRunQueued,
			entities.EvaluationRunRunning,
		}).
		Count(&count).
		Error

	return count > 0, err
}

func (r *evaluationRunRepository) CreateWithCases(
	ctx context.Context,
	run *entities.EvaluationRun,
	cases []entities.EvaluationRunCase,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(run).Error; err != nil {
			return err
		}

		for index := range cases {
			cases[index].RunID = run.ID
		}

		if err := tx.Create(&cases).Error; err != nil {
			return err
		}

		if run.GoldenDatasetID == nil {
			return nil
		}

		return tx.Model(&entities.GoldenDataset{}).
			Where("id = ? AND used_at IS NULL", *run.GoldenDatasetID).
			Update("used_at", time.Now()).
			Error
	})
}

func (r *evaluationRunRepository) MarkRunning(ctx context.Context, runID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&entities.EvaluationRun{}).
		Where("id = ?", runID).
		Update("status", entities.EvaluationRunRunning).
		Error
}

func (r *evaluationRunRepository) SaveCaseResult(
	ctx context.Context,
	caseID uuid.UUID,
	llm, ruleBased *entities.EvaluationCaseMethodResult,
) error {
	return r.db.WithContext(ctx).
		Model(&entities.EvaluationRunCase{}).
		Where("id = ?", caseID).
		// Struct updates apply serializer:json before pgx encodes parameters.
		// Select also writes nil results when a method is unavailable.
		Select("LLM", "RuleBased", "Processed").
		Updates(entities.EvaluationRunCase{
			LLM:       llm,
			RuleBased: ruleBased,
			Processed: true,
		}).
		Error
}

func (r *evaluationRunRepository) GetProgress(
	ctx context.Context,
	runID uuid.UUID,
) (dto.EvaluationRunProgress, error) {
	var row struct {
		Status       entities.EvaluationRunStatus
		TotalCases   int
		Processed    int
		ErrorMessage *string
	}

	err := r.db.WithContext(ctx).
		Model(&entities.EvaluationRun{}).
		Select(
			"evaluation_runs.status, evaluation_runs.total_cases, evaluation_runs.error_message, "+
				"(SELECT count(*) FROM evaluation_run_cases WHERE run_id = evaluation_runs.id AND processed) AS processed",
		).
		Where("evaluation_runs.id = ?", runID).
		Take(&row).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.EvaluationRunProgress{}, ErrEvaluationRunNotFound
	}
	if err != nil {
		return dto.EvaluationRunProgress{}, err
	}

	return dto.EvaluationRunProgress{
		RunID:          runID.String(),
		Status:         row.Status,
		TotalCases:     row.TotalCases,
		ProcessedCases: row.Processed,
		ErrorMessage:   row.ErrorMessage,
	}, nil
}

func (r *evaluationRunRepository) Finish(
	ctx context.Context,
	runID uuid.UUID,
	status entities.EvaluationRunStatus,
	metrics *entities.EvaluationRunMetrics,
	message *string,
) error {
	return r.db.WithContext(ctx).
		Model(&entities.EvaluationRun{}).
		Where("id = ?", runID).
		Select("Status", "Metrics", "ErrorMessage").
		Updates(entities.EvaluationRun{
			Status:       status,
			Metrics:      metrics,
			ErrorMessage: message,
		}).
		Error
}
