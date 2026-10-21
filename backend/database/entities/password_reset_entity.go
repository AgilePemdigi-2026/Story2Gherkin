package entities

import (
    "time"
    "github.com/google/uuid"
)

type PasswordResetToken struct {
    ID uuid.UUID `gorm:"type:uuid;primaryKey"`
    UserID uuid.UUID `gorm:"type:uuid;not null;index"`
    User User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
    TokenHash string `gorm:"type:char(64);not null;uniqueIndex"`
    ExpiresAt time.Time `gorm:"type:timestamptz;not null"`
    UsedAt *time.Time `gorm:"type:timestamptz"`
    Timestamp
}