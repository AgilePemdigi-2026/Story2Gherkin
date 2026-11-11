package bdd_test

import (
	"context"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/middlewares"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/service"
	"gorm.io/gorm"
)

// ------------------------------------
// TEST IDENTIFIERS
// ------------------------------------

var (
	historyOwnerID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000051",
	)
	historyOtherID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000052",
	)
	historyQAID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000053",
	)
	historySuccessID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000061",
	)
	historyFailedID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000062",
	)
	historyOtherReqID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000063",
	)
	historyScenarioID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000071",
	)
	historySecondSuccessID = uuid.MustParse(
		"00000000-0000-4000-8000-000000000064",
	)
)

// ------------------------------------
// TEST DATA
// ------------------------------------

const (
	historySuccessStory = "As a user, I want to log in " +
		"so that I can access my dashboard."

	historyFailedStory = "As a user, I want to reset my " +
		"password so I can regain access."

	historyOtherStory = "As a user, I want to search " +
		"products by keyword."

	historySuccessFree = "User enters valid credentials " +
		"and clicks Login."

	historyFailedFree = "User requests a password reset."

	historyFailureMessage = "AI Service unavailable"

	historySuccessGherkin = "Feature: User Login\n" +
		"  Scenario: Successful login\n" +
		"    Given the user is on the login page\n" +
		"    When the user submits valid credentials\n" +
		"    Then the dashboard is displayed\n"
)

var (
	historySuccessElements = []string{
		"Email field",
		"Password field",
		"Login button",
	}

	historyFailedElements = []string{
		"Email field",
		"Reset password button",
	}
)

// ------------------------------------
// FAKE USER REPOSITORY
// ------------------------------------

type historyBDDUsers map[uuid.UUID]entities.Role

func (u historyBDDUsers) GetUserById(
	_ context.Context,
	_ *gorm.DB,
	id string,
) (entities.User, error) {

	parsed, err := uuid.Parse(id)
	if err != nil {
		return entities.User{}, gorm.ErrRecordNotFound
	}

	role, ok := u[parsed]
	if !ok {
		return entities.User{}, gorm.ErrRecordNotFound
	}

	return entities.User{
		ID:   parsed,
		Role: role,
	}, nil
}

// ------------------------------------
// FAKE HISTORY REPOSITORY
// ------------------------------------

type historyBDDStore struct {
	requests  map[uuid.UUID]entities.GenerationRequest
	scenarios map[uuid.UUID]entities.GherkinScenario

	listCalls   int
	detailCalls int
}

func newHistoryBDDStore() *historyBDDStore {
	return &historyBDDStore{
		requests: make(
			map[uuid.UUID]entities.GenerationRequest,
		),
		scenarios: make(
			map[uuid.UUID]entities.GherkinScenario,
		),
	}
}

// Prevent accidental modification of fixture data.
func cloneHistoryRequest(
	r entities.GenerationRequest,
) entities.GenerationRequest {

	r.GuiElements = append(
		[]string(nil),
		r.GuiElements...,
	)

	return r
}

func cloneHistoryScenario(
	s entities.GherkinScenario,
) entities.GherkinScenario {

	if s.GenerationRequestID != nil {
		id := *s.GenerationRequestID
		s.GenerationRequestID = &id
	}

	s.GenerationRequest = cloneHistoryRequest(
		s.GenerationRequest,
	)

	return s
}

// ------------------------------------
// FIXTURE: SUCCESS
// ------------------------------------

func (s *historyBDDStore) seedSuccess() {
	created := time.Date(
		2026, 11, 11, 9, 0, 0, 0, time.UTC,
	)

	request := entities.GenerationRequest{
		ID:           historySuccessID,
		UserID:       historyOwnerID,
		UserStory:    historySuccessStory,
		GuiElements:  append([]string(nil), historySuccessElements...),
		FreeScenario: historySuccessFree,
		Status:       entities.RequestStatus("SUCCESS"),
		Timestamp: entities.Timestamp{
			CreatedAt: created,
			UpdatedAt: created,
		},
	}

	s.requests[request.ID] = request

	requestID := request.ID

	s.scenarios[historyScenarioID] = entities.GherkinScenario{
		ID:                  historyScenarioID,
		ScenarioNumber:      7,
		GenerationRequestID: &requestID,
		FeatureName:         "User Login",
		GherkinText:         historySuccessGherkin,
		SourceType:          entities.SourceType("LLM"),
		Status:              entities.ScenarioStatus("VALID"),
		CoveredGuiCount:     3,
		TotalGuiCount:       3,
		CoveragePercent:     100,
		Timestamp: entities.Timestamp{
			CreatedAt: created,
			UpdatedAt: created,
		},
	}
}

