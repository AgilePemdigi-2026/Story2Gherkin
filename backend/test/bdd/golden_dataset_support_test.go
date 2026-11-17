package bdd_test

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/middlewares"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
	"gorm.io/gorm"
)

// Test accounts
var (
	goldenQAID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000031",
	)
	goldenSAID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000032",
	)
)

// ------------------------------------
// FAKE USER REPOSITORY
// ------------------------------------

type goldenBDDUsers map[uuid.UUID]entities.Role

func (u goldenBDDUsers) GetUserById(
	_ context.Context,
	_ *gorm.DB,
	id string,
) (entities.User, error) {

	parsed, err := uuid.Parse(id)
	if err != nil {
		return entities.User{}, gorm.ErrRecordNotFound
	}

	role, exists := u[parsed]
	if !exists {
		return entities.User{}, gorm.ErrRecordNotFound
	}

	return entities.User{
		ID:   parsed,
		Role: role,
	}, nil
}

// ------------------------------------
// FAKE GHERKIN VALIDATOR
// ------------------------------------

type goldenBDDValidator struct {
	valid bool
	err   error
	calls int
}

func (v *goldenBDDValidator) Validate(
	_ context.Context,
	_ string,
	elements []string,
) (dto.ScenarioValidation, error) {

	v.calls++

	if v.err != nil {
		return dto.ScenarioValidation{}, v.err
	}

	if !v.valid {
		return dto.ScenarioValidation{
			Valid: false,
		}, nil
	}

	return dto.ScenarioValidation{
		Valid:       true,
		FeatureName: "BDD Golden Dataset Test",
		Checks: dto.StructureChecks{
			Feature:  true,
			Scenario: true,
			Given:    true,
			When:     true,
			Then:     true,
		},
		Issues:             []dto.ValidationIssue{},
		UsedElements:       append([]string{}, elements...),
		NotCoveredElements: []string{},
		CoveragePercent:    100,
	}, nil
}

// ------------------------------------
// FAKE GOLDEN DATASET REPOSITORY
// ------------------------------------

type goldenBDDStore struct {
	rows       map[uuid.UUID]entities.GoldenDataset
	writes     int
	failCreate bool
}

func newGoldenBDDStore() *goldenBDDStore {
	return &goldenBDDStore{
		rows: make(map[uuid.UUID]entities.GoldenDataset),
	}
}

// Copy data to avoid changing stored references.
func cloneGoldenDataset(
	row entities.GoldenDataset,
) entities.GoldenDataset {

	row.Items = append(
		[]entities.GoldenDatasetItem(nil),
		row.Items...,
	)

	for i := range row.Items {
		row.Items[i].GUIElements = append(
			[]string(nil),
			row.Items[i].GUIElements...,
		)
	}

	return row
}

// Simulate saving a dataset and its items.
func (s *goldenBDDStore) CreateWithItems(
	ctx context.Context,
	dataset *entities.GoldenDataset,
	items []entities.GoldenDatasetItem,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	if dataset == nil || len(items) == 0 {
		return errors.New("dataset and items are required")
	}

	if s.failCreate {
		return errors.New("simulated database failure")
	}

	if dataset.ID == uuid.Nil {
		dataset.ID = uuid.New()
	}

	stamp := time.Now().UTC()

	dataset.CreatedAt = stamp
	dataset.UpdatedAt = stamp
	dataset.ItemCount = len(items)

	stored := cloneGoldenDataset(*dataset)

	stored.Items = make(
		[]entities.GoldenDatasetItem,
		len(items),
	)

	for i, item := range items {
		if item.ID == uuid.Nil {
			item.ID = uuid.New()
		}

		item.GoldenDatasetID = dataset.ID
		item.CreatedAt = stamp
		item.UpdatedAt = stamp

		item.GUIElements = append(
			[]string(nil),
			item.GUIElements...,
		)

		stored.Items[i] = item
	}

	s.rows[dataset.ID] = stored
	s.writes++

	return nil
}

