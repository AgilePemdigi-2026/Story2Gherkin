package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/gorm"
)

var (
	ErrHistoryForbidden         = errors.New("only System Analysts may access history")
	ErrHistoryNotFound          = errors.New("history not found")
	ErrHistoryInvalidPagination = errors.New("invalid history pagination")
)

type HistoryService interface {
	List(
		ctx context.Context,
		userID string,
		page int,
		pageSize int,
	) (dto.HistoryListResponse, error)

	Detail(
		ctx context.Context,
		userID string,
		requestID string,
	) (dto.HistoryDetail, error)
}

type historyService struct {
	users   ReviewUserRepository
	history repository.HistoryRepository
}

func NewHistoryService(
	users ReviewUserRepository,
	history repository.HistoryRepository,
) HistoryService {
	return &historyService{
		users:   users,
		history: history,
	}
}

// Memeriksa role dari database, bukan hanya JWT.
func (s *historyService) authorize(
	ctx context.Context,
	userID string,
) (uuid.UUID, error) {

	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, ErrHistoryForbidden
	}

	user, err := s.users.GetUserById(ctx, nil, id.String())

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, ErrHistoryForbidden
	}

	if err != nil {
		return uuid.Nil, err
	}

	if user.Role != entities.RoleSystemAnalyst {
		return uuid.Nil, ErrHistoryForbidden
	}

	return id, nil
}

// Mengubah entity GenerationRequest menjadi DTO.
func historySummary(
	row entities.GenerationRequest,
) dto.HistorySummary {

	return dto.HistorySummary{
		ID:            row.ID.String(),
		UserStory:     row.UserStory,
		RequestStatus: string(row.Status),
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
}

// Mengambil daftar riwayat Generate milik pengguna.
func (s *historyService) List(
	ctx context.Context,
	userID string,
	page int,
	pageSize int,
) (dto.HistoryListResponse, error) {

	result := dto.HistoryListResponse{
		History:  []dto.HistorySummary{},
		Page:     page,
		PageSize: pageSize,
	}

	owner, err := s.authorize(ctx, userID)
	if err != nil {
		return result, err
	}

	if page < 1 || page > 1000000 ||
		pageSize < 1 || pageSize > 100 {
		return result, ErrHistoryInvalidPagination
	}

	rows, total, err := s.history.ListOwned(
		ctx,
		owner,
		page,
		pageSize,
	)

	if err != nil {
		return result, err
	}

	result.Total = total
	result.TotalPages = int(
		(total + int64(pageSize) - 1) / int64(pageSize),
	)

	for _, row := range rows {
		result.History = append(
			result.History,
			historySummary(row),
		)
	}

	return result, nil
}

// Mengambil detail input dan hasil Generate.
func (s *historyService) Detail(
	ctx context.Context,
	userID string,
	requestID string,
) (dto.HistoryDetail, error) {

	owner, err := s.authorize(ctx, userID)
	if err != nil {
		return dto.HistoryDetail{}, err
	}

	id, err := uuid.Parse(requestID)
	if err != nil {
		return dto.HistoryDetail{}, ErrHistoryNotFound
	}

	request, scenario, err := s.history.GetOwned(
		ctx,
		owner,
		id,
	)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.HistoryDetail{}, ErrHistoryNotFound
	}

	if err != nil {
		return dto.HistoryDetail{}, err
	}

	result := dto.HistoryDetail{
		HistorySummary: historySummary(request),
		GUIElements:    append([]string{}, request.GuiElements...),
		FreeScenario:   request.FreeScenario,
		ErrorMessage:   request.ErrorMessage,
	}

	// Generate gagal bisa tidak memiliki hasil Gherkin.
	if scenario != nil {
		result.Scenario = &dto.HistoryScenario{
			ID:              scenario.ID.String(),
			ScenarioCode:    fmt.Sprintf("SCN-%04d", scenario.ScenarioNumber),
			FeatureName:     scenario.FeatureName,
			Gherkin:         scenario.GherkinText,
			SourceType:      string(scenario.SourceType),
			Status:          string(scenario.Status),
			CoveragePercent: scenario.CoveragePercent,
			UpdatedAt:       scenario.UpdatedAt,
		}
	}

	return result, nil
}
