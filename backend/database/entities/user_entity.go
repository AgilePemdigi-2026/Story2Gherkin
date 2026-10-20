package entities

import "github.com/google/uuid"

type Role string

const (
	RoleSystemAnalyst Role = "system_analyst"
	RoleQA            Role = "qa"
)

type User struct {
	ID       uuid.UUID `gorm:"primaryKey;default:uuid_generate_v4()"`
	Name     string    `gorm:"size:150;not null"`
	Email    string    `gorm:"size:200;uniqueIndex;not null"`
	Password string    `gorm:"size:100;not null"`
	Role     Role      `gorm:"type:varchar(32);not null;check:role IN ('system_analyst','qa')"`
	Timestamp
}
