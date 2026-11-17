package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/gorm"
)

var (
	ErrGoldenDatasetForbidden = errors.New(
		"only Quality Assurance users may manage golden datasets",
	)

	ErrGoldenDatasetNotFound = errors.New("golden dataset not found")

	ErrGoldenDatasetLocked = errors.New(
		"golden dataset has already been used in an evaluation",
	)

	ErrGoldenDatasetInvalid = errors.New(
		"invalid golden dataset",
	)
)

const (
	maxGoldenDatasetFileBytes = 5 << 20
	maxGoldenDatasetItems     = 100
)

// GoldenDatasetService handles Golden Dataset uploads.

type GoldenDatasetService interface {
	Upload(
		ctx context.Context,
		userID, name, description string,
		raw []byte,
	) (dto.GoldenDatasetSummary, error)

	List(
		ctx context.Context,
		userID string,
		search string,
		page int,
		pageSize int,
	) (dto.GoldenDatasetListResponse, error)

	GetDetail(
		ctx context.Context,
		userID string,
		datasetID string,
	) (dto.GoldenDatasetDetailResponse, error)

	Update(
		ctx context.Context,
		userID, datasetID, name, description string,
		raw []byte,
	) (dto.GoldenDatasetSummary, error)

	Delete(
		ctx context.Context,
		userID, datasetID string,
	) error
}

type goldenDatasetService struct {
	repo      repository.GoldenDatasetRepository
	users     ReviewUserRepository
	validator ScenarioValidator
}

// Constructor
func NewGoldenDatasetService(
	repo repository.GoldenDatasetRepository,
	users ReviewUserRepository,
	validator ScenarioValidator,
) GoldenDatasetService {
	return &goldenDatasetService{
		repo:      repo,
		users:     users,
		validator: validator,
	}
}

// authorizeQA verifies that the user exists and has QA role.
func (s *goldenDatasetService) authorizeQA(
	ctx context.Context,
	userID string,
) (string, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return "", ErrGoldenDatasetForbidden
	}

	user, err := s.users.GetUserById(
		ctx, nil, id.String(),
	)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrGoldenDatasetForbidden
	}

	if err != nil {
		return "", err
	}

	if user.Role != entities.RoleQA {
		return "", ErrGoldenDatasetForbidden
	}

	return user.ID.String(), nil
}

// Upload validates and stores a Golden Dataset.
func (s *goldenDatasetService) Upload(
	ctx context.Context,
	userID, name, description string,
	raw []byte,
) (dto.GoldenDatasetSummary, error) {

	var empty dto.GoldenDatasetSummary

	// 1. Verify QA authorization
	createdBy, err := s.authorizeQA(ctx, userID)
	if err != nil {
		return empty, err
	}

	// 2. Validate dataset metadata
	name, desc, err := validateGoldenMetadata(name, description)
	if err != nil {
		return empty, err
	}

	// 3. Parse and validate items
	items, err := s.buildItems(ctx, raw)
	if err != nil {
		return empty, err
	}

	// 4. Prepare dataset metadata
	dataset := &entities.GoldenDataset{
		Name:        name,
		Description: desc,
		CreatedBy:   createdBy,
	}

	// 5. Save dataset and items atomically
	if err := s.repo.CreateWithItems(
		ctx,
		dataset,
		items,
	); err != nil {
		return empty, err
	}

	return toGoldenDatasetSummary(*dataset), nil
}

func validateGoldenMetadata(
	name, description string,
) (string, *string, error) {

	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if utf8.RuneCountInString(name) < 1 ||
		utf8.RuneCountInString(name) > 200 ||
		utf8.RuneCountInString(description) > 2000 {

		return "", nil, fmt.Errorf(
			"%w: invalid dataset name or description",
			ErrGoldenDatasetInvalid,
		)
	}

	var desc *string
	if description != "" {
		desc = &description
	}

	return name, desc, nil
}

