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
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if utf8.RuneCountInString(name) < 1 ||
		utf8.RuneCountInString(name) > 200 ||
		utf8.RuneCountInString(description) > 2000 {

		return empty, fmt.Errorf(
			"%w: invalid dataset name or description",
			ErrGoldenDatasetInvalid,
		)
	}

	// 3. Validate JSON file size
	if len(raw) == 0 ||
		len(raw) > maxGoldenDatasetFileBytes {

		return empty, fmt.Errorf(
			"%w: file must be nonempty and at most 5 MB",
			ErrGoldenDatasetInvalid,
		)
	}

	// 4. Parse JSON
	var file dto.GoldenDatasetFile

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&file); err != nil {
		return empty, fmt.Errorf(
			"%w: malformed JSON file: %v",
			ErrGoldenDatasetInvalid, err,
		)
	}

	// Reject additional JSON objects
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return empty, fmt.Errorf(
			"%w: file must contain exactly one JSON object",
			ErrGoldenDatasetInvalid,
		)
	}

	// 5. Validate item count
	if len(file.Items) < 1 ||
		len(file.Items) > maxGoldenDatasetItems {

		return empty, fmt.Errorf(
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

	// 6. Validate each item
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

			return empty, fmt.Errorf(
				"%w: item %d contains missing or out-of-range fields",
				ErrGoldenDatasetInvalid, n,
			)
		}

		// 7. Reject duplicate item titles
		key := strings.ToLower(
			strings.Join(strings.Fields(input.ItemTitle), " "),
		)

		if _, exists := seen[key]; exists {
			return empty, fmt.Errorf(
				"%w: duplicate item_title at item %d",
				ErrGoldenDatasetInvalid, n,
			)
		}

		seen[key] = struct{}{}

		// 8. Validate GUI elements
		for j, element := range input.GUIElements {
			element = strings.TrimSpace(element)

			if utf8.RuneCountInString(element) < 1 ||
				utf8.RuneCountInString(element) > 200 {

				return empty, fmt.Errorf(
					"%w: invalid GUI element at item %d, position %d",
					ErrGoldenDatasetInvalid, n, j+1,
				)
			}

			input.GUIElements[j] = element
		}

		// 9. Validate expected Gherkin
		validation, err := s.validator.Validate(
			ctx,
			input.ExpectedGherkin,
			input.GUIElements,
		)

		if err != nil {
			return empty, fmt.Errorf(
				"validation service failed for item %d: %w",
				n, err,
			)
		}

		if !validation.Valid {
			return empty, fmt.Errorf(
				"%w: invalid expected_gherkin at item %d",
				ErrGoldenDatasetInvalid, n,
			)
		}

		// 10. Prepare validated item
		items = append(items, entities.GoldenDatasetItem{
			ItemTitle:        input.ItemTitle,
			UserStoryText:    input.UserStoryText,
			GUIElements:      input.GUIElements,
			FreeScenarioText: input.FreeScenarioText,
			ExpectedGherkin:  input.ExpectedGherkin,
		})
	}

	// 11. Prepare dataset metadata
	var desc *string
	if description != "" {
		desc = &description
	}

	dataset := &entities.GoldenDataset{
		Name:        name,
		Description: desc,
		CreatedBy:   createdBy,
	}

	// 12. Save dataset and items atomically
	if err := s.repo.CreateWithItems(
		ctx,
		dataset,
		items,
	); err != nil {
		return empty, err
	}

	// 13. Return response
	return dto.GoldenDatasetSummary{
		ID:          dataset.ID.String(),
		Name:        dataset.Name,
		Description: dataset.Description,
		CreatedBy:   dataset.CreatedBy,
		ItemCount:   dataset.ItemCount,
		UsedAt:      dataset.UsedAt,
		CreatedAt:   dataset.CreatedAt,
		UpdatedAt:   dataset.UpdatedAt,
	}, nil
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
