package service

import (
	"testing"

	"github.com/story2gherkin/database/entities"
)

func TestNormalizeBaseURL(t *testing.T) {
	valid := map[string]string{
		"https://staging.example.com":  "https://staging.example.com",
		"https://staging.example.com/": "https://staging.example.com",
		"http://localhost:3000":        "http://localhost:3000",
	}
	for input, want := range valid {
		got, err := normalizeBaseURL(input)
		if err != nil || got != want {
			t.Errorf("normalizeBaseURL(%q) = %q, %v; want %q", input, got, err, want)
		}
	}

	for _, input := range []string{
		"",
		"staging.example.com",
		"ftp://example.com",
		"https://user:pass@example.com",
		"https://example.com?x=1",
	} {
		if _, err := normalizeBaseURL(input); err == nil {
			t.Errorf("normalizeBaseURL(%q) should fail", input)
		}
	}
}

func TestParseFeature(t *testing.T) {
	gherkin := `Feature: Login

  Scenario: Valid login
    Given I open "/login"
    When I fill "Email" with "<email>"
    And I click "Masuk"
    Then I see "Dashboard"

  Scenario: Wrong password
    Given I open "/login"
    Then I see "Error"
`
	scenarios := parseFeature(gherkin)
	if len(scenarios) != 2 {
		t.Fatalf("scenarios = %d; want 2", len(scenarios))
	}
	if scenarios[0].Name != "Valid login" || len(scenarios[0].Steps) != 4 {
		t.Errorf("first scenario = %+v", scenarios[0])
	}
	if scenarios[1].Name != "Wrong password" || len(scenarios[1].Steps) != 2 {
		t.Errorf("second scenario = %+v", scenarios[1])
	}
}

func TestParseSteps(t *testing.T) {
	steps, err := parseSteps([]string{
		`I open "/login"`,
		`I fill "Email" with "<email>"`,
		`I click "Masuk"`,
		`I see "Dashboard"`,
		`the URL contains "/dashboard"`,
	}, map[string]string{"email": "qa@example.com"})
	if err != nil {
		t.Fatalf("parseSteps() error = %v", err)
	}
	if len(steps) != 5 {
		t.Fatalf("steps = %d; want 5", len(steps))
	}
	if steps[1].Action != "fill" || steps[1].Target != "Email" || steps[1].Value != "qa@example.com" {
		t.Errorf("fill step = %+v", steps[1])
	}
	if steps[1].Text != `I fill "Email" with "<email>"` {
		t.Errorf("original text must be preserved, got %q", steps[1].Text)
	}

	if _, err := parseSteps([]string{`I do something else`}, nil); err == nil {
		t.Error("unsupported step should fail")
	}
	if _, err := parseSteps([]string{`I fill "Email" with "<missing>"`}, nil); err == nil {
		t.Error("missing placeholder value should fail")
	}
}

func TestCleanLabel(t *testing.T) {
	cases := map[string]string{
		"Tombol Masuk": "Masuk",
		"Kolom Email":  "Email",
		"Email":        "Email",
	}
	for input, want := range cases {
		if got := cleanLabel(input); got != want {
			t.Errorf("cleanLabel(%q) = %q; want %q", input, got, want)
		}
	}
}

func TestOverallStatus(t *testing.T) {
	cases := []struct {
		statuses []string
		want     entities.BehatRunStatus
	}{
		{[]string{"PASS", "PASS"}, entities.BehatRunPass},
		{[]string{"PASS", "FAIL"}, entities.BehatRunFail},
		{[]string{"FAIL", "ERROR"}, entities.BehatRunError},
		{[]string{}, entities.BehatRunPass},
	}
	for _, tc := range cases {
		results := make([]entities.BehatStepResult, 0, len(tc.statuses))
		for _, status := range tc.statuses {
			results = append(results, entities.BehatStepResult{Status: status})
		}
		if got := overallStatus(results); got != tc.want {
			t.Errorf("overallStatus(%v) = %q; want %q", tc.statuses, got, tc.want)
		}
	}
}
