package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
)

type HistoryRepository interface {
	ListOwned(
		ctx context.Context,
		owner uuid.UUID,
		page int,
		pageSize int,
	) ([]entities.GenerationRequest, int64, error)

	GetOwned(
		ctx context.Context,
		owner uuid.UUID,
		requestID uuid.UUID,
	) (
		entities.GenerationRequest,
		*entities.GherkinScenario,
		error,
	)
}

type historyRepository struct {
	db *gorm.DB
}

func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return &historyRepository{db: db}
}

// ListOwned mengambil riwayat Generate SUCCESS milik user.
// Request FAILED dan PENDING tidak ditampilkan di History.
func (r *historyRepository) ListOwned(
	ctx context.Context,
	owner uuid.UUID,
	page int,
	pageSize int,
) ([]entities.GenerationRequest, int64, error) {

	rows := make([]entities.GenerationRequest, 0)
	var total int64

	// Pada database proyek, user_id bertipe text.
	// Karena itu UUID dikonversi menjadi string.
	ownerID := owner.String()

	// Hitung hanya request SUCCESS milik user.
	err := r.db.WithContext(ctx).
		Model(&entities.GenerationRequest{}).
		Where(
			"user_id = ? AND status = ?",
			ownerID,
			"SUCCESS",
		).
		Count(&total).Error

	if err != nil {
		return nil, 0, err
	}

	// Ambil hanya request SUCCESS milik user.
	err = r.db.WithContext(ctx).
		Model(&entities.GenerationRequest{}).
		Where(
			"user_id = ? AND status = ?",
			ownerID,
			"SUCCESS",
		).
		Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	return rows, total, nil
}

// GetOwned mengambil detail Generate SUCCESS milik user.
// Request FAILED, PENDING, dan milik user lain
// dianggap tidak ditemukan.
func (r *historyRepository) GetOwned(
	ctx context.Context,
	owner uuid.UUID,
	requestID uuid.UUID,
) (
	entities.GenerationRequest,
	*entities.GherkinScenario,
	error,
) {

	var request entities.GenerationRequest

	// Periksa kepemilikan dan status SUCCESS
	// langsung melalui query database.
	err := r.db.WithContext(ctx).
		Where(
			"id = ? AND user_id = ? AND status = ?",
			requestID,
			owner.String(),
			"SUCCESS",
		).
		Take(&request).Error

	if err != nil {
		return entities.GenerationRequest{}, nil, err
	}

	var scenario entities.GherkinScenario

	// Ambil Generated Gherkin yang terhubung
	// dengan request SUCCESS tersebut.
	err = r.db.WithContext(ctx).
		Where("generation_request_id = ?", requestID).
		Order("created_at DESC, id DESC").
		Take(&scenario).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return request, nil, nil
	}

	if err != nil {
		return entities.GenerationRequest{}, nil, err
	}

	return request, &scenario, nil
}
