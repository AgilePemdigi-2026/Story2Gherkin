package dto

import "github.com/story2gherkin/database/entities"

const (
	RequestPending entities.RequestStatus = "PENDING"
	RequestSuccess entities.RequestStatus = "SUCCESS"
	RequestFailed  entities.RequestStatus = "FAILED"

	SourceLLM       entities.SourceType = "LLM"
	SourceRuleBased entities.SourceType = "RULE_BASED"
	SourceCorrected entities.SourceType = "CORRECTED"

	ScenarioValid   entities.ScenarioStatus = "VALID"
	ScenarioInvalid entities.ScenarioStatus = "INVALID"
)

func IsValidRequestStatus(s entities.RequestStatus) bool {
	switch s {
	case RequestPending, RequestSuccess, RequestFailed:
		return true
	default:
		return false
	}
}

func IsValidSourceType(s entities.SourceType) bool {
	switch s {
	case SourceLLM, SourceRuleBased, SourceCorrected:
		return true
	default:
		return false
	}
}

func IsValidScenarioStatus(s entities.ScenarioStatus) bool {
	switch s {
	case ScenarioValid, ScenarioInvalid:
		return true
	default:
		return false
	}
}

type GenerateRequest struct {
	UserStory    string   `json:"user_story" binding:"required,min=10,max=8000"`
	GuiElements  []string `json:"gui_elements" binding:"required,min=1,max=30"`
	FreeScenario string   `json:"free_scenario" binding:"required,min=10,max=4000"`
}

type Scenario struct {
	Name          string   `json:"name"`
	Kind          string   `json:"kind"`
	GuiElementIDs []int    `json:"gui_element_ids"`
	Given         []string `json:"given"`
	When          []string `json:"when"`
	Then          []string `json:"then"`
}

type Feature struct {
	FeatureName string     `json:"feature_name"`
	Scenarios   []Scenario `json:"scenarios"`
	Warnings    []string   `json:"warnings"`
}

type GenerateResponse struct {
	Feature    Feature `json:"feature"`
	Gherkin    string  `json:"gherkin"`
	DurationMS int64   `json:"duration_ms"`
}