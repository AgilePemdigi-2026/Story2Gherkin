package entities

import "time"

type PasswordResetToken struct {
	UserID string `gorm:"type:text;primaryKey" json:"-"`

	TokenHash string `gorm:"type:char(64);not null;uniqueIndex" json:"-"`

	ExpiresAt time.Time `gorm:"type:timestamptz;not null;index" json:"-"`

	CreatedAt time.Time `gorm:"type:timestamptz;not null" json:"-"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
