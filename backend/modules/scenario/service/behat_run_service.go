package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/gorm"
)

var (
	ErrBehatForbidden = errors.New("only System Analysts may run end-to-end tests")
	ErrBehatNotFound  = errors.New("scenario not found")
	ErrBehatInvalid   = errors.New("invalid end-to-end test request")
)

const (
	behatStepTimeout  = 10 * time.Second
	behatTotalTimeout = 60 * time.Second
	behatHistoryLimit = 10
)

type BehatRunService interface {
	Run(ctx context.Context, userID, scenarioID string, req dto.BehatRunRequest) (dto.BehatRunResponse, error)
	History(ctx context.Context, userID, scenarioID string) ([]dto.BehatRunSummary, error)
}

type behatRunService struct {
	runs      repository.BehatRunRepository
	scenarios repository.ReviewRepository
	users     ReviewUserRepository
}

func NewBehatRunService(
	runs repository.BehatRunRepository,
	scenarios repository.ReviewRepository,
	users ReviewUserRepository,
) BehatRunService {
	return &behatRunService{runs: runs, scenarios: scenarios, users: users}
}

func (s *behatRunService) authorize(ctx context.Context, userID string) (uuid.UUID, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, ErrBehatForbidden
	}
	user, err := s.users.GetUserById(ctx, nil, id.String())
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, ErrBehatForbidden
	}
	if err != nil {
		return uuid.Nil, err
	}
	if user.Role != entities.RoleSystemAnalyst {
		return uuid.Nil, ErrBehatForbidden
	}
	return id, nil
}

func (s *behatRunService) ownedScenario(
	ctx context.Context,
	owner uuid.UUID,
	scenarioID string,
) (entities.GherkinScenario, error) {
	id, err := uuid.Parse(scenarioID)
	if err != nil {
		return entities.GherkinScenario{}, ErrBehatNotFound
	}
	scenario, err := s.scenarios.GetOwned(ctx, owner, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return entities.GherkinScenario{}, ErrBehatNotFound
	}
	return scenario, err
}

func (s *behatRunService) Run(
	ctx context.Context,
	userID, scenarioID string,
	req dto.BehatRunRequest,
) (dto.BehatRunResponse, error) {
	owner, err := s.authorize(ctx, userID)
	if err != nil {
		return dto.BehatRunResponse{}, err
	}
	scenario, err := s.ownedScenario(ctx, owner, scenarioID)
	if err != nil {
		return dto.BehatRunResponse{}, err
	}
	if scenario.Status != dto.ScenarioValid {
		return dto.BehatRunResponse{}, ErrBehatInvalid
	}
	base, err := normalizeBaseURL(req.BaseURL)
	if err != nil {
		return dto.BehatRunResponse{}, ErrBehatInvalid
	}

	run := &entities.BehatRun{
		ID:         uuid.New(),
		ScenarioID: scenario.ID,
		OwnerID:    owner,
		BaseURL:    base,
		Status:     entities.BehatRunError,
		Steps:      []entities.BehatStepResult{},
	}
	if err := s.runs.Create(ctx, run); err != nil {
		return dto.BehatRunResponse{}, err
	}

	started := time.Now()
	results, message := executeSteps(ctx, base, req.Steps)
	run.Steps = results
	run.Message = message
	run.DurationMS = time.Since(started).Milliseconds()
	run.Status = overallStatus(results)
	_ = s.runs.Update(ctx, run)

	return dto.BehatRunResponse{
		RunID:      run.ID.String(),
		Status:     string(run.Status),
		DurationMS: run.DurationMS,
		Steps:      toStepDTO(results),
		Message:    message,
	}, nil
}

func (s *behatRunService) History(
	ctx context.Context,
	userID, scenarioID string,
) ([]dto.BehatRunSummary, error) {
	owner, err := s.authorize(ctx, userID)
	if err != nil {
		return nil, err
	}
	scenario, err := s.ownedScenario(ctx, owner, scenarioID)
	if err != nil {
		return nil, err
	}
	rows, err := s.runs.ListByScenario(ctx, owner, scenario.ID, behatHistoryLimit)
	if err != nil {
		return nil, err
	}

	summaries := make([]dto.BehatRunSummary, 0, len(rows))
	for _, row := range rows {
		summaries = append(summaries, dto.BehatRunSummary{
			RunID:      row.ID.String(),
			Status:     string(row.Status),
			BaseURL:    row.BaseURL,
			DurationMS: row.DurationMS,
			Steps:      toStepDTO(row.Steps),
			Message:    row.Message,
			CreatedAt:  row.CreatedAt,
		})
	}
	return summaries, nil
}

func normalizeBaseURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") ||
		u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", ErrBehatInvalid
	}
	return strings.TrimRight(u.String(), "/"), nil
}

func overallStatus(results []entities.BehatStepResult) entities.BehatRunStatus {
	status := entities.BehatRunPass
	for _, result := range results {
		switch result.Status {
		case "ERROR":
			return entities.BehatRunError
		case "FAIL":
			status = entities.BehatRunFail
		}
	}
	return status
}

func toStepDTO(results []entities.BehatStepResult) []dto.BehatStepResult {
	steps := make([]dto.BehatStepResult, 0, len(results))
	for _, result := range results {
		steps = append(steps, dto.BehatStepResult{
			Index:   result.Index,
			Text:    result.Text,
			Status:  result.Status,
			Message: result.Message,
		})
	}
	return steps
}

// assertionError marks a failed check, not an infrastructure failure.
type assertionError struct{ message string }

func (e assertionError) Error() string { return e.message }

func executeSteps(
	parent context.Context,
	base string,
	steps []dto.BehatStepInput,
) ([]entities.BehatStepResult, string) {
	ctx, cancel := context.WithTimeout(parent, behatTotalTimeout)
	defer cancel()

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Flag("disable-dev-shm-usage", true),
	)...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	results := make([]entities.BehatStepResult, 0, len(steps))
	for index, step := range steps {
		result := entities.BehatStepResult{Index: index + 1, Text: step.Text, Status: "PASS"}
		stepCtx, cancelStep := context.WithTimeout(browserCtx, behatStepTimeout)
		err := runStep(stepCtx, base, step)
		cancelStep()

		if err != nil {
			result.Status = "ERROR"
			var assertion assertionError
			if errors.As(err, &assertion) {
				result.Status = "FAIL"
			}
			result.Message = err.Error()
			results = append(results, result)
			for rest := index + 1; rest < len(steps); rest++ {
				results = append(results, entities.BehatStepResult{
					Index:  rest + 1,
					Text:   steps[rest].Text,
					Status: "SKIPPED",
				})
			}
			return results, result.Message
		}
		results = append(results, result)
	}
	return results, ""
}

func runStep(
	ctx context.Context,
	base string,
	step dto.BehatStepInput,
) error {
	switch step.Action {
	case "open":
		return chromedp.Run(ctx, chromedp.Navigate(base+step.Value))

	case "fill":
		if step.Selector == "" {
			return assertionError{"selector is required for fill"}
		}
		return chromedp.Run(ctx,
			chromedp.WaitVisible(step.Selector, chromedp.ByQuery),
			chromedp.SendKeys(step.Selector, step.Value, chromedp.ByQuery),
		)

	case "click":
		if step.Selector == "" {
			return assertionError{"selector is required for click"}
		}
		return chromedp.Run(ctx,
			chromedp.WaitVisible(step.Selector, chromedp.ByQuery),
			chromedp.Click(step.Selector, chromedp.ByQuery),
		)

	case "assert_visible":
		if step.Selector == "" {
			return assertionError{"selector is required for assert_visible"}
		}
		if err := chromedp.Run(ctx, chromedp.WaitVisible(step.Selector, chromedp.ByQuery)); err != nil {
			return assertionError{fmt.Sprintf("element %q is not visible", step.Selector)}
		}
		return nil

	case "assert_text":
		var body string
		if err := chromedp.Run(ctx,
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.Text("body", &body, chromedp.ByQuery),
		); err != nil {
			return err
		}
		if !strings.Contains(body, step.Value) {
			return assertionError{fmt.Sprintf("text %q not found on the page", step.Value)}
		}
		return nil

	case "assert_url":
		var location string
		if err := chromedp.Run(ctx, chromedp.Location(&location)); err != nil {
			return err
		}
		if !strings.Contains(location, step.Value) {
			return assertionError{fmt.Sprintf("url %q does not contain %q", location, step.Value)}
		}
		return nil
	}

	return assertionError{"unsupported action: " + step.Action}
}
