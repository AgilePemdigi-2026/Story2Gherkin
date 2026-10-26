package repository

import (
	"context"

	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

type GenerationRequestRepository interface {
	Create(ctx context.Context, tx *gorm.DB, req entities.GenerationRequest) (entities.GenerationRequest, error)
	Update(ctx context.Context, tx *gorm.DB, req entities.GenerationRequest) (entities.GenerationRequest, error)
}

type generationRequestRepository struct {
	db *gorm.DB
}

func NewGenerationRequestRepository(db *gorm.DB) GenerationRequestRepository {
	return &generationRequestRepository{
		db: db,
	}
}

func (r *generationRequestRepository) Create(ctx context.Context, tx *gorm.DB, req entities.GenerationRequest) (entities.GenerationRequest, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Create(&req).Error; err != nil {
		return entities.GenerationRequest{}, err
	}

	return req, nil
}

func (r *generationRequestRepository) Update(ctx context.Context, tx *gorm.DB, req entities.GenerationRequest) (entities.GenerationRequest, error) {
	if tx == nil {
		tx = r.db
	}

	if err := tx.WithContext(ctx).Updates(&req).Error; err != nil {
		return entities.GenerationRequest{}, err
	}

	return req, nil
}
