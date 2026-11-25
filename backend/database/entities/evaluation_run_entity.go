package entities

import "github.com/google/uuid"

type EvaluationRunStatus string

const (
	EvaluationRunQueued    EvaluationRunStatus = "QUEUED"
	EvaluationRunRunning   EvaluationRunStatus = "RUNNING"
	EvaluationRunCompleted EvaluationRunStatus = "COMPLETED"
	EvaluationRunFailed    EvaluationRunStatus = "FAILED"
)

// EvaluationMethodMetrics stores calculated summary metrics for one method.
// Nil percentages mean unavailable; a pointer to zero means a calculated zero.
type EvaluationMethodMetrics struct {
	ValidationPercent                  *float64 `json:"validation_percent"`
	GUICoveragePercent                 *float64 `json:"gui_coverage_percent"`
	AcceptanceCriteriaPrecisionPercent *float64 `json:"acceptance_criteria_precision_percent"`
	AcceptanceCriteriaRecallPercent    *float64 `json:"acceptance_criteria_recall_percent"`
	AcceptanceCriteriaF1Percent        *float64 `json:"acceptance_criteria_f1_percent"`
}

// EvaluationRunMetrics stores the summaries produced by Run Evaluation.
type EvaluationRunMetrics struct {
	LLM       EvaluationMethodMetrics `json:"llm"`
	RuleBased EvaluationMethodMetrics `json:"rule_based"`
}

// EvaluationRun stores metadata and summary results for one evaluation run.
type EvaluationRun struct {
	// ID is exposed as run_id by the Report API.
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	// Database-generated number used for the human-readable report code.
	RunNumber int64 `gorm:"autoIncrement;not null;uniqueIndex;<-:create"`

	// The source dataset reference is optional so snapshots can survive deletion.
	GoldenDatasetID *uuid.UUID `gorm:"type:uuid;index"`

	// Snapshot of the dataset name when the run was created.
	DatasetName string `gorm:"type:varchar(200);not null"`

	// Matches the TEXT user identifier used by the existing dataset entities.
	// This records the creator; Report access is shared across all QA accounts.
	CreatedBy string `gorm:"type:text;not null;index"`

	// Number of cases planned for this run.
	TotalCases int `gorm:"not null;check:total_cases >= 0"`

	Status EvaluationRunStatus `gorm:"type:varchar(16);not null;default:QUEUED;check:status IN ('QUEUED','RUNNING','COMPLETED','FAILED')"`

	// Store only a public message suitable for display to QA.
	ErrorMessage *string `gorm:"type:text"`

	// SQL NULL means the summary is not available yet.
	Metrics *EvaluationRunMetrics `gorm:"type:jsonb;serializer:json"`

	GoldenDataset *GoldenDataset `gorm:"foreignKey:GoldenDatasetID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Timestamp
}

func (EvaluationRun) TableName() string {
	return "evaluation_runs"
}
