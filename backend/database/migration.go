package database

import (
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entities.User{},
	)

	if err != nil {
		return err
	}

	return nil
}
