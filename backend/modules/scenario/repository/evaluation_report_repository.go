package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story2gherkin/modules/scenario/dto"
)

// ErrEvaluationReportNotFound indicates that the requested run does not exist.
var ErrEvaluationReportNotFound = errors.New("evaluation report not found")
var ErrEvaluationReportActive = errors.New("evaluation is still running")

// EvaluationReportRepository manages stored evaluation results.
// It must not execute evaluations or recalculate metrics.
type EvaluationReportRepository interface {
	// List returns one page of runs and the total number of runs.
	//
	// Runs are ordered by created_at DESC, then run_id DESC.
	// An empty page returns an empty slice without a not-found error.
	// All runs are visible to authorized QA users, regardless of ownership.
	List(
		ctx context.Context,
		page int,
		pageSize int,
	) ([]dto.EvaluationRunSummary, int64, error)

	// GetByID returns one stored report, including available case results.
	//
	// Dataset names, case titles, and expected Gherkin must come from
	// snapshots captured when the evaluation run was created.
	// Unavailable metrics and method results remain nil.
	// Cases retain their evaluation order.
	//
	// Returns ErrEvaluationReportNotFound when the run does not exist.
	GetByID(
		ctx context.Context,
		runID uuid.UUID,
	) (dto.EvaluationReportDetailResponse, error)

	DeleteFinished(ctx context.Context, runID uuid.UUID, creatorID string) error
}
