package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	scenarioHandler "github.com/story2gherkin/modules/scenario/handler"
	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
)

type serviceStub struct {
	calls int
	err   error
}

func (s *serviceStub) Generate(context.Context, dto.GenerateRequest) (dto.GenerateResponse, error) {
	s.calls++
	return dto.GenerateResponse{Gherkin: "Feature: Example"}, s.err
}

func TestGenerateHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name   string
		body   string
		err    error
		status int
		calls  int
	}{
		{"success", `{}`, nil, 200, 1},
		{"invalid json", `{`, nil, 400, 0},
		{"invalid request", `{}`, service.ErrInvalidRequest, 400, 1},
		{"busy", `{}`, service.ErrUnavailable, 503, 1},
		{"timeout", `{}`, service.ErrTimeout, 504, 1},
		{"bad result", `{}`, service.ErrBadGateway, 502, 1},
		{"internal", `{}`, errors.New("private database details"), 500, 1},
		{"large request", `{"user_story":"` + strings.Repeat("a", 65536) + `"}`, nil, 413, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &serviceStub{err: tc.err}
			h := scenarioHandler.NewScenarioHandler(stub)
			router := gin.New()
			router.POST("/generate", h.Generate)
			req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			if rec.Code != tc.status || stub.calls != tc.calls {
				t.Fatalf("status=%d calls=%d body=%s", rec.Code, stub.calls, rec.Body.String())
			}
			if rec.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("missing cache policy")
			}
			if strings.Contains(rec.Body.String(), "private database details") {
				t.Fatal("internal error exposed")
			}
		})
	}
}
