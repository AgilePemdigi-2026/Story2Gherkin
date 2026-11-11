package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
	"strings"
)

// GoldenDatasetRepository defines database operations
// for Golden Dataset.
type GoldenDatasetRepository interface {
	CreateWithItems(
		ctx context.Context,
		dataset *entities.GoldenDataset,
		items []entities.GoldenDatasetItem,
	) error

	List(
		ctx context.Context,
		search string,
		page int,
		pageSize int,
	) ([]entities.GoldenDataset, int64, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*entities.GoldenDataset, error)
}

// Repository implementation
type goldenDatasetRepository struct {
	db *gorm.DB
}

// Constructor
func NewGoldenDatasetRepository(db *gorm.DB) GoldenDatasetRepository {
	return &goldenDatasetRepository{
		db: db,
	}
}

// CreateWithItems saves the dataset and its items
// using a single database transaction.
func (r *goldenDatasetRepository) CreateWithItems(
	ctx context.Context,
	dataset *entities.GoldenDataset,
	items []entities.GoldenDatasetItem,
) error {

	if dataset == nil {
		return errors.New("dataset is required")
	}

	if len(items) == 0 {
		return errors.New("dataset must contain at least one item")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Set item count
		dataset.ItemCount = len(items)

		// Insert parent dataset
		if err := tx.Omit("Items").Create(dataset).Error; err != nil {
			return err
		}

		// Assign foreign key
		for i := range items {
			items[i].GoldenDatasetID = dataset.ID
		}

		// Insert all items
		if err := tx.Create(&items).Error; err != nil {
			return err
		}

		return nil
	})
}

// List retrieves datasets with search and pagination.
func (r *goldenDatasetRepository) List(
	ctx context.Context,
	search string,
	page int,
	pageSize int,
) ([]entities.GoldenDataset, int64, error) {

	// Default pagination
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	// Maximum page size
	if pageSize > 100 {
		pageSize = 100
	}

	search = strings.TrimSpace(search)

	// Base query
	query := r.db.WithContext(ctx).
		Model(&entities.GoldenDataset{})

	// Search by name or description
	if search != "" {
		keyword := "%" + search + "%"

		query = query.Where(
			"(name ILIKE ? OR description ILIKE ?)",
			keyword,
			keyword,
		)
	}

	// Count matching datasets
	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * pageSize

	datasets := make([]entities.GoldenDataset, 0)

	// Retrieve newest datasets first
	err := query.
		Order("created_at DESC, id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&datasets).Error

	if err != nil {
		return nil, 0, err
	}

	return datasets, total, nil
}

// GetByID retrieves a Golden Dataset and its reference items.
func (r *goldenDatasetRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*entities.GoldenDataset, error) {

	var dataset entities.GoldenDataset

	err := r.db.WithContext(ctx).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC, id ASC")
		}).
		First(&dataset, "id = ?", id).
		Error

	if err != nil {
		return nil, err
	}

	return &dataset, nil
}
