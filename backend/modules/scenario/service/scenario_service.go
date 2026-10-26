package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/story2gherkin/modules/scenario/dto"
)

var (
	ErrInvalidRequest = errors.New("invalid scenario request")
	ErrUnavailable    = errors.New("scenario generator unavailable")
	ErrTimeout        = errors.New("scenario generation timed out")
	ErrBadGateway     = errors.New("invalid scenario service response")
)

type Generator interface {
	Generate(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error)
}

type ScenarioService interface {
	Generate(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error)
}

type scenarioService struct{ generator Generator }

func NewScenarioService(generator Generator) ScenarioService {
	return &scenarioService{generator: generator}
}

func (s *scenarioService) Generate(ctx context.Context, req dto.GenerateRequest) (dto.GenerateResponse, error) {
	req.UserStory = strings.TrimSpace(req.UserStory)
	if n := utf8.RuneCountInString(req.UserStory); n < 10 || n > 8000 {
		return dto.GenerateResponse{}, ErrInvalidRequest
	}
	if len(req.AcceptanceCriteria) == 0 || len(req.AcceptanceCriteria) > 20 {
		return dto.GenerateResponse{}, ErrInvalidRequest
	}
	// Copy the slice so normalizing input does not mutate the caller's data.
	req.AcceptanceCriteria = append([]string(nil), req.AcceptanceCriteria...)
	for i, criterion := range req.AcceptanceCriteria {
		criterion = strings.TrimSpace(criterion)
		if n := utf8.RuneCountInString(criterion); n == 0 || n > 500 || strings.ContainsAny(criterion, "\r\n") {
			return dto.GenerateResponse{}, ErrInvalidRequest
		}
		req.AcceptanceCriteria[i] = criterion
	}
	if req.MaxScenarios == 0 {
		req.MaxScenarios = 4
	}
	if req.Language == "" {
		req.Language = "en"
	}
	if req.MaxScenarios < 1 || req.MaxScenarios > 6 || (req.Language != "en" && req.Language != "id") {
		return dto.GenerateResponse{}, ErrInvalidRequest
	}
	return s.generator.Generate(ctx, req)
}
