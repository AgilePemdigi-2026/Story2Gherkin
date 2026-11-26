package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
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

	parsed := parseFeature(scenario.GherkinText)
	if len(parsed) == 0 {
		return dto.BehatRunResponse{}, ErrBehatInvalid
	}
	target := parsed[0]
	if req.Scenario != "" {
		found := false
		for _, item := range parsed {
			if item.Name == req.Scenario {
				target, found = item, true
				break
			}
		}
		if !found {
			return dto.BehatRunResponse{}, ErrBehatInvalid
		}
	}

	steps, err := parseSteps(target.Steps, req.Values)
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
	results, message := executeSteps(context.Background(), base, steps)
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

type behatScenario struct {
	Name  string
	Steps []string
}

// parseFeature extracts scenarios and their step text from saved Gherkin.
func parseFeature(gherkin string) []behatScenario {
	scenarios := []behatScenario{}
	current := -1
	for _, line := range strings.Split(gherkin, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trimmed), "scenario:") {
			scenarios = append(scenarios, behatScenario{
				Name: strings.TrimSpace(trimmed[len("scenario:"):]),
			})
			current = len(scenarios) - 1
			continue
		}
		if current < 0 {
			continue
		}
		for _, keyword := range []string{"Given ", "When ", "Then ", "And ", "But "} {
			if strings.HasPrefix(trimmed, keyword) {
				scenarios[current].Steps = append(
					scenarios[current].Steps,
					strings.TrimSpace(trimmed[len(keyword):]),
				)
				break
			}
		}
	}
	return scenarios
}

type behatStep struct {
	Text   string
	Action string
	Target string
	Value  string
}

var behatStepPatterns = []struct {
	action  string
	pattern *regexp.Regexp
}{
	{"open", regexp.MustCompile(`^I open "([^"]+)"$`)},
	{"fill", regexp.MustCompile(`^I fill "([^"]+)" with "([^"]*)"$`)},
	{"click", regexp.MustCompile(`^I click "([^"]+)"$`)},
	{"assert_text", regexp.MustCompile(`^I see "([^"]+)"$`)},
	{"assert_url", regexp.MustCompile(`^the URL contains "([^"]+)"$`)},
}

var placeholderPattern = regexp.MustCompile(`<([A-Za-z0-9_]+)>`)

func substitute(text string, values map[string]string) (string, error) {
	missing := ""
	result := placeholderPattern.ReplaceAllStringFunc(text, func(match string) string {
		key := match[1 : len(match)-1]
		value, ok := values[key]
		if !ok {
			missing = key
			return match
		}
		return value
	})
	if missing != "" {
		return "", fmt.Errorf("missing value for <%s>", missing)
	}
	return result, nil
}

func parseSteps(texts []string, values map[string]string) ([]behatStep, error) {
	steps := make([]behatStep, 0, len(texts))
	for _, original := range texts {
		resolved, err := substitute(original, values)
		if err != nil {
			return nil, err
		}
		matched := false
		for _, candidate := range behatStepPatterns {
			if match := candidate.pattern.FindStringSubmatch(resolved); match != nil {
				step := behatStep{Text: original, Action: candidate.action, Target: match[1]}
				if len(match) > 2 {
					step.Value = match[2]
				}
				steps = append(steps, step)
				matched = true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("unsupported step: %s", original)
		}
	}
	return steps, nil
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
	steps []behatStep,
) ([]entities.BehatStepResult, string) {
	ctx, cancel := context.WithTimeout(parent, behatTotalTimeout)
	defer cancel()

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.CombinedOutput(os.Stderr),
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
			log.Printf("behat step %d failed: %v", index+1, err)
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

func runStep(ctx context.Context, base string, step behatStep) error {
	switch step.Action {
	case "open":
		return chromedp.Do(ctx, chromedp.Navigate(base+step.Target))

	case "fill":
		selector := inputXPath(step.Target)
		return chromedp.Do(ctx,
			chromedp.WaitVisible(selector),
			chromedp.SendKeys(selector, step.Value),
		)

	case "click":
		selector := clickXPath(step.Target)
		return chromedp.Do(ctx,
			chromedp.WaitVisible(selector),
			chromedp.Click(selector),
		)

	case "assert_text":
		body, err := chromedp.Run(ctx, chromedp.Text("body"))
		if err != nil {
			return err
		}
		if !strings.Contains(body, step.Target) {
			return assertionError{fmt.Sprintf("text %q not found on the page", step.Target)}
		}
		return nil

	case "assert_url":
		location, err := chromedp.Run(ctx, chromedp.Location())
		if err != nil {
			return err
		}
		if !strings.Contains(location, step.Target) {
			return assertionError{fmt.Sprintf("url %q does not contain %q", location, step.Target)}
		}
		return nil
	}

	return assertionError{"unsupported action: " + step.Action}
}

const asciiUpper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const asciiLower = "abcdefghijklmnopqrstuvwxyz"

var labelPrefixes = []string{
	"tombol ", "kolom ", "field ", "button ", "input ", "link ", "tautan ", "menu ",
}

// cleanLabel drops common UI prefixes so "Tombol Masuk" matches the text "Masuk".
func cleanLabel(label string) string {
	lower := strings.ToLower(label)
	for _, prefix := range labelPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSpace(label[len(prefix):])
		}
	}
	return label
}

func lowerContains(expr, value string) string {
	return fmt.Sprintf("contains(translate(%s,%q,%q),%q)", expr, asciiUpper, asciiLower, strings.ToLower(value))
}

func inputXPath(label string) string {
	value := cleanLabel(label)
	return "//input[" + strings.Join([]string{
		lowerContains("@placeholder", value),
		lowerContains("@name", value),
		lowerContains("@id", value),
		lowerContains("@aria-label", value),
		fmt.Sprintf("@id = //label[%s]/@for", lowerContains(".", value)),
	}, " or ") + "]"
}

func clickXPath(label string) string {
	value := cleanLabel(label)
	text := lowerContains(".", value)
	return "//button[" + text + "] | //a[" + text + "] | //input[" +
		lowerContains("@value", value) + "] | //*[@role='button' and " + text + "]"
}
