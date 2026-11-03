package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/gorm"
)

var (
	ErrReviewForbidden  = errors.New("only System Analysts may review scenarios")
	ErrScenarioNotFound = errors.New("scenario not found")
)

type InvalidCorrectionError struct{ Validation dto.ScenarioValidation }

func (*InvalidCorrectionError) Error() string { return "Save failed. Invalid Gherkin structure." }

type ScenarioValidator interface {
	Validate(context.Context, string, []string) (dto.ScenarioValidation, error)
}

type ReviewUserRepository interface {
	GetUserById(context.Context, *gorm.DB, string) (entities.User, error)
}

type ReviewService interface {
	List(context.Context, string) ([]dto.ScenarioSummary, error)
	Detail(context.Context, string, string) (dto.ReviewScenario, error)
	Save(context.Context, string, string, dto.SaveCorrectionRequest) (dto.ReviewScenario, error)
	Download(context.Context, string, string) (dto.ReviewScenario, error)
}

type reviewService struct {
	validator ScenarioValidator
	users     ReviewUserRepository
	scenarios repository.ReviewRepository
}

func NewReviewService(validator ScenarioValidator, users ReviewUserRepository, scenarios repository.ReviewRepository) ReviewService {
	return &reviewService{validator: validator, users: users, scenarios: scenarios}
}

func (s *reviewService) authorize(ctx context.Context, userID string) (uuid.UUID, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, ErrReviewForbidden
	}
	user, err := s.users.GetUserById(ctx, nil, id.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, ErrReviewForbidden
	}
	if err != nil {
		return uuid.Nil, err
	}
	if user.Role != entities.RoleSystemAnalyst {
		return uuid.Nil, ErrReviewForbidden
	}
	return id, nil
}

func (s *reviewService) List(ctx context.Context, userID string) ([]dto.ScenarioSummary, error) {
	owner, err := s.authorize(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.scenarios.ListOwned(ctx, owner)
	if err != nil {
		return nil, err
	}
	items := make([]dto.ScenarioSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, reviewSummary(row))
	}
	return items, nil
}

func reviewSummary(row entities.GherkinScenario) dto.ScenarioSummary {
	return dto.ScenarioSummary{ID: row.ID.String(), FeatureName: row.FeatureName,
		ScenarioCode: fmt.Sprintf("SCN-%04d", row.ScenarioNumber),
		SourceType:   string(row.SourceType), Status: string(row.Status), UpdatedAt: row.UpdatedAt}
}

func (s *reviewService) load(ctx context.Context, userID, scenarioID string) (uuid.UUID, entities.GherkinScenario, error) {
	owner, err := s.authorize(ctx, userID)
	if err != nil {
		return uuid.Nil, entities.GherkinScenario{}, err
	}
	id, err := uuid.Parse(scenarioID)
	if err != nil {
		return owner, entities.GherkinScenario{}, ErrScenarioNotFound
	}
	row, err := s.scenarios.GetOwned(ctx, owner, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = ErrScenarioNotFound
	}
	return owner, row, err
}

func (s *reviewService) describe(ctx context.Context, row entities.GherkinScenario) (dto.ReviewScenario, error) {
	validation, err := s.validator.Validate(ctx, row.GherkinText, row.GenerationRequest.GuiElements)
	if err != nil {
		return dto.ReviewScenario{}, err
	}
	return dto.ReviewScenario{ScenarioSummary: reviewSummary(row), Gherkin: row.GherkinText,
		GUIElements: row.GenerationRequest.GuiElements, Validation: validation}, nil
}

func (s *reviewService) Detail(ctx context.Context, userID, scenarioID string) (dto.ReviewScenario, error) {
	_, row, err := s.load(ctx, userID, scenarioID)
	if err != nil {
		return dto.ReviewScenario{}, err
	}
	return s.describe(ctx, row)
}

func (s *reviewService) Save(ctx context.Context, userID, scenarioID string, input dto.SaveCorrectionRequest) (dto.ReviewScenario, error) {
	owner, row, err := s.load(ctx, userID, scenarioID)
	if err != nil {
		return dto.ReviewScenario{}, err
	}
	if utf8.RuneCountInString(input.Gherkin) > 65536 || input.UpdatedAt.IsZero() {
		return dto.ReviewScenario{}, ErrInvalidRequest
	}
	if !row.UpdatedAt.Equal(input.UpdatedAt) {
		return dto.ReviewScenario{}, repository.ErrCorrectionConflict
	}
	validation, err := s.validator.Validate(ctx, input.Gherkin, row.GenerationRequest.GuiElements)
	if err != nil {
		return dto.ReviewScenario{}, err
	}
	if !validation.Valid {
		return dto.ReviewScenario{}, &InvalidCorrectionError{Validation: validation}
	}
	row.GherkinText = input.Gherkin // Preserve every byte of the validated draft.
	row.FeatureName = validation.FeatureName
	row.SourceType, row.Status = dto.SourceCorrected, dto.ScenarioValid
	row.ParserMessage = ""
	row.CoveredGuiCount, row.TotalGuiCount = len(validation.UsedElements), len(row.GenerationRequest.GuiElements)
	row.CoveragePercent = validation.CoveragePercent
	row, err = s.scenarios.SaveCorrectionOwned(ctx, owner, row, input.UpdatedAt)
	if err != nil {
		return dto.ReviewScenario{}, err
	}
	return dto.ReviewScenario{ScenarioSummary: reviewSummary(row), Gherkin: row.GherkinText,
		GUIElements: row.GenerationRequest.GuiElements, Validation: validation}, nil
}

func (s *reviewService) Download(ctx context.Context, userID, scenarioID string) (dto.ReviewScenario, error) {
	result, err := s.Detail(ctx, userID, scenarioID)
	if err != nil {
		return result, err
	}
	if result.Status != string(dto.ScenarioValid) || !result.Validation.Valid {
		return result, &InvalidCorrectionError{Validation: result.Validation}
	}
	return result, nil
}

func FeatureFilename(title string) string {
	var slug strings.Builder
	for _, r := range strings.ToLower(title) {
		if slug.Len() >= 70 {
			break
		}
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			slug.WriteRune(r)
		} else {
			slug.WriteByte('-')
		}
	}
	name := strings.Trim(slug.String(), "-")
	if name == "" {
		name = "gherkin"
	}
	return "scenario-" + name + ".feature"
}
