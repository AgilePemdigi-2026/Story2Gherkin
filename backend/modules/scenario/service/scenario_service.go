package service

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
)

var (
	ErrInvalidRequest = errors.New("invalid scenario request")
	ErrUnavailable    = errors.New("scenario generator unavailable")
	ErrTimeout        = errors.New("scenario generation timed out")
	ErrBadGateway     = errors.New("invalid scenario service response")
)

const (
	// MaxAttempts batas retry per request (hasil riset: iterasi perbaikan
	// converge dalam sedikit putaran; tanpa batas membakar kuota LLM).
	MaxAttempts = 3
	// MaxScenarios batas scenario per Feature; fitur kompleks dipecah
	// jadi beberapa request, bukan satu request raksasa.
	MaxScenarios = 6
)

type Generator interface {
	Generate(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error)
}

type ScenarioService interface {
	Generate(ctx context.Context, userID string, req dto.GenerateRequest) (dto.GenerateResponse, error)
}

type scenarioService struct {
	generator Generator
	reqRepo   repository.GenerationRequestRepository
	scenRepo  repository.GherkinScenarioRepository
}

func NewScenarioService(
	generator Generator,
	reqRepo repository.GenerationRequestRepository,
	scenRepo repository.GherkinScenarioRepository,
) ScenarioService {
	return &scenarioService{
		generator: generator,
		reqRepo:   reqRepo,
		scenRepo:  scenRepo,
	}
}

func (s *scenarioService) Generate(ctx context.Context, userID string, req dto.GenerateRequest) (dto.GenerateResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return dto.GenerateResponse{}, ErrInvalidRequest
	}

	req.UserStory = strings.TrimSpace(req.UserStory)
	req.FreeScenario = strings.TrimSpace(req.FreeScenario)
	elements := make([]string, 0, len(req.GuiElements))
	for _, e := range req.GuiElements {
		if e = strings.TrimSpace(e); e != "" {
			elements = append(elements, e)
		}
	}
	req.GuiElements = elements

	if !validInput(req) {
		return dto.GenerateResponse{}, ErrInvalidRequest
	}

	genReq := entities.GenerationRequest{
		ID:           uuid.New(),
		UserID:       uid,
		UserStory:    req.UserStory,
		GuiElements:  req.GuiElements,
		FreeScenario: req.FreeScenario,
		Status:       dto.RequestPending,
	}
	genReq, err = s.reqRepo.Create(ctx, nil, genReq)
	if err != nil {
		return dto.GenerateResponse{}, err
	}
	fail := func(failErr error, msg string) (dto.GenerateResponse, error) {
		genReq.Status = dto.RequestFailed
		genReq.ErrorMessage = msg
		_, _ = s.reqRepo.Update(ctx, nil, genReq)
		return dto.GenerateResponse{}, failErr
	}

	start := time.Now()
	var last dto.GenerateResponse
	var lastErr string
	failErr := ErrBadGateway
	prevGherkin := ""
	gotValid := false

	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		res, err := s.generator.Generate(ctx, req)
		if err != nil {
			if errors.Is(err, ErrTimeout) {
				lastErr, failErr = err.Error(), ErrTimeout
				continue
			}
			if errors.Is(err, ErrUnavailable) {
				lastErr, failErr = err.Error(), ErrUnavailable
				continue
			}
			lastErr = err.Error()
			break
		}
		if attempt > 1 && res.Gherkin != "" && res.Gherkin == prevGherkin {
			// ponytail: output identik = model macet, stop dini hemat kuota.
			lastErr, last = "generator returned identical output", res
			break
		}
		prevGherkin = res.Gherkin
		if msg := validateResponse(req, res); msg != "" {
			lastErr, last = msg, res
			continue
		}
		last, gotValid = res, true
		break
	}

	if !gotValid {
		if last.Gherkin != "" {
			_, _ = s.scenRepo.Create(ctx, nil, entities.GherkinScenario{
				ID:                  uuid.New(),
				GenerationRequestID: &genReq.ID,
				FeatureName:         last.Feature.FeatureName,
				GherkinText:         last.Gherkin,
				SourceType:          dto.SourceLLM,
				Status:              dto.ScenarioInvalid,
				ParserMessage:       lastErr,
				TotalGuiCount:       len(req.GuiElements),
			})
		}
		if lastErr == "" {
			lastErr = failErr.Error()
		}
		return fail(failErr, lastErr)
	}

	covered, uncovered := guiCoverage(req.GuiElements, last)
	if len(uncovered) > 0 {
		warnings := last.Feature.Warnings
		if len(warnings) > 19 {
			warnings = warnings[:19]
		}
		last.Feature.Warnings = append(warnings, "Uncovered GUI elements: "+strings.Join(uncovered, ", "))
	}
	last.DurationMS = time.Since(start).Milliseconds()

	_, err = s.scenRepo.Create(ctx, nil, entities.GherkinScenario{
		ID:                  uuid.New(),
		GenerationRequestID: &genReq.ID,
		FeatureName:         last.Feature.FeatureName,
		GherkinText:         last.Gherkin,
		SourceType:          dto.SourceLLM,
		Status:              dto.ScenarioValid,
		CoveredGuiCount:     len(covered),
		TotalGuiCount:       len(req.GuiElements),
		CoveragePercent:     coveragePercent(len(covered), len(req.GuiElements)),
	})
	if err != nil {
		return fail(err, err.Error())
	}

	genReq.Status = dto.RequestSuccess
	if _, err := s.reqRepo.Update(ctx, nil, genReq); err != nil {
		return dto.GenerateResponse{}, err
	}
	return last, nil
}

