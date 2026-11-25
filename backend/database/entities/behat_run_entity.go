package entities

import "github.com/google/uuid"

type BehatRunStatus string

const (
	BehatRunPass  BehatRunStatus = "PASS"
	BehatRunFail  BehatRunStatus = "FAIL"
	BehatRunError BehatRunStatus = "ERROR"
)

// BehatStepResult is the outcome of one mapped Gherkin step.
type BehatStepResult struct {
	Index   int    `json:"index"`
	Text    string `json:"text"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// BehatRun stores one end-to-end execution of a saved scenario.
type BehatRun struct {
	ID         uuid.UUID         `gorm:"type:uuid;primaryKey"`
	ScenarioID uuid.UUID         `gorm:"type:uuid;not null;index"`
	OwnerID    uuid.UUID         `gorm:"type:uuid;not null;index"`
	BaseURL    string            `gorm:"type:text;not null"`
	Status     BehatRunStatus    `gorm:"type:varchar(10);not null"`
	Steps      []BehatStepResult `gorm:"type:jsonb;serializer:json;not null"`
	Message    string            `gorm:"type:text;not null;default:''"`
	DurationMS int64             `gorm:"not null;default:0"`
	Timestamp
}

func (BehatRun) TableName() string { return "behat_runs" }
