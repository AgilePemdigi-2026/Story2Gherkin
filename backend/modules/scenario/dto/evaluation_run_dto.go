package dto

import "github.com/story2gherkin/database/entities"

type EvaluationRunCreateRequest struct {
	GoldenDatasetID string `json:"golden_dataset_id" binding:"required,uuid"`
}

type EvaluationRunCreated struct {
	RunID      string                       `json:"run_id"`
	ReportCode string                       `json:"report_code"`
	Status     entities.EvaluationRunStatus `json:"status"`
	TotalCases int                          `json:"total_cases"`
}

type EvaluationRunProgress struct {
	RunID          string                       `json:"run_id"`
	Status         entities.EvaluationRunStatus `json:"status"`
	TotalCases     int                          `json:"total_cases"`
	ProcessedCases int                          `json:"processed_cases"`
	ErrorMessage   *string                      `json:"error_message"`
}
