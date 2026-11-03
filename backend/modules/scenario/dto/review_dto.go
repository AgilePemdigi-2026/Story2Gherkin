package dto

import "time"

type ValidationIssue struct {
	Message string `json:"message"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
}

type StructureChecks struct {
	Feature  bool `json:"feature"`
	Scenario bool `json:"scenario"`
	Given    bool `json:"given"`
	When     bool `json:"when"`
	Then     bool `json:"then"`
}

type ScenarioValidation struct {
	Valid              bool              `json:"valid"`
	FeatureName        string            `json:"feature_name"`
	Checks             StructureChecks   `json:"checks"`
	Issues             []ValidationIssue `json:"issues"`
	UsedElements       []string          `json:"used_elements"`
	NotCoveredElements []string          `json:"not_covered_elements"`
	CoveragePercent    float64           `json:"coverage_percent"`
}

type ScenarioSummary struct {
	ID          string    `json:"id"`
	FeatureName string    `json:"feature_name"`
	SourceType  string    `json:"source_type"`
	Status      string    `json:"status"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ReviewScenario struct {
	ScenarioSummary
	Gherkin     string             `json:"gherkin"`
	GUIElements []string           `json:"gui_elements"`
	Validation  ScenarioValidation `json:"validation"`
}

type SaveCorrectionRequest struct {
	Gherkin   string    `json:"gherkin" binding:"max=65536"`
	UpdatedAt time.Time `json:"updated_at" binding:"required"`
}