func validInput(req dto.GenerateRequest) bool {
	if n := utf8.RuneCountInString(req.UserStory); n < 10 || n > 8000 {
		return false
	}
	if n := utf8.RuneCountInString(req.FreeScenario); n == 0 || n > 4000 {
		return false
	}
	if len(req.GuiElements) == 0 || len(req.GuiElements) > 30 {
		return false
	}
	for _, e := range req.GuiElements {
		if n := utf8.RuneCountInString(e); n == 0 || n > 200 {
			return false
		}
	}
	return true
}

// validateResponse cek struktur hasil generator tanpa parser Gherkin
// (validasi sintaks penuh tetap di AI-service yang punya parser resmi).
func validateResponse(req dto.GenerateRequest, res dto.GenerateResponse) string {
	if strings.TrimSpace(res.Gherkin) == "" {
		return "empty gherkin"
	}
	if strings.TrimSpace(res.Feature.FeatureName) == "" {
		return "empty feature name"
	}
	if len(res.Feature.Scenarios) == 0 || len(res.Feature.Scenarios) > MaxScenarios {
		return "scenario count out of range"
	}
	seen := make(map[string]struct{}, len(res.Feature.Scenarios))
	for _, sc := range res.Feature.Scenarios {
		if strings.TrimSpace(sc.Name) == "" {
			return "empty scenario name"
		}
		key := strings.ToLower(sc.Name)
		if _, ok := seen[key]; ok {
			return "duplicate scenario name: " + sc.Name
		}
		seen[key] = struct{}{}
		if sc.Kind != "positive" && sc.Kind != "negative" && sc.Kind != "edge" {
			return "invalid scenario kind: " + sc.Name
		}
		if len(sc.GuiElementIDs) == 0 {
			return "scenario without gui element reference: " + sc.Name
		}
		for _, id := range sc.GuiElementIDs {
			if id < 1 || id > len(req.GuiElements) {
				return "unknown gui element reference"
			}
		}
		for _, steps := range [][]string{sc.Given, sc.When, sc.Then} {
			if len(steps) == 0 {
				return "scenario missing Given/When/Then: " + sc.Name
			}
			for _, step := range steps {
				if strings.TrimSpace(step) == "" {
					return "empty step in scenario: " + sc.Name
				}
			}
		}
	}
	return ""
}

// guiCoverage implementasi SRS 6.3: normalisasi lalu substring match
// terhadap langkah Given/When/Then (judul Feature/Scenario tidak dihitung).
func guiCoverage(elements []string, res dto.GenerateResponse) (covered, uncovered []string) {
	var steps []string
	for _, sc := range res.Feature.Scenarios {
		steps = append(steps, sc.Given...)
		steps = append(steps, sc.When...)
		steps = append(steps, sc.Then...)
	}
	normSteps := make([]string, 0, len(steps))
	for _, s := range steps {
		normSteps = append(normSteps, normalizeGUIText(s))
	}
	for _, e := range elements {
		hit := false
		ne := normalizeGUIText(e)
		for _, s := range normSteps {
			if ne != "" && strings.Contains(s, ne) {
				hit = true
				break
			}
		}
		if hit {
			covered = append(covered, e)
		} else {
			uncovered = append(uncovered, e)
		}
	}
	return covered, uncovered
}

func normalizeGUIText(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return -1
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

func coveragePercent(covered, total int) float64 {
	if total == 0 {
		return 0
	}
	return math.Round(float64(covered)/float64(total)*100*100) / 100
}
