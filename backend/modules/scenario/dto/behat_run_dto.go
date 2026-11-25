package dto

import "time"

// BehatRunRequest selects one scenario and supplies placeholder values.
type BehatRunRequest struct {
	BaseURL  string            `json:"base_url" binding:"required,url"`
	Scenario string            `json:"scenario"`
	Values   map[string]string `json:"values"`
}

type BehatStepResult struct {
	Index   int    `json:"index"`
	Text    string `json:"text"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type BehatRunResponse struct {
	RunID      string            `json:"run_id"`
	Status     string            `json:"status"`
	DurationMS int64             `json:"duration_ms"`
	Steps      []BehatStepResult `json:"steps"`
	Message    string            `json:"message"`
}

type BehatRunSummary struct {
	RunID      string            `json:"run_id"`
	Status     string            `json:"status"`
	BaseURL    string            `json:"base_url"`
	DurationMS int64             `json:"duration_ms"`
	Steps      []BehatStepResult `json:"steps"`
	Message    string            `json:"message"`
	CreatedAt  time.Time         `json:"created_at"`
}
