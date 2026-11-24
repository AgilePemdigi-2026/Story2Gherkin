package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

type BehatRunRepository interface {
	Create(ctx context.Context, run *entities.BehatRun) error
	Update(ctx context.Context, run *entities.BehatRun) error
	ListByScenario(ctx context.Context, owner, scenarioID uuid.UUID, limit int) ([]entities.BehatRun, error)
}

type behatRunRepository struct {
	db *gorm.DB
}

func NewBehatRunRepository(db *gorm.DB) BehatRunRepository {
	return &behatRunRepository{db: db}
}

func (r *behatRunRepository) Create(ctx context.Context, run *entities.BehatRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *behatRunRepository) Update(ctx context.Context, run *entities.BehatRun) error {
	return r.db.WithContext(ctx).Model(run).
		Select("status", "steps", "message", "duration_ms", "updated_at").
		Updates(run).Error
}

func (r *behatRunRepository) ListByScenario(
	ctx context.Context,
	owner, scenarioID uuid.UUID,
	limit int,
) ([]entities.BehatRun, error) {
	rows := []entities.BehatRun{}
	err := r.db.WithContext(ctx).
		Where("owner_id = ? AND scenario_id = ?", owner, scenarioID).
		Order("created_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
