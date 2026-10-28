package entities

import "github.com/google/uuid"

type SourceType string
type ScenarioStatus string

type GherkinScenario struct {
	ID                  uuid.UUID         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	GenerationRequestID *uuid.UUID        `gorm:"type:uuid;index"`
	GenerationRequest   GenerationRequest `gorm:"foreignKey:GenerationRequestID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	FeatureName         string            `gorm:"size:200;not null"`
	GherkinText         string            `gorm:"type:text;not null"`
	SourceType          SourceType        `gorm:"type:varchar(16);not null;default:LLM;check:source_type IN ('LLM','RULE_BASED','CORRECTED')"`
	Status              ScenarioStatus    `gorm:"type:varchar(16);not null;check:status IN ('VALID','INVALID')"`
	ParserMessage       string            `gorm:"type:text"`
	CoveredGuiCount     int               `gorm:"not null;default:0"`
	TotalGuiCount       int               `gorm:"not null;default:0"`
	CoveragePercent     float64           `gorm:"type:numeric(5,2);not null;default:0"`
	Timestamp
}
