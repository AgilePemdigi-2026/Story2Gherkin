package database

import (
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&entities.User{},
		&entities.RefreshToken{},
		&entities.GenerationRequest{},
		&entities.GherkinScenario{},
		&entities.GoldenDataset{},
		&entities.GoldenDatasetItem{},
	)

	if err != nil {
		return err
	}

	return nil
}
