package service_test

import (
	"bytes"
	"encoding/csv"
	"errors"
	"math"
	"reflect"
	"testing"
	"unicode/utf8"

	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/service"
)

func TestEvaluationReportCSVRoundTrip(t *testing.T) {
	golden := "Feature: Login\n" +
		"  Scenario: Valid login\n" +
		"    Given the user enters \"a,b\"\n" +
		"    When the user clicks Login\n" +
		"    Then the dashboard appears"

	report := evaluationCSVTestReport(
		dto.EvaluationReportCase{
			CaseID:          "GC-001",
			Title:           "Login, \"admin\" — café",
			ExpectedGherkin: golden,
			LLM: &dto.EvaluationMethodResult{
				Status:             "FAIL",
				GUICoveragePercent: evaluationCSVTestPercent(0),
			},
			RuleBased: &dto.EvaluationMethodResult{
				Status:             "WARN",
				GUICoveragePercent: evaluationCSVTestPercent(74.25),
			},
		},
		dto.EvaluationReportCase{
			CaseID:          "GC-002",
			Title:           "Pending method result",
			ExpectedGherkin: golden,
			RuleBased: &dto.EvaluationMethodResult{
				Status: "PASS",
			},
		},
	)

	data, err := service.BuildEvaluationReportCSV(report)
	if err != nil {
		t.Fatal(err)
	}

	if !utf8.Valid(data) {
		t.Fatal("CSV must contain valid UTF-8")
	}

	got := readEvaluationCSVTest(t, data)

	want := [][]string{
		{
			"case_id",
			"title",
			"expected_gherkin",
			"llm_status",
			"llm_gui_coverage_percent",
			"rule_based_status",
			"rule_based_gui_coverage_percent",
		},
		{
			"GC-001",
			"Login, \"admin\" — café",
			golden,
			"FAIL",
			"0",
			"WARN",
			"74.25",
		},
		{
			"GC-002",
			"Pending method result",
			golden,
			"",
			"",
			"PASS",
			"",
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CSV records = %#v, want %#v", got, want)
	}
}

func TestEvaluationReportCSVSpreadsheetText(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"equals", "=1+1", "\t=1+1"},
		{"plus", "+SUM(1,2)", "\t+SUM(1,2)"},
		{"minus", "-1+2", "\t-1+2"},
		{"at", "@SUM(1,2)", "\t@SUM(1,2)"},
		{"leading whitespace", " \t=1+1", "\t \t=1+1"},
		{"leading newline", "\n=1+1", "\t\n=1+1"},
		{"leading carriage return", "\r=1+1", "\t\r=1+1"},
		{"leading BOM", "\uFEFF=1+1", "\t\uFEFF=1+1"},
		{"fullwidth equals", "＝1+1", "\t＝1+1"},
		{"fullwidth plus", "＋1+1", "\t＋1+1"},
		{"fullwidth minus", "－1+1", "\t－1+1"},
		{"fullwidth at", "＠SUM(1,2)", "\t＠SUM(1,2)"},
		{"ordinary text", "Successful login", "Successful login"},
		{"internal equals", "Expected value = 1", "Expected value = 1"},
		{"leading spaces", "  Login", "  Login"},
		{
			"multiline text",
			"Feature: Login\n  Scenario: Success",
			"Feature: Login\n  Scenario: Success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := evaluationCSVTestReport(dto.EvaluationReportCase{
				CaseID:          tt.text,
				Title:           tt.text,
				ExpectedGherkin: tt.text,
			})

			data, err := service.BuildEvaluationReportCSV(report)
			if err != nil {
				t.Fatal(err)
			}

			records := readEvaluationCSVTest(t, data)
			if len(records) != 2 {
				t.Fatalf("record count = %d, want 2", len(records))
			}

			for column := 0; column < 3; column++ {
				if records[1][column] != tt.want {
					t.Errorf(
						"column %d = %q, want %q",
						column,
						records[1][column],
						tt.want,
					)
				}
			}
		})
	}
}

