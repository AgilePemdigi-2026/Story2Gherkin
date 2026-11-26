//go:build browser

package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExecuteStepsKeepsBrowserAcrossLoginSteps(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /login", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, `<form onsubmit="event.preventDefault(); if (this.email.value === 'qa@example.com' && this.password.value === 'test') this.replaceWith('Dashboard')">
			<input name="email"><input name="password" type="password"><button>Login</button>
		</form>`)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	steps, err := parseSteps([]string{
		`I open "/login"`,
		`I fill "Email" with "<email>"`,
		`I fill "Password" with "<password>"`,
		`I click "Login"`,
		`I see "Dashboard"`,
	}, map[string]string{"email": "qa@example.com", "password": "test"})
	if err != nil {
		t.Fatal(err)
	}
	results, message := executeSteps(t.Context(), server.URL, steps)
	if len(results) != len(steps) {
		t.Fatalf("got %d step results, want %d: %s", len(results), len(steps), message)
	}
	for _, result := range results {
		if result.Status != "PASS" {
			t.Fatalf("step %d (%s): %s: %s", result.Index, result.Text, result.Status, result.Message)
		}
	}
}
