package dto

import "time"

// EvaluationRunSummary contains the metadata shown in the run list.
type EvaluationRunSummary struct {
	RunID       string    `json:"run_id"`
	ReportCode  string    `json:"report_code"`
	DatasetName string    `json:"dataset_name"`
	CreatedAt   time.Time `json:"created_at"`
	TotalCases  int       `json:"total_cases"`
	Status      string    `json:"status"`
	CanDelete   bool      `json:"can_delete"`
	CreatedBy   string    `json:"-"`
}

// EvaluationRunListResponse contains one page of evaluation runs.
// Runs must be initialized as an empty slice when no runs are available.
type EvaluationRunListResponse struct {
	Runs       []EvaluationRunSummary `json:"runs"`
	TotalPages int                    `json:"total_pages"`
}

// EvaluationMethodSummary contains the stored summary metrics for one method.
// Nil percentages represent unavailable metrics; zero remains a valid result.
type EvaluationMethodSummary struct {
	ValidationPercent  *float64 `json:"validation_percent"`
	GUICoveragePercent *float64 `json:"gui_coverage_percent"`
}

// EvaluationReportMetrics compares the stored summaries of both methods.
type EvaluationReportMetrics struct {
	LLM       EvaluationMethodSummary `json:"llm"`
	RuleBased EvaluationMethodSummary `json:"rule_based"`
}

// EvaluationMethodResult contains the stored outcome for one evaluated case.
// Status is PASS, WARN, or FAIL when the method result is available.
type EvaluationMethodResult struct {
	Status             string   `json:"status"`
	GUICoveragePercent *float64 `json:"gui_coverage_percent"`
}

// EvaluationReportCase contains the reference snapshot and method results.
// A nil method result means that method has not produced a result yet.
type EvaluationReportCase struct {
	CaseID          string                  `json:"case_id"`
	Title           string                  `json:"title"`
	ExpectedGherkin string                  `json:"expected_gherkin"`
	LLM             *EvaluationMethodResult `json:"llm"`
	RuleBased       *EvaluationMethodResult `json:"rule_based"`
}

// EvaluationReportDetailResponse contains the stored report for one run.
// Cases must be initialized as an empty slice when no results are available.
type EvaluationReportDetailResponse struct {
	EvaluationRunSummary
	ErrorMessage *string                  `json:"error_message"`
	Metrics      *EvaluationReportMetrics `json:"metrics"`
	Cases        []EvaluationReportCase   `json:"cases"`
}