func TestEvaluationReportCSVRejectsUnfinishedRuns(t *testing.T) {
	for _, status := range []string{"QUEUED", "RUNNING", "FAILED"} {
		t.Run(status, func(t *testing.T) {
			report := evaluationCSVTestReport()
			report.Status = status

			data, err := service.BuildEvaluationReportCSV(report)

			if !errors.Is(err, service.ErrEvaluationReportNotReady) {
				t.Fatalf(
					"error = %v, want ErrEvaluationReportNotReady",
					err,
				)
			}

			if len(data) != 0 {
				t.Fatal("an unfinished run must not produce CSV data")
			}
		})
	}
}

func TestEvaluationReportCSVRejectsInvalidMethodResults(t *testing.T) {
	tests := []struct {
		name   string
		result dto.EvaluationMethodResult
	}{
		{
			"unknown status",
			dto.EvaluationMethodResult{Status: "UNKNOWN"},
		},
		{
			"negative coverage",
			dto.EvaluationMethodResult{
				Status:             "PASS",
				GUICoveragePercent: evaluationCSVTestPercent(-1),
			},
		},
		{
			"coverage above 100",
			dto.EvaluationMethodResult{
				Status:             "PASS",
				GUICoveragePercent: evaluationCSVTestPercent(101),
			},
		},
		{
			"NaN",
			dto.EvaluationMethodResult{
				Status:             "PASS",
				GUICoveragePercent: evaluationCSVTestPercent(math.NaN()),
			},
		},
		{
			"positive infinity",
			dto.EvaluationMethodResult{
				Status:             "PASS",
				GUICoveragePercent: evaluationCSVTestPercent(math.Inf(1)),
			},
		},
		{
			"negative infinity",
			dto.EvaluationMethodResult{
				Status:             "PASS",
				GUICoveragePercent: evaluationCSVTestPercent(math.Inf(-1)),
			},
		},
	}

	for _, method := range []string{"llm", "rule_based"} {
		t.Run(method, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					item := dto.EvaluationReportCase{
						CaseID: "GC-001",
					}

					if method == "llm" {
						item.LLM = &tt.result
					} else {
						item.RuleBased = &tt.result
					}

					report := evaluationCSVTestReport(item)
					data, err := service.BuildEvaluationReportCSV(report)

					if err == nil {
						t.Fatal("expected an error for an invalid method result")
					}

					if len(data) != 0 {
						t.Fatal("an invalid report must not expose partial CSV data")
					}
				})
			}
		})
	}
}

func TestEvaluationReportCSVRejectsInvalidUTF8(t *testing.T) {
	for _, field := range []string{"case_id", "title", "expected_gherkin"} {
		t.Run(field, func(t *testing.T) {
			item := dto.EvaluationReportCase{
				CaseID: "GC-001",
				Title:  "Login",
			}

			invalid := string([]byte{0xff})

			switch field {
			case "case_id":
				item.CaseID = invalid
			case "title":
				item.Title = invalid
			case "expected_gherkin":
				item.ExpectedGherkin = invalid
			}

			report := evaluationCSVTestReport(item)
			data, err := service.BuildEvaluationReportCSV(report)

			if err == nil || len(data) != 0 {
				t.Fatal(
					"invalid UTF-8 must be rejected without returning CSV data",
				)
			}
		})
	}
}

func TestEvaluationReportCSVEmptyCases(t *testing.T) {
	data, err := service.BuildEvaluationReportCSV(evaluationCSVTestReport())
	if err != nil {
		t.Fatal(err)
	}

	records := readEvaluationCSVTest(t, data)

	if len(records) != 1 ||
		len(records[0]) != 7 ||
		records[0][0] != "case_id" {
		t.Fatalf(
			"expected only the seven-column header, got %#v",
			records,
		)
	}
}

func evaluationCSVTestReport(
	cases ...dto.EvaluationReportCase,
) dto.EvaluationReportDetailResponse {
	return dto.EvaluationReportDetailResponse{
		EvaluationRunSummary: dto.EvaluationRunSummary{
			Status:     "COMPLETED",
			TotalCases: len(cases),
		},
		Cases: cases,
	}
}

func evaluationCSVTestPercent(value float64) *float64 {
	return &value
}

func readEvaluationCSVTest(t *testing.T, data []byte) [][]string {
	t.Helper()

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = 7

	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("cannot read exported CSV: %v", err)
	}

	return records
}