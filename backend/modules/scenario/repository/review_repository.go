package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

var ErrCorrectionConflict = errors.New("scenario changed since it was loaded")

type ReviewRepository interface {
	ListOwned(context.Context, uuid.UUID) ([]entities.GherkinScenario, error)
	GetOwned(context.Context, uuid.UUID, uuid.UUID) (entities.GherkinScenario, error)
	SaveCorrectionOwned(context.Context, uuid.UUID, entities.GherkinScenario, time.Time) (entities.GherkinScenario, error)
}

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &gherkinScenarioRepository{db: db}
}

// All access is scoped to the generation request owner, including updates.
func (r *gherkinScenarioRepository) owned(ctx context.Context, owner uuid.UUID) *gorm.DB {
	return r.db.WithContext(ctx).Model(&entities.GherkinScenario{}).
		Where("generation_request_id IN (?)", r.db.Model(&entities.GenerationRequest{}).
			Select("id").Where("user_id = ?", owner))
}

func (r *gherkinScenarioRepository) ListOwned(ctx context.Context, owner uuid.UUID) ([]entities.GherkinScenario, error) {
	rows := []entities.GherkinScenario{}
	err := r.owned(ctx, owner).Select("id", "scenario_number", "feature_name", "source_type", "status", "updated_at").
		Order("created_at DESC, id DESC").Find(&rows).Error
	return rows, err
}

func (r *gherkinScenarioRepository) GetOwned(ctx context.Context, owner, id uuid.UUID) (entities.GherkinScenario, error) {
	var row entities.GherkinScenario
	err := r.owned(ctx, owner).Preload("GenerationRequest").Where("id = ?", id).Take(&row).Error
	return row, err
}

func (r *gherkinScenarioRepository) SaveCorrectionOwned(ctx context.Context, owner uuid.UUID, row entities.GherkinScenario, expected time.Time) (entities.GherkinScenario, error) {
	// An atomic conditional update prevents overwriting another tab's correction.
	updatedAt := time.Now().UTC().Truncate(time.Microsecond)
	result := r.owned(ctx, owner).Where("id = ? AND updated_at = ?", row.ID, expected).
		Updates(map[string]any{
			"feature_name": row.FeatureName, "gherkin_text": row.GherkinText,
			"source_type": row.SourceType, "status": row.Status,
			"parser_message": "", "covered_gui_count": row.CoveredGuiCount,
			"total_gui_count": row.TotalGuiCount, "coverage_percent": row.CoveragePercent,
			"updated_at": updatedAt,
		})
	if result.Error != nil {
		return entities.GherkinScenario{}, result.Error
	}
	if result.RowsAffected != 1 {
		return entities.GherkinScenario{}, ErrCorrectionConflict
	}
	row.UpdatedAt = updatedAt
	return row, nil
}
