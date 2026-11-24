package dto

import "time"

// BehatStepInput maps one Gherkin step to a browser action.
type BehatStepInput struct {
	Text     string `json:"text"`
	Action   string `json:"action" binding:"required,oneof=open fill click assert_text assert_visible assert_url"`
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

type BehatRunRequest struct {
	BaseURL string           `json:"base_url" binding:"required,url"`
	Steps   []BehatStepInput `json:"steps" binding:"required,min=1,max=50,dive"`
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