func (s *goldenDatasetService) buildItems(
	ctx context.Context,
	raw []byte,
) ([]entities.GoldenDatasetItem, error) {

	if len(raw) == 0 ||
		len(raw) > maxGoldenDatasetFileBytes {

		return nil, fmt.Errorf(
			"%w: file must be nonempty and at most 5 MB",
			ErrGoldenDatasetInvalid,
		)
	}

	var file dto.GoldenDatasetFile

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf(
			"%w: malformed JSON file: %v",
			ErrGoldenDatasetInvalid, err,
		)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf(
			"%w: file must contain exactly one JSON object",
			ErrGoldenDatasetInvalid,
		)
	}

	if len(file.Items) < 1 ||
		len(file.Items) > maxGoldenDatasetItems {

		return nil, fmt.Errorf(
			"%w: item count must be between 1 and %d",
			ErrGoldenDatasetInvalid,
			maxGoldenDatasetItems,
		)
	}

	items := make(
		[]entities.GoldenDatasetItem,
		0,
		len(file.Items),
	)

	seen := make(map[string]struct{}, len(file.Items))

	for index, input := range file.Items {

		input.ItemTitle = strings.TrimSpace(input.ItemTitle)
		input.UserStoryText = strings.TrimSpace(input.UserStoryText)
		input.FreeScenarioText = strings.TrimSpace(
			input.FreeScenarioText,
		)

		n := index + 1

		titleLen := utf8.RuneCountInString(input.ItemTitle)
		storyLen := utf8.RuneCountInString(input.UserStoryText)
		freeLen := utf8.RuneCountInString(input.FreeScenarioText)

		if titleLen < 1 || titleLen > 200 ||
			storyLen < 10 || storyLen > 8000 ||
			freeLen < 10 || freeLen > 4000 ||
			len(input.GUIElements) < 1 ||
			len(input.GUIElements) > 30 ||
			strings.TrimSpace(input.ExpectedGherkin) == "" ||
			utf8.RuneCountInString(input.ExpectedGherkin) > 65536 {

			return nil, fmt.Errorf(
				"%w: item %d contains missing or out-of-range fields",
				ErrGoldenDatasetInvalid, n,
			)
		}

		key := strings.ToLower(
			strings.Join(strings.Fields(input.ItemTitle), " "),
		)

		if _, exists := seen[key]; exists {
			return nil, fmt.Errorf(
				"%w: duplicate item_title at item %d",
				ErrGoldenDatasetInvalid, n,
			)
		}

		seen[key] = struct{}{}

		for j, element := range input.GUIElements {
			element = strings.TrimSpace(element)

			if utf8.RuneCountInString(element) < 1 ||
				utf8.RuneCountInString(element) > 200 {

				return nil, fmt.Errorf(
					"%w: invalid GUI element at item %d, position %d",
					ErrGoldenDatasetInvalid, n, j+1,
				)
			}

			input.GUIElements[j] = element
		}

		validation, err := s.validator.Validate(
			ctx,
			input.ExpectedGherkin,
			input.GUIElements,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"validation service failed for item %d: %w",
				n, err,
			)
		}

		if !validation.Valid {
			return nil, fmt.Errorf(
				"%w: invalid expected_gherkin at item %d",
				ErrGoldenDatasetInvalid, n,
			)
		}

		items = append(items, entities.GoldenDatasetItem{
			ItemTitle:        input.ItemTitle,
			UserStoryText:    input.UserStoryText,
			GUIElements:      input.GUIElements,
			FreeScenarioText: input.FreeScenarioText,
			ExpectedGherkin:  input.ExpectedGherkin,
		})
	}

	return items, nil
}

func (s *goldenDatasetService) Update(
	ctx context.Context,
	userID, datasetID, name, description string,
	raw []byte,
) (dto.GoldenDatasetSummary, error) {

	var empty dto.GoldenDatasetSummary

	if _, err := s.authorizeQA(ctx, userID); err != nil {
		return empty, err
	}

	id, err := uuid.Parse(datasetID)
	if err != nil {
		return empty, ErrGoldenDatasetNotFound
	}

	name, desc, err := validateGoldenMetadata(name, description)
	if err != nil {
		return empty, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return empty, mapGoldenWriteError(err)
	}

	if existing.UsedAt != nil {
		return empty, ErrGoldenDatasetLocked
	}

	var items []entities.GoldenDatasetItem

	if raw != nil {
		items, err = s.buildItems(ctx, raw)
		if err != nil {
			return empty, err
		}
	}

	if err := s.repo.UpdateWithItems(
		ctx, id, name, desc, items,
	); err != nil {
		return empty, mapGoldenWriteError(err)
	}

	dataset, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return empty, err
	}

	return toGoldenDatasetSummary(*dataset), nil
}