// Simulate updating a dataset and optionally replacing its items.
func (s *goldenBDDStore) UpdateWithItems(
	ctx context.Context,
	id uuid.UUID,
	name string,
	description *string,
	items []entities.GoldenDatasetItem,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	row, exists := s.rows[id]

	if !exists {
		return gorm.ErrRecordNotFound
	}

	if row.UsedAt != nil {
		return repository.ErrGoldenDatasetLocked
	}

	stamp := time.Now().UTC()

	row.Name = name
	row.Description = description
	row.UpdatedAt = stamp

	if items != nil {
		if len(items) == 0 {
			return errors.New("dataset must contain at least one item")
		}

		stored := make(
			[]entities.GoldenDatasetItem,
			len(items),
		)

		for i, item := range items {
			if item.ID == uuid.Nil {
				item.ID = uuid.New()
			}

			item.GoldenDatasetID = id
			item.CreatedAt = stamp
			item.UpdatedAt = stamp

			item.GUIElements = append(
				[]string(nil),
				item.GUIElements...,
			)

			stored[i] = item
		}

		row.Items = stored
		row.ItemCount = len(stored)
	}

	s.rows[id] = row
	s.writes++

	return nil
}

// Simulate deleting a dataset and its items.
func (s *goldenBDDStore) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	row, exists := s.rows[id]

	if !exists {
		return gorm.ErrRecordNotFound
	}

	if row.UsedAt != nil {
		return repository.ErrGoldenDatasetLocked
	}

	delete(s.rows, id)
	s.writes++

	return nil
}

// Simulate searching and paginating datasets.
func (s *goldenBDDStore) List(
	ctx context.Context,
	search string,
	page int,
	pageSize int,
) ([]entities.GoldenDataset, int64, error) {

	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	search = strings.ToLower(strings.TrimSpace(search))

	all := make([]entities.GoldenDataset, 0)

	for _, row := range s.rows {
		description := ""

		if row.Description != nil {
			description = *row.Description
		}

		text := strings.ToLower(
			row.Name + " " + description,
		)

		if search != "" && !strings.Contains(text, search) {
			continue
		}

		copyRow := cloneGoldenDataset(row)
		copyRow.Items = nil

		all = append(all, copyRow)
	}

	sort.Slice(all, func(i, j int) bool {
		if all[i].CreatedAt.Equal(all[j].CreatedAt) {
			return all[i].ID.String() > all[j].ID.String()
		}

		return all[i].CreatedAt.After(all[j].CreatedAt)
	})

	total := int64(len(all))
	start := (page - 1) * pageSize

	if start >= len(all) {
		return []entities.GoldenDataset{}, total, nil
	}

	end := start + pageSize

	if end > len(all) {
		end = len(all)
	}

	return all[start:end], total, nil
}

// Simulate retrieving a dataset with its items.
func (s *goldenBDDStore) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*entities.GoldenDataset, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	row, exists := s.rows[id]

	if !exists {
		return nil, gorm.ErrRecordNotFound
	}

	copied := cloneGoldenDataset(row)

	return &copied, nil
}

// ------------------------------------
// TEST HTTP ROUTER
// ------------------------------------

// Use real middleware, handler, and service.
// Only external dependencies are replaced.
func goldenBDDRouter(
	store *goldenBDDStore,
	validator *goldenBDDValidator,
	users goldenBDDUsers,
	jwt authService.JWTService,
) *gin.Engine {

	svc := service.NewGoldenDatasetService(
		store,
		users,
		validator,
	)

	h := handler.NewGoldenDatasetHandler(svc)

	router := gin.New()

	routes := router.Group(
		"/api/v1/golden-datasets",
		middlewares.Authenticate(jwt),
	)

	scenario.RegisterGoldenDatasetRoutes(routes, h)

	return router
}
