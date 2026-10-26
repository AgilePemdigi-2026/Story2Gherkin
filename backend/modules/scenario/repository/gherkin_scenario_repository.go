package repository

import (
	"context"

	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

type GherkinScenarioRepository interface {
	Create(ctx context.Context, tx *gorm.DB, scenario entities.GherkinScenario) (entities.GherkinScenario, error)
}

type gherkinScenarioRepository struct {
	db *gorm.DB
}

func NewGherkinScenarioRepository(db *gorm.DB) GherkinScenarioRepository {
	return &gherkinScenarioRepository{
		db: db,
	}
}

func (r *gherkinScenarioRepository) Create(ctx context.Context, tx *gorm.DB, scenario entities.GherkinScenario) (entities.GherkinScenario, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&scenario).Error; err != nil {
		return entities.GherkinScenario{}, err
	}

	return scenario, nil
}
