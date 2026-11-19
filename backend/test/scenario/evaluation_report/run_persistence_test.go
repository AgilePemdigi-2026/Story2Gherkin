package evaluation_report_test

import (
	"context"
	"testing"

	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/repository"
)

func TestEvaluationRunPostgresSavesCaseResults(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	repo := repository.NewEvaluationRunRepository(db)
	ctx := context.Background()
	run := evaluationPersistenceRun(entities.EvaluationRunRunning)
	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}
	item := evaluationPersistenceCase(run.ID, "GC-001", 1)
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	zero, hundred := 0.0, 100.0
	llm := &entities.EvaluationCaseMethodResult{
		Status: entities.EvaluationCaseFail, GUICoveragePercent: &zero,
	}
	rule := &entities.EvaluationCaseMethodResult{
		Status: entities.EvaluationCasePass, GUICoveragePercent: &hundred,
	}
	if err := repo.SaveCaseResult(ctx, item.ID, llm, rule); err != nil {
		t.Fatalf("save case results using production PostgreSQL protocol: %v", err)
	}
	var stored entities.EvaluationRunCase
	if err := db.First(&stored, "id = ?", item.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.Processed || stored.LLM == nil || stored.RuleBased == nil ||
		stored.LLM.Status != llm.Status || stored.RuleBased.Status != rule.Status ||
		stored.LLM.GUICoveragePercent == nil || *stored.LLM.GUICoveragePercent != zero ||
		stored.RuleBased.GUICoveragePercent == nil || *stored.RuleBased.GUICoveragePercent != hundred {
		t.Fatalf("case results not preserved: %+v", stored)
	}
	progress, err := repo.GetProgress(ctx, run.ID)
	if err != nil || progress.ProcessedCases != 1 {
		t.Fatalf("progress = %+v, error = %v", progress, err)
	}

	// Both methods may fail; SQL NULL results still count as processed.
	if err := repo.SaveCaseResult(ctx, item.ID, nil, nil); err != nil {
		t.Fatal(err)
	}
	stored = entities.EvaluationRunCase{}
	if err := db.First(&stored, "id = ?", item.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.Processed || stored.LLM != nil || stored.RuleBased != nil {
		t.Fatalf("unavailable methods must clear previous results: %+v", stored)
	}
	var nullResults bool
	if err := db.Raw("SELECT llm IS NULL AND rule_based IS NULL FROM evaluation_run_cases WHERE id = ?", item.ID).
		Scan(&nullResults).Error; err != nil || !nullResults {
		t.Fatalf("unavailable results must be SQL NULL: null=%v error=%v", nullResults, err)
	}
}

func TestEvaluationRunPostgresFinishesWithMetrics(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	repo := repository.NewEvaluationRunRepository(db)
	ctx := context.Background()
	run := evaluationPersistenceRun(entities.EvaluationRunRunning)
	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}
	zero, hundred, fractional := 0.0, 100.0, 66.67
	metrics := &entities.EvaluationRunMetrics{
		LLM: entities.EvaluationMethodMetrics{ValidationPercent: &zero},
		RuleBased: entities.EvaluationMethodMetrics{
			ValidationPercent: &hundred, GUICoveragePercent: &fractional,
		},
	}
	if err := repo.Finish(ctx, run.ID, entities.EvaluationRunCompleted, metrics, nil); err != nil {
		t.Fatalf("finish run using production PostgreSQL protocol: %v", err)
	}
	var stored entities.EvaluationRun
	if err := db.First(&stored, "id = ?", run.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != entities.EvaluationRunCompleted || stored.Metrics == nil ||
		stored.Metrics.LLM.ValidationPercent == nil || *stored.Metrics.LLM.ValidationPercent != zero ||
		stored.Metrics.LLM.GUICoveragePercent != nil ||
		stored.Metrics.RuleBased.ValidationPercent == nil || *stored.Metrics.RuleBased.ValidationPercent != hundred ||
		stored.Metrics.RuleBased.GUICoveragePercent == nil || *stored.Metrics.RuleBased.GUICoveragePercent != fractional {
		t.Fatalf("metrics not preserved: %+v", stored.Metrics)
	}
	message := "Evaluation failed due to an internal error."
	if err := repo.Finish(ctx, run.ID, entities.EvaluationRunFailed, nil, &message); err != nil {
		t.Fatal(err)
	}
	stored = entities.EvaluationRun{}
	if err := db.First(&stored, "id = ?", run.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != entities.EvaluationRunFailed || stored.Metrics != nil ||
		stored.ErrorMessage == nil || *stored.ErrorMessage != message {
		t.Fatalf("failed run state not preserved: %+v", stored)
	}
}
