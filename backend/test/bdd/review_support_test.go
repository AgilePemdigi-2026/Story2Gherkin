package bdd_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/middlewares"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
	"gorm.io/gorm"
)

var (
	reviewOwnerID    = uuid.MustParse("00000000-0000-4000-8000-000000000011")
	reviewOtherID    = uuid.MustParse("00000000-0000-4000-8000-000000000012")
	reviewQAID       = uuid.MustParse("00000000-0000-4000-8000-000000000013")
	reviewEmptyID    = uuid.MustParse("00000000-0000-4000-8000-000000000014")
	reviewScenarioID = uuid.MustParse("00000000-0000-4000-8000-000000000021")
)

const reviewOriginal = "Feature: User Login\r\n  Scenario: Successful login\r\n    Given the Email field and Password field are filled\r\n    When the user clicks the Login button\r\n    Then the Dashboard heading is displayed\r\n"

var reviewElements = []string{"Email field", "Password field", "Login button", "Dashboard heading"}

// These adapters isolate external persistence and Python validation. The actual
// JWT middleware, HTTP handlers, and ReviewService execute in every scenario.
type reviewBDDUsers map[uuid.UUID]entities.Role

func (users reviewBDDUsers) GetUserById(_ context.Context, _ *gorm.DB, id string) (entities.User, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return entities.User{}, gorm.ErrRecordNotFound
	}
	role, exists := users[parsed]
	if !exists {
		return entities.User{}, gorm.ErrRecordNotFound
	}
	return entities.User{ID: parsed, Role: role}, nil
}

type reviewBDDStore struct {
	rows   map[uuid.UUID]entities.GherkinScenario
	writes int
}

func cloneReviewRow(row entities.GherkinScenario) entities.GherkinScenario {
	row.GenerationRequest.GuiElements = append([]string(nil), row.GenerationRequest.GuiElements...)
	if row.GenerationRequestID != nil {
		id := *row.GenerationRequestID
		row.GenerationRequestID = &id
	}
	return row
}

func (store *reviewBDDStore) ListOwned(_ context.Context, owner uuid.UUID) ([]entities.GherkinScenario, error) {
	rows := []entities.GherkinScenario{}
	for _, row := range store.rows {
		if row.GenerationRequest.UserID == owner {
			rows = append(rows, cloneReviewRow(row))
		}
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ScenarioNumber < rows[j].ScenarioNumber })
	return rows, nil
}

func (store *reviewBDDStore) GetOwned(_ context.Context, owner, id uuid.UUID) (entities.GherkinScenario, error) {
	row, exists := store.rows[id]
	if !exists || row.GenerationRequest.UserID != owner {
		return entities.GherkinScenario{}, gorm.ErrRecordNotFound
	}
	return cloneReviewRow(row), nil
}

func (store *reviewBDDStore) SaveCorrectionOwned(ctx context.Context, owner uuid.UUID, row entities.GherkinScenario, expected time.Time) (entities.GherkinScenario, error) {
	current, err := store.GetOwned(ctx, owner, row.ID)
	if err != nil {
		return entities.GherkinScenario{}, err
	}
	if !current.UpdatedAt.Equal(expected) {
		return entities.GherkinScenario{}, repository.ErrCorrectionConflict
	}
	row.ScenarioNumber = current.ScenarioNumber
	row.UpdatedAt = current.UpdatedAt.Add(time.Second)
	store.rows[row.ID] = cloneReviewRow(row)
	store.writes++
	return cloneReviewRow(row), nil
}

type reviewBDDValidator struct {
	reports      map[string]dto.ScenarioValidation
	err          error
	calls        int
	lastText     string
	lastElements []string
}

func (validator *reviewBDDValidator) Validate(_ context.Context, text string, elements []string) (dto.ScenarioValidation, error) {
	validator.calls++
	validator.lastText = text
	validator.lastElements = append([]string(nil), elements...)
	if validator.err != nil {
		return dto.ScenarioValidation{}, validator.err
	}
	report, exists := validator.reports[text]
	if !exists {
		return dto.ScenarioValidation{}, fmt.Errorf("unexpected text sent to review validator fixture")
	}
	return report, nil
}

func reviewValidReport(title string, used int) dto.ScenarioValidation {
	available := true
	return dto.ScenarioValidation{
		Valid: true, FeatureName: title,
		Checks:             dto.StructureChecks{Feature: true, Scenario: true, Given: true, When: true, Then: true},
		Issues:             []dto.ValidationIssue{},
		UsedElements:       append([]string{}, reviewElements[:used]...),
		NotCoveredElements: append([]string{}, reviewElements[used:]...),
		CoveragePercent:    float64(used) / float64(len(reviewElements)) * 100,
		CoverageAvailable:  &available,
	}
}

func reviewBDDRouter(store *reviewBDDStore, validator *reviewBDDValidator, users reviewBDDUsers, jwt authService.JWTService) *gin.Engine {
	svc := service.NewReviewService(validator, users, store)
	h := handler.NewReviewHandler(svc)
	router := gin.New()
	group := router.Group("/api/v1/scenarios", middlewares.Authenticate(jwt))
	group.GET("", h.List)
	group.GET("/:id", h.Detail)
	group.PUT("/:id", h.Save)
	group.GET("/:id/download", h.Download)
	return router
}
