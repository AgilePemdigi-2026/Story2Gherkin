package dto

import "time"

// Ringkasan satu percobaan Generate.
type HistorySummary struct {
	ID            string    `json:"id"`
	UserStory     string    `json:"user_story"`
	RequestStatus string    `json:"request_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Respons daftar History dengan pagination.
type HistoryListResponse struct {
	History    []HistorySummary `json:"history"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// Hasil Gherkin yang terhubung dengan request.
// Nilainya boleh null jika Generate gagal.
type HistoryScenario struct {
	ID              string    `json:"id"`
	ScenarioCode    string    `json:"scenario_code"`
	FeatureName     string    `json:"feature_name"`
	Gherkin         string    `json:"gherkin"`
	SourceType      string    `json:"source_type"`
	Status          string    `json:"status"`
	CoveragePercent float64   `json:"coverage_percent"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Detail lengkap satu percobaan Generate.
type HistoryDetail struct {
	HistorySummary

	GUIElements  []string         `json:"gui_elements"`
	FreeScenario string           `json:"free_scenario"`
	ErrorMessage string           `json:"error_message"`
	Scenario     *HistoryScenario `json:"scenario"`
}
