package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

var ErrGoldenDatasetLocked = errors.New(
	"golden dataset has already been used in an evaluation",
)

// GoldenDatasetRepository defines database operations
// for Golden Dataset.
type GoldenDatasetRepository interface {
	CreateWithItems(
		ctx context.Context,
		dataset *entities.GoldenDataset,
		items []entities.GoldenDatasetItem,
	) error

	UpdateWithItems(
		ctx context.Context,
		id uuid.UUID,
		name string,
		description *string,
		items []entities.GoldenDatasetItem,
	) error

	Delete(
		ctx context.Context,
		id uuid.UUID,
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

func lockGoldenDataset(
	tx *gorm.DB,
	id uuid.UUID,
) (*entities.GoldenDataset, error) {

	var dataset entities.GoldenDataset

	if err := tx.Clauses(
		clause.Locking{Strength: "UPDATE"},
	).First(&dataset, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if dataset.UsedAt != nil {
		return nil, ErrGoldenDatasetLocked
	}

	return &dataset, nil
}

func (r *goldenDatasetRepository) UpdateWithItems(
	ctx context.Context,
	id uuid.UUID,
	name string,
	description *string,
	items []entities.GoldenDatasetItem,
) error {

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockGoldenDataset(tx, id); err != nil {
			return err
		}

		updates := map[string]any{
			"name":        name,
			"description": description,
		}

		if items != nil {
			if len(items) == 0 {
				return errors.New(
					"dataset must contain at least one item",
				)
			}

			updates["item_count"] = len(items)
		}

		if err := tx.Model(&entities.GoldenDataset{}).
			Where("id = ?", id).
			Updates(updates).Error; err != nil {

			return err
		}

		if items == nil {
			return nil
		}

		if err := tx.Where("golden_dataset_id = ?", id).
			Delete(&entities.GoldenDatasetItem{}).Error; err != nil {

			return err
		}

		for i := range items {
			items[i].GoldenDatasetID = id
		}

		return tx.Create(&items).Error
	})
}

func (r *goldenDatasetRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockGoldenDataset(tx, id); err != nil {
			return err
		}

		if err := tx.Where("golden_dataset_id = ?", id).
			Delete(&entities.GoldenDatasetItem{}).Error; err != nil {

			return err
		}

		return tx.Where("id = ?", id).
			Delete(&entities.GoldenDataset{}).Error
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
