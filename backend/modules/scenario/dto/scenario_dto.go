package dto

type GenerateRequest struct {
	UserStory          string   `json:"user_story"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	MaxScenarios       int      `json:"max_scenarios"`
	Language           string   `json:"language"`
}

type Scenario struct {
	Name         string   `json:"name"`
	Kind         string   `json:"kind"`
	CriterionIDs []int    `json:"criterion_ids"`
	Given        []string `json:"given"`
	When         []string `json:"when"`
	Then         []string `json:"then"`
}

type Feature struct {
	FeatureName string     `json:"feature_name"`
	Scenarios   []Scenario `json:"scenarios"`
	Warnings    []string   `json:"warnings"`
}

type GenerateResponse struct {
	Feature       Feature `json:"feature"`
	Gherkin       string  `json:"gherkin"`
	Model         string  `json:"model"`
	PromptVersion string  `json:"prompt_version"`
	DurationMS    int64   `json:"duration_ms"`
}
