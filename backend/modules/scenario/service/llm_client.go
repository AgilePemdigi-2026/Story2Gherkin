package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/story2gherkin/modules/scenario/dto"
)

// LLMClient memanggil AI-service internal untuk generate skenario.
// Namanya mengikuti llm_client.py di ai-service; URL/key menunjuk ke
// AI-service (bukan langsung ke provider) sehingga tetap agnostik
// mau backend-nya Ollama maupun OpenRouter.
type LLMClient struct {
	endpoint string
	key      string
	client   *http.Client
}

func NewLLMClient(baseURL, key string) (*LLMClient, error) {
	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("AI_SERVICE_URL must be an HTTP(S) base URL")
	}
	if len(strings.TrimSpace(key)) < 32 || strings.HasPrefix(key, "REPLACE_") {
		return nil, fmt.Errorf("AI_SERVICE_KEY must contain a random secret of at least 32 characters")
	}
	return &LLMClient{
		endpoint: strings.TrimRight(baseURL, "/") + "/internal/scenarios/generate",
		key:      key,
		client: &http.Client{
			Timeout: 150 * time.Second,
			// Never forward the internal service key to a redirected URL.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}, nil
}

func (c *LLMClient) Generate(ctx context.Context, input dto.GenerateRequest) (dto.GenerateResponse, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return dto.GenerateResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return dto.GenerateResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-AI-Service-Key", c.key)
	res, err := c.client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return dto.GenerateResponse{}, ErrTimeout
		}
		return dto.GenerateResponse{}, ErrUnavailable
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusOK:
	case http.StatusServiceUnavailable:
		return dto.GenerateResponse{}, ErrUnavailable
	case http.StatusGatewayTimeout:
		return dto.GenerateResponse{}, ErrTimeout
	default:
		return dto.GenerateResponse{}, ErrBadGateway
	}
	const maxResponse = 1 << 20
	data, err := io.ReadAll(io.LimitReader(res.Body, maxResponse+1))
	if err != nil || len(data) > maxResponse {
		return dto.GenerateResponse{}, ErrBadGateway
	}
	var result dto.GenerateResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return dto.GenerateResponse{}, ErrBadGateway
	}
	// Light shape check; deep validation lives in scenarioService.
	if strings.TrimSpace(result.Gherkin) == "" || strings.TrimSpace(result.Feature.FeatureName) == "" || len(result.Feature.Scenarios) == 0 {
		return dto.GenerateResponse{}, ErrBadGateway
	}
	return result, nil
}