func (s *historyBDDStore) seedSecondSuccess() {
	created := time.Date(
		2026, 11, 11, 12, 0, 0, 0, time.UTC,
	)

	s.requests[historySecondSuccessID] =
		entities.GenerationRequest{
			ID:        historySecondSuccessID,
			UserID:    historyOwnerID,
			UserStory: "As an admin, I want to manage users.",
			GuiElements: []string{
				"User table",
				"Add user button",
			},
			FreeScenario: "Admin opens the user management page.",
			Status:       entities.RequestStatus("SUCCESS"),
			Timestamp: entities.Timestamp{
				CreatedAt: created,
				UpdatedAt: created,
			},
		}
}

// ------------------------------------
// FIXTURE: FAILED
// ------------------------------------

func (s *historyBDDStore) seedFailed() {
	created := time.Date(
		2026, 11, 11, 10, 0, 0, 0, time.UTC,
	)

	s.requests[historyFailedID] = entities.GenerationRequest{
		ID:           historyFailedID,
		UserID:       historyOwnerID,
		UserStory:    historyFailedStory,
		GuiElements:  append([]string(nil), historyFailedElements...),
		FreeScenario: historyFailedFree,
		Status:       entities.RequestStatus("FAILED"),
		ErrorMessage: historyFailureMessage,
		Timestamp: entities.Timestamp{
			CreatedAt: created,
			UpdatedAt: created,
		},
	}

	// No GherkinScenario is created for FAILED.
}

// ------------------------------------
// FIXTURE: OTHER USER
// ------------------------------------

func (s *historyBDDStore) seedOther() {
	created := time.Date(
		2026, 11, 11, 11, 0, 0, 0, time.UTC,
	)

	s.requests[historyOtherReqID] = entities.GenerationRequest{
		ID:        historyOtherReqID,
		UserID:    historyOtherID,
		UserStory: historyOtherStory,
		GuiElements: []string{
			"Search field",
			"Search button",
		},
		FreeScenario: "User searches for a product.",
		Status:       entities.RequestStatus("SUCCESS"),
		Timestamp: entities.Timestamp{
			CreatedAt: created,
			UpdatedAt: created,
		},
	}
}

// ------------------------------------
// LIST OWNED
// ------------------------------------

func (s *historyBDDStore) ListOwned(
	ctx context.Context,
	owner uuid.UUID,
	page int,
	pageSize int,
) ([]entities.GenerationRequest, int64, error) {

	s.listCalls++

	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}

	rows := make(
		[]entities.GenerationRequest,
		0,
	)

	// Only include requests belonging to the owner.
	for _, request := range s.requests {
		if request.UserID == owner &&
			string(request.Status) == "SUCCESS" {
			rows = append(
				rows,
				cloneHistoryRequest(request),
			)
		}
	}

	// Sort by newest request first.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].CreatedAt.Equal(rows[j].CreatedAt) {
			return rows[i].ID.String() > rows[j].ID.String()
		}

		return rows[i].CreatedAt.After(rows[j].CreatedAt)
	})

	total := int64(len(rows))

	start := (page - 1) * pageSize

	if start >= len(rows) {
		return []entities.GenerationRequest{}, total, nil
	}

	end := start + pageSize
	if end > len(rows) {
		end = len(rows)
	}

	return rows[start:end], total, nil
}

// ------------------------------------
// GET OWNED DETAIL
// ------------------------------------

func (s *historyBDDStore) GetOwned(
	ctx context.Context,
	owner uuid.UUID,
	requestID uuid.UUID,
) (
	entities.GenerationRequest,
	*entities.GherkinScenario,
	error,
) {

	s.detailCalls++

	if err := ctx.Err(); err != nil {
		return entities.GenerationRequest{}, nil, err
	}

	request, ok := s.requests[requestID]

	if !ok ||
		request.UserID != owner ||
		string(request.Status) != "SUCCESS" {
		return entities.GenerationRequest{},
			nil,
			gorm.ErrRecordNotFound
	}

	// Find the newest scenario belonging to this request.
	var latest *entities.GherkinScenario

	for _, scenario := range s.scenarios {
		if scenario.GenerationRequestID == nil ||
			*scenario.GenerationRequestID != requestID {
			continue
		}

		copied := cloneHistoryScenario(scenario)

		if latest == nil ||
			copied.CreatedAt.After(latest.CreatedAt) ||
			(copied.CreatedAt.Equal(latest.CreatedAt) &&
				copied.ID.String() > latest.ID.String()) {
			latest = &copied
		}
	}

	// latest remains nil when generation has no result.
	return cloneHistoryRequest(request), latest, nil
}

// ------------------------------------
// HTTP TEST ROUTER
// ------------------------------------

func historyBDDRouter(
	store *historyBDDStore,
	users historyBDDUsers,
	jwt authService.JWTService,
) *gin.Engine {

	// Real HistoryService.
	svc := service.NewHistoryService(
		users,
		store,
	)

	// Real HistoryHandler.
	h := handler.NewHistoryHandler(svc)

	router := gin.New()

	// Real JWT authentication middleware.
	group := router.Group(
		"/api/v1/history",
		middlewares.Authenticate(jwt),
	)

	group.GET("", h.List)
	group.GET("/:id", h.Detail)

	return router
}
