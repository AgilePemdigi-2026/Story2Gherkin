package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/story2gherkin/modules/scenario/dto"
)

// BuildEvaluationReportCSV exports a completed report.
// The caller must first authorize access through the service.
func BuildEvaluationReportCSV(
	report dto.EvaluationReportDetailResponse,
) ([]byte, error) {
	if report.Status != "COMPLETED" {
		return nil, ErrEvaluationReportNotReady
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	header := []string{
		"case_id",
		"title",
		"expected_gherkin",
		"llm_status",
		"llm_gui_coverage_percent",
		"rule_based_status",
		"rule_based_gui_coverage_percent",
	}

	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("write evaluation CSV header: %w", err)
	}

	for index, item := range report.Cases {
		llmStatus, llmCoverage, err := evaluationCSVMethod(item.LLM)
		if err != nil {
			return nil, fmt.Errorf(
				"export LLM result for case %d: %w",
				index+1,
				err,
			)
		}

		ruleStatus, ruleCoverage, err := evaluationCSVMethod(item.RuleBased)
		if err != nil {
			return nil, fmt.Errorf(
				"export rule-based result for case %d: %w",
				index+1,
				err,
			)
		}

		record := []string{
			item.CaseID,
			item.Title,
			item.ExpectedGherkin,
			llmStatus,
			llmCoverage,
			ruleStatus,
			ruleCoverage,
		}

		for column, value := range record {
			if !utf8.ValidString(value) {
				return nil, fmt.Errorf(
					"invalid UTF-8 in case %d, column %d",
					index+1,
					column+1,
				)
			}

			record[column] = evaluationCSVText(value)
		}

		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf(
				"write evaluation CSV case %d: %w",
				index+1,
				err,
			)
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush evaluation CSV: %w", err)
	}

	return buffer.Bytes(), nil
}

// evaluationCSVMethod preserves unavailable results as empty cells.
func evaluationCSVMethod(
	result *dto.EvaluationMethodResult,
) (string, string, error) {
	if result == nil {
		return "", "", nil
	}

	switch result.Status {
	case "PASS", "WARN", "FAIL":
	default:
		return "", "", fmt.Errorf("invalid evaluation case status")
	}

	coverage, err := evaluationCSVPercent(result.GUICoveragePercent)
	if err != nil {
		return "", "", err
	}

	return result.Status, coverage, nil
}

// evaluationCSVPercent keeps zero distinct from an unavailable percentage.
func evaluationCSVPercent(value *float64) (string, error) {
	if value == nil {
		return "", nil
	}

	number := *value

	if math.IsNaN(number) || math.IsInf(number, 0) ||
		number < 0 || number > 100 {
		return "", fmt.Errorf("GUI coverage must be between 0 and 100")
	}

	if number == 0 {
		return "0", nil
	}

	return strconv.FormatFloat(number, 'f', -1, 64), nil
}

// evaluationCSVText mitigates spreadsheet formula interpretation.
// A leading tab also causes encoding/csv to quote the field.
// Only the exported cell changes; stored report data remains unchanged.
func evaluationCSVText(value string) string {
	candidate := strings.TrimLeftFunc(value, func(char rune) bool {
		return unicode.IsSpace(char) ||
			unicode.IsControl(char) ||
			unicode.Is(unicode.Cf, char)
	})

	for _, char := range candidate {
		switch char {
		case '=', '+', '-', '@', '＝', '＋', '－', '＠':
			return "\t" + value
		}

		break
	}

	return value
}
