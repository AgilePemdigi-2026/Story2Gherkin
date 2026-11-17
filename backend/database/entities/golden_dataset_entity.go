
package entities

import (
	"time"

	"github.com/google/uuid"
)

// GoldenDataset represents a reference dataset
// used for evaluating generated Gherkin scenarios.
type GoldenDataset struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

	Name        string  `gorm:"type:varchar(200);not null"`
	Description *string `gorm:"type:text"`

	// Must match users.id, which is TEXT in PostgreSQL.
	CreatedBy string `gorm:"type:text;not null;index"`

	ItemCount int `gorm:"not null;default:0"`

	// Marks the dataset as used in an evaluation.
	UsedAt *time.Time `gorm:"type:timestamp with time zone"`

	// One dataset contains multiple reference items.
	Items []GoldenDatasetItem `gorm:"foreignKey:GoldenDatasetID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`

	Timestamp
}
