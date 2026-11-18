package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/story2gherkin/modules/scenario/dto"
)

func TestLLMClientGenerateRuleBased(t *testing.T) {
	var gotPath, gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotKey = r.URL.Path, r.Header.Get("X-AI-Service-Key")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"feature":{"feature_name":"F","scenarios":[{"name":"S","kind":"positive","gui_element_ids":[1],"given":["g"],"when":["w"],"then":["t"]}],"warnings":[]},"gherkin":"Feature: F\n","model":"rule-based","prompt_version":"rule-based-v1","duration_ms":1}`))
	}))
	defer server.Close()

	client, err := NewLLMClient(server.URL, "secret-key-at-least-32-characters-long")
	if err != nil {
		t.Fatalf("NewLLMClient: %v", err)
	}

	res, err := client.GenerateRuleBased(context.Background(), dto.GenerateRequest{
		UserStory:    "story long enough",
		GuiElements:  []string{"Tombol"},
		FreeScenario: "scenario long enough",
	})
	if err != nil {
		t.Fatalf("GenerateRuleBased: %v", err)
	}
	if gotPath != "/internal/scenarios/generate-rule-based" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotKey == "" || res.Gherkin == "" {
		t.Fatalf("missing key or gherkin")
	}
}
