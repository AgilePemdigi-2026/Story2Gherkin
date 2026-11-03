package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/story2gherkin/modules/scenario/dto"
)

func (c *LLMClient) Validate(ctx context.Context, text string, elements []string) (dto.ScenarioValidation, error) {
	var result dto.ScenarioValidation
	body, err := json.Marshal(map[string]any{"gherkin": text, "gui_elements": elements})
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	endpoint := strings.TrimSuffix(c.endpoint, "/generate") + "/validate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AI-Service-Key", c.key)
	res, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return result, ErrTimeout
		}
		return result, ErrUnavailable
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return result, ErrBadGateway
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return result, ErrBadGateway
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return result, ErrBadGateway
	}
	// Refuse malformed upstream reports before updating persisted validation.
	if result.Issues == nil || result.UsedElements == nil || result.NotCoveredElements == nil ||
		math.IsNaN(result.CoveragePercent) || result.CoveragePercent < 0 || result.CoveragePercent > 100 ||
		len(result.UsedElements)+len(result.NotCoveredElements) != len(elements) {
		return result, ErrBadGateway
	}
	if result.Valid && (strings.TrimSpace(result.FeatureName) == "" || len(result.Issues) != 0 ||
		!result.Checks.Feature || !result.Checks.Scenario || !result.Checks.Given || !result.Checks.When || !result.Checks.Then) {
		return result, ErrBadGateway
	}
	return result, nil
}
