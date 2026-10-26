package entities

import "github.com/google/uuid"

type RequestStatus string

type GenerationRequest struct {
	ID           uuid.UUID     `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       uuid.UUID     `gorm:"type:uuid;not null;index"`
	User         User          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	UserStory    string        `gorm:"type:text;not null"`
	GuiElements  []string      `gorm:"type:text;serializer:json;not null"`
	FreeScenario string        `gorm:"type:text;not null"`
	Status       RequestStatus `gorm:"type:varchar(16);not null;default:PENDING;check:status IN ('PENDING','SUCCESS','FAILED')"`
	ErrorMessage string        `gorm:"type:text"`
	Timestamp
}
