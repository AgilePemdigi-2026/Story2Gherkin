package entities

import "github.com/google/uuid"

type Role string

type User struct {
	ID       uuid.UUID `gorm:"primaryKey;default:uuid_generate_v4()"`
	Name     string    `gorm:"size:150;not null"`
	Email    string    `gorm:"size:200;uniqueIndex;not null"`
	Password string    `gorm:"size:100;not null"`
	Role     string    `gorm:"type:enum('system_analyst', 'qa');not null"`
	Timestamp
}