func (s *goldenDatasetService) Delete(
	ctx context.Context,
	userID, datasetID string,
) error {

	if _, err := s.authorizeQA(ctx, userID); err != nil {
		return err
	}

	id, err := uuid.Parse(datasetID)
	if err != nil {
		return ErrGoldenDatasetNotFound
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return mapGoldenWriteError(err)
	}

	return nil
}

func mapGoldenWriteError(err error) error {
	switch {
	case errors.Is(err, repository.ErrGoldenDatasetLocked):
		return ErrGoldenDatasetLocked

	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrGoldenDatasetNotFound
	}

	return err
}

// Converts a database entity into a response DTO.
func toGoldenDatasetSummary(
	dataset entities.GoldenDataset,
) dto.GoldenDatasetSummary {

	return dto.GoldenDatasetSummary{
		ID:          dataset.ID.String(),
		Name:        dataset.Name,
		Description: dataset.Description,
		CreatedBy:   dataset.CreatedBy,
		ItemCount:   dataset.ItemCount,
		UsedAt:      dataset.UsedAt,
		CreatedAt:   dataset.CreatedAt,
		UpdatedAt:   dataset.UpdatedAt,
	}
}

// List returns paginated Golden Datasets for QA users.
func (s *goldenDatasetService) List(
	ctx context.Context,
	userID string,
	search string,
	page int,
	pageSize int,
) (dto.GoldenDatasetListResponse, error) {

	var empty dto.GoldenDatasetListResponse

	// 1. Verify QA authorization
	if _, err := s.authorizeQA(ctx, userID); err != nil {
		return empty, err
	}

	// 2. Normalize pagination
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	// 3. Retrieve datasets
	datasets, total, err := s.repo.List(
		ctx,
		search,
		page,
		pageSize,
	)

	if err != nil {
		return empty, err
	}

	// 4. Map entities to response DTOs
	summaries := make(
		[]dto.GoldenDatasetSummary,
		0,
		len(datasets),
	)

	for _, dataset := range datasets {
		summaries = append(
			summaries,
			toGoldenDatasetSummary(dataset),
		)
	}

	// 5. Calculate total pages
	totalPages := int(
		(total + int64(pageSize) - 1) / int64(pageSize),
	)

	return dto.GoldenDatasetListResponse{
		Datasets:   summaries,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetDetail returns one Golden Dataset and its items.
func (s *goldenDatasetService) GetDetail(
	ctx context.Context,
	userID string,
	datasetID string,
) (dto.GoldenDatasetDetailResponse, error) {

	var empty dto.GoldenDatasetDetailResponse

	// 1. Verify QA authorization
	if _, err := s.authorizeQA(ctx, userID); err != nil {
		return empty, err
	}

	// 2. Parse dataset UUID
	id, err := uuid.Parse(datasetID)
	if err != nil {
		return empty, ErrGoldenDatasetNotFound
	}

	// 3. Retrieve dataset and its items
	dataset, err := s.repo.GetByID(ctx, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return empty, ErrGoldenDatasetNotFound
	}

	if err != nil {
		return empty, err
	}

	// 4. Convert reference items to DTOs
	items := make(
		[]dto.GoldenDatasetItemResponse,
		0,
		len(dataset.Items),
	)

	for _, item := range dataset.Items {
		items = append(items, dto.GoldenDatasetItemResponse{
			ID:               item.ID.String(),
			ItemTitle:        item.ItemTitle,
			UserStoryText:    item.UserStoryText,
			GUIElements:      item.GUIElements,
			FreeScenarioText: item.FreeScenarioText,
			ExpectedGherkin:  item.ExpectedGherkin,
			CreatedAt:        item.CreatedAt,
		})
	}

	// 5. Build detail response
	return dto.GoldenDatasetDetailResponse{
		GoldenDatasetSummary: toGoldenDatasetSummary(
			*dataset,
		),
		Items: items,
	}, nil
}
