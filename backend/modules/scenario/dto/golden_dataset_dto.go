package dto

import (
	"mime/multipart"
	"time"
)

// GoldenDatasetUploadForm receives the multipart
// form when QA uploads a Golden Dataset.
type GoldenDatasetUploadForm struct {
	Name        string                `form:"name" binding:"required,max=200"`
	Description string                `form:"description"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
}

// GoldenDatasetFile represents the JSON file
// uploaded by QA.
type GoldenDatasetFile struct {
	Items []GoldenDatasetItemInput `json:"items"`
}

// GoldenDatasetItemInput represents a reference
// scenario inside the uploaded JSON file.
type GoldenDatasetItemInput struct {
	ItemTitle        string   `json:"item_title"`
	UserStoryText    string   `json:"user_story_text"`
	GUIElements      []string `json:"gui_elements"`
	FreeScenarioText string   `json:"free_scenario_text"`
	ExpectedGherkin  string   `json:"expected_gherkin"`
}

// GoldenDatasetSummary is returned after upload
// and when listing available datasets.
type GoldenDatasetSummary struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	CreatedBy   string     `json:"created_by"`
	ItemCount   int        `json:"item_count"`
	UsedAt      *time.Time `json:"used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// GoldenDatasetListResponse contains paginated
// datasets and pagination metadata.
type GoldenDatasetListResponse struct {
	Datasets   []GoldenDatasetSummary `json:"datasets"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// GoldenDatasetItemResponse contains the full
// reference scenario stored in PostgreSQL.
type GoldenDatasetItemResponse struct {
	ID               string    `json:"id"`
	ItemTitle        string    `json:"item_title"`
	UserStoryText    string    `json:"user_story_text"`
	GUIElements      []string  `json:"gui_elements"`
	FreeScenarioText string    `json:"free_scenario_text"`
	ExpectedGherkin  string    `json:"expected_gherkin"`
	CreatedAt        time.Time `json:"created_at"`
}

// GoldenDatasetDetailResponse includes dataset
// metadata and all its reference items.
type GoldenDatasetDetailResponse struct {
	GoldenDatasetSummary
	Items []GoldenDatasetItemResponse `json:"items"`
}
