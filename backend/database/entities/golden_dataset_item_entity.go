
package entities

import "github.com/google/uuid"

// GoldenDatasetItem stores one reference test case
// inside a Golden Dataset.
type GoldenDatasetItem struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	GoldenDatasetID uuid.UUID `gorm:"type:uuid;not null;index"`

	ItemTitle string `gorm:"type:varchar(200);not null"`

	UserStoryText string `gorm:"type:text;not null"`

	// Store GUI elements as a JSON array in PostgreSQL.
	GUIElements []string `gorm:"column:gui_elements;type:jsonb;serializer:json;not null"`

	FreeScenarioText string `gorm:"type:text;not null"`

	ExpectedGherkin string `gorm:"type:text;not null"`

	Timestamp
}
