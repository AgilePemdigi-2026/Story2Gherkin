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
