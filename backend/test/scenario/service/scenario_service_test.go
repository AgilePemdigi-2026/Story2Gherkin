package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/story2gherkin/modules/scenario/dto"
)

type generatorStub struct {
	calls   int
	request dto.GenerateRequest
	err     error
}

func (s *generatorStub) Generate(_ context.Context, req dto.GenerateRequest) (dto.GenerateResponse, error) {
	s.calls++
	s.request = req
	return dto.GenerateResponse{Gherkin: "Feature: Test"}, s.err
}

func validRequest() dto.GenerateRequest {
	return dto.GenerateRequest{UserStory: " As a user I want to log in. ", AcceptanceCriteria: []string{" Login succeeds. "}}
}

func TestNormalizeAndDefault(t *testing.T) {
	stub := &generatorStub{}
	req := validRequest()
	_, err := NewScenarioService(stub).Generate(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if stub.request.UserStory != strings.TrimSpace(req.UserStory) || stub.request.AcceptanceCriteria[0] != "Login succeeds." || stub.request.MaxScenarios != 4 || stub.request.Language != "en" {
		t.Fatalf("unexpected normalized input: %+v", stub.request)
	}
	if req.AcceptanceCriteria[0] != " Login succeeds. " {
		t.Fatal("caller input mutated")
	}
}

func TestInvalidRequestsDoNotCallModel(t *testing.T) {
	for _, change := range []func(*dto.GenerateRequest){
		func(r *dto.GenerateRequest) { r.UserStory = "short" },
		func(r *dto.GenerateRequest) { r.AcceptanceCriteria = nil },
		func(r *dto.GenerateRequest) { r.AcceptanceCriteria = []string{"Given x\nThen y"} },
		func(r *dto.GenerateRequest) { r.MaxScenarios = 7 },
		func(r *dto.GenerateRequest) { r.Language = "unknown" },
	} {
		req := validRequest()
		change(&req)
		stub := &generatorStub{}
		_, err := NewScenarioService(stub).Generate(context.Background(), req)
		if !errors.Is(err, ErrInvalidRequest) || stub.calls != 0 {
			t.Fatalf("validation failed: calls=%d error=%v", stub.calls, err)
		}
	}
}

func TestAIClientResponses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"success", 200, `{"feature":{"feature_name":"Login","scenarios":[{"name":"Valid"}],"warnings":[]},"gherkin":"Feature: Login\n","model":"fake","prompt_version":"v1","duration_ms":10}`, nil},
		{"busy", 503, "private details", ErrUnavailable},
		{"timeout", 504, "private details", ErrTimeout},
		{"invalid output", 502, "private details", ErrBadGateway},
		{"bad json", 200, "invalid", ErrBadGateway},
		{"empty result", 200, `{}`, ErrBadGateway},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/internal/scenarios/generate" || r.Header.Get("X-AI-Service-Key") != strings.Repeat("x", 32) {
					t.Error("incorrect upstream request")
				}
				var req dto.GenerateRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserStory == "" {
					t.Error("invalid body")
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			client, err := NewAIClient(server.URL, strings.Repeat("x", 32))
			if err != nil {
				t.Fatal(err)
			}
			req := validRequest()
			req.MaxScenarios = 4
			_, err = client.Generate(context.Background(), req)
			if !errors.Is(err, tc.want) {
				t.Fatalf("error=%v want=%v", err, tc.want)
			}
		})
	}
}
