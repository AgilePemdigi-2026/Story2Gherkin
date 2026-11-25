package entities

import "github.com/google/uuid"

type EvaluationCaseStatus string

const (
	EvaluationCasePass EvaluationCaseStatus = "PASS"
	EvaluationCaseWarn EvaluationCaseStatus = "WARN"
	EvaluationCaseFail EvaluationCaseStatus = "FAIL"
)

// EvaluationCaseMethodResult stores the result calculated by one method.
// Nil coverage means unavailable; a pointer to zero means calculated zero.
type EvaluationCaseMethodResult struct {
	Status                             EvaluationCaseStatus `json:"status"`
	GUICoveragePercent                 *float64             `json:"gui_coverage_percent"`
	AcceptanceCriteriaPrecisionPercent *float64             `json:"acceptance_criteria_precision_percent"`
	AcceptanceCriteriaRecallPercent    *float64             `json:"acceptance_criteria_recall_percent"`
	AcceptanceCriteriaF1Percent        *float64             `json:"acceptance_criteria_f1_percent"`
	AcceptanceCriteriaMatched          int                  `json:"-"`
	AcceptanceCriteriaTotal            int                  `json:"-"`
	GeneratedBehaviorMatched           int                  `json:"-"`
	GeneratedBehaviorTotal             int                  `json:"-"`
}

// EvaluationRunCase stores one reference snapshot and its evaluation results.
type EvaluationRunCase struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	// A case belongs to exactly one evaluation run.
	RunID uuid.UUID `gorm:"type:uuid;not null;index:idx_evaluation_run_cases_run_case,unique,priority:1;index:idx_evaluation_run_cases_run_position,unique,priority:1"`

	// The case identifier must be unique within its run.
	CaseID string `gorm:"type:text;not null;check:case_id <> '';index:idx_evaluation_run_cases_run_case,unique,priority:2"`

	// One-based position preserving the original evaluation order.
	Position int `gorm:"not null;check:position >= 1;index:idx_evaluation_run_cases_run_position,unique,priority:2"`

	// Snapshots captured when Run Evaluation creates the case.
	Title              string   `gorm:"type:varchar(200);not null"`
	ExpectedGherkin    string   `gorm:"type:text;not null;check:expected_gherkin <> ''"`
	AcceptanceCriteria []string `gorm:"column:acceptance_criteria;type:jsonb;serializer:json"`

	// Processed flips once both generators finished, even when both failed.
	Processed bool `gorm:"not null;default:false"`

	// Each method can remain unavailable independently.
	LLM       *EvaluationCaseMethodResult `gorm:"column:llm;type:jsonb;serializer:json"`
	RuleBased *EvaluationCaseMethodResult `gorm:"column:rule_based;type:jsonb;serializer:json"`

	Run EvaluationRun `gorm:"foreignKey:RunID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Timestamp
}

func (EvaluationRunCase) TableName() string {
	return "evaluation_run_cases"
}
