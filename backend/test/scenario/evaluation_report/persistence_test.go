package evaluation_report_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const evaluationPersistenceGolden = `Feature: Login
  Scenario: Successful login
    Given the login page is open
    When the user submits valid credentials
    Then the dashboard appears`

func TestEvaluationReportPostgresList(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	repo := repository.NewEvaluationReportRepository(db)
	ctx := context.Background()

	runs, total, err := repo.List(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if runs == nil || len(runs) != 0 || total != 0 {
		t.Fatalf("expected an empty list, got runs=%v total=%d", runs, total)
	}

	base := time.Date(2026, 11, 17, 2, 0, 0, 0, time.UTC)
	firstCreator := uuid.NewString()
	secondCreator := uuid.NewString()

	rows := []entities.EvaluationRun{
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			DatasetName: "Older run",
			CreatedBy:   firstCreator,
			TotalCases:  1,
			Status:      entities.EvaluationRunQueued,
			Timestamp:   entities.Timestamp{CreatedAt: base},
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			DatasetName: "Newer run with lower ID",
			CreatedBy:   secondCreator,
			TotalCases:  2,
			Status:      entities.EvaluationRunRunning,
			Timestamp:   entities.Timestamp{CreatedAt: base.Add(time.Hour)},
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			DatasetName: "Newer run with higher ID",
			CreatedBy:   firstCreator,
			TotalCases:  3,
			Status:      entities.EvaluationRunFailed,
			Timestamp:   entities.Timestamp{CreatedAt: base.Add(time.Hour)},
		},
	}

	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}

	runs, total, err = repo.List(ctx, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(runs) != 2 {
		t.Fatalf("first page: total=%d count=%d", total, len(runs))
	}
	if runs[0].RunID != rows[2].ID.String() ||
		runs[1].RunID != rows[1].ID.String() {
		t.Fatal("runs must be ordered by created_at DESC, then ID DESC")
	}

	runs, total, err = repo.List(ctx, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(runs) != 1 || runs[0].RunID != rows[0].ID.String() {
		t.Fatalf("unexpected second page: runs=%v total=%d", runs, total)
	}

	runs, total, err = repo.List(ctx, 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	if runs == nil || len(runs) != 0 || total != 3 {
		t.Fatal("a page beyond the results must return an empty array and the total")
	}
}

func TestEvaluationReportPostgresDetail(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	repo := repository.NewEvaluationReportRepository(db)
	ctx := context.Background()

	run := evaluationPersistenceRun(entities.EvaluationRunCompleted)
	zero, hundred, fractional := 0.0, 100.0, 55.5

	run.Metrics = &entities.EvaluationRunMetrics{
		LLM: entities.EvaluationMethodMetrics{
			ValidationPercent: &zero,
		},
		RuleBased: entities.EvaluationMethodMetrics{
			ValidationPercent:  &hundred,
			GUICoveragePercent: &fractional,
		},
	}

	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}

	// Insert in reverse order to verify ordering by Position.
	second := evaluationPersistenceCase(run.ID, "GC-002", 2)
	second.RuleBased = &entities.EvaluationCaseMethodResult{
		Status: entities.EvaluationCasePass,
	}

	first := evaluationPersistenceCase(run.ID, "GC-001", 1)
	first.Title = "Login, \"admin\" — café"
	first.LLM = &entities.EvaluationCaseMethodResult{
		Status:             entities.EvaluationCaseFail,
		GUICoveragePercent: &zero,
	}

	rows := []entities.EvaluationRunCase{second, first}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}

	report, err := repo.GetByID(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}

	if report.RunID != run.ID.String() ||
		report.DatasetName != run.DatasetName ||
		report.TotalCases != run.TotalCases {
		t.Fatal("stored run metadata was not preserved")
	}

	if len(report.Cases) != 2 ||
		report.Cases[0].CaseID != "GC-001" ||
		report.Cases[1].CaseID != "GC-002" {
		t.Fatal("cases must follow their evaluation positions")
	}

	if report.Cases[0].Title != first.Title ||
		report.Cases[0].ExpectedGherkin != evaluationPersistenceGolden {
		t.Fatal("case snapshots were not preserved")
	}

	llm := report.Cases[0].LLM
	if llm == nil || llm.Status != "FAIL" ||
		llm.GUICoveragePercent == nil || *llm.GUICoveragePercent != 0 {
		t.Fatal("a FAIL result with calculated zero coverage must remain available")
	}

	if report.Cases[1].LLM != nil ||
		report.Cases[1].RuleBased == nil ||
		report.Cases[1].RuleBased.GUICoveragePercent != nil {
		t.Fatal("unavailable method results and coverage must remain null")
	}

	if report.Metrics == nil ||
		report.Metrics.LLM.ValidationPercent == nil ||
		*report.Metrics.LLM.ValidationPercent != 0 ||
		report.Metrics.LLM.GUICoveragePercent != nil {
		t.Fatal("summary metrics must distinguish zero from null")
	}

	if report.Metrics.RuleBased.GUICoveragePercent == nil ||
		*report.Metrics.RuleBased.GUICoveragePercent != fractional {
		t.Fatal("fractional summary coverage was not preserved")
	}

	for _, status := range []entities.EvaluationRunStatus{
		entities.EvaluationRunQueued,
		entities.EvaluationRunRunning,
		entities.EvaluationRunFailed,
	} {
		pending := evaluationPersistenceRun(status)

		if status == entities.EvaluationRunFailed {
			message := "Evaluation could not be completed."
			pending.ErrorMessage = &message
		}

		if err := db.Create(&pending).Error; err != nil {
			t.Fatal(err)
		}

		loaded, err := repo.GetByID(ctx, pending.ID)
		if err != nil {
			t.Fatal(err)
		}

		if loaded.Status != string(status) ||
			loaded.Metrics != nil ||
			loaded.Cases == nil ||
			len(loaded.Cases) != 0 ||
			loaded.TotalCases != pending.TotalCases {
			t.Fatal("unfinished and failed runs must retain metadata and empty results")
		}

		if status == entities.EvaluationRunFailed &&
			(loaded.ErrorMessage == nil ||
				*loaded.ErrorMessage != *pending.ErrorMessage) {
			t.Fatal("the public failure message was not preserved")
		}
	}

	_, err = repo.GetByID(ctx, uuid.New())
	if !errors.Is(err, repository.ErrEvaluationReportNotFound) {
		t.Fatalf("missing run error = %v", err)
	}
}

func TestEvaluationReportPostgresSnapshots(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	repo := repository.NewEvaluationReportRepository(db)
	ctx := context.Background()

	dataset := entities.GoldenDataset{
		ID:        uuid.New(),
		Name:      "Original dataset",
		CreatedBy: uuid.NewString(),
		ItemCount: 1,
	}
	if err := db.Create(&dataset).Error; err != nil {
		t.Fatal(err)
	}

	source := entities.GoldenDatasetItem{
		ID:               uuid.New(),
		GoldenDatasetID:  dataset.ID,
		ItemTitle:        "Original case",
		UserStoryText:    "As a user, I want to log in to access my dashboard.",
		GUIElements:      []string{"Login button"},
		FreeScenarioText: "The user submits valid login credentials.",
		ExpectedGherkin:  evaluationPersistenceGolden,
	}
	if err := db.Create(&source).Error; err != nil {
		t.Fatal(err)
	}

	run := evaluationPersistenceRun(entities.EvaluationRunCompleted)
	run.GoldenDatasetID = &dataset.ID
	run.DatasetName = dataset.Name
	run.TotalCases = 1

	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}

	item := evaluationPersistenceCase(run.ID, "GC-001", 1)
	item.Title = source.ItemTitle
	item.ExpectedGherkin = source.ExpectedGherkin

	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.Model(&dataset).
		Update("name", "Changed dataset").Error; err != nil {
		t.Fatal(err)
	}

	if err := db.Model(&source).Updates(map[string]any{
		"item_title":       "Changed case",
		"expected_gherkin": "Feature: Changed reference",
	}).Error; err != nil {
		t.Fatal(err)
	}

	report, err := repo.GetByID(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}

	if report.DatasetName != "Original dataset" ||
		len(report.Cases) != 1 ||
		report.Cases[0].Title != "Original case" ||
		report.Cases[0].ExpectedGherkin != evaluationPersistenceGolden {
		t.Fatal("source dataset changes must not overwrite report snapshots")
	}

	// Remove only the source records created by this test.
	if err := db.Delete(&source).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(&dataset).Error; err != nil {
		t.Fatal(err)
	}

	var stored entities.EvaluationRun
	if err := db.Where("id = ?", run.ID).Take(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.GoldenDatasetID != nil {
		t.Fatal("deleting the source dataset must clear its optional reference")
	}

	report, err = repo.GetByID(ctx, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if report.DatasetName != "Original dataset" ||
		len(report.Cases) != 1 ||
		report.Cases[0].ExpectedGherkin != evaluationPersistenceGolden {
		t.Fatal("report snapshots must survive deletion of the source dataset")
	}
}

func TestEvaluationReportPostgresConstraintsAndInvalidResults(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	repo := repository.NewEvaluationReportRepository(db)
	ctx := context.Background()

	run := evaluationPersistenceRun(entities.EvaluationRunCompleted)
	if err := db.Create(&run).Error; err != nil {
		t.Fatal(err)
	}

	first := evaluationPersistenceCase(run.ID, "GC-001", 1)
	if err := db.Create(&first).Error; err != nil {
		t.Fatal(err)
	}

	duplicateID := evaluationPersistenceCase(run.ID, "GC-001", 2)
	duplicatePosition := evaluationPersistenceCase(run.ID, "GC-002", 1)
	orphan := evaluationPersistenceCase(uuid.New(), "GC-003", 3)

	for _, trial := range []struct {
		name string
		row  entities.EvaluationRunCase
	}{
		{"duplicate case ID", duplicateID},
		{"duplicate position", duplicatePosition},
		{"missing parent run", orphan},
	} {
		t.Run(trial.name, func(t *testing.T) {
			if err := db.Create(&trial.row).Error; err == nil {
				t.Fatal("expected the database constraint to reject this case")
			}
		})
	}

	otherRun := evaluationPersistenceRun(entities.EvaluationRunCompleted)
	if err := db.Create(&otherRun).Error; err != nil {
		t.Fatal(err)
	}

	otherCase := evaluationPersistenceCase(otherRun.ID, "GC-001", 1)
	if err := db.Create(&otherCase).Error; err != nil {
		t.Fatalf("another run must be allowed to reuse case IDs and positions: %v", err)
	}

	// JSONB accepts this number, but the Report API contract rejects it.
	invalidCoverage := 101.0
	badResult := evaluationPersistenceCase(run.ID, "GC-002", 2)
	badResult.LLM = &entities.EvaluationCaseMethodResult{
		Status:             entities.EvaluationCasePass,
		GUICoveragePercent: &invalidCoverage,
	}

	if err := db.Create(&badResult).Error; err != nil {
		t.Fatal(err)
	}

	_, err := repo.GetByID(ctx, run.ID)
	if err == nil || errors.Is(err, repository.ErrEvaluationReportNotFound) {
		t.Fatal("invalid stored coverage must produce a data error")
	}

	if err := db.Delete(&run).Error; err != nil {
		t.Fatal(err)
	}

	var count int64
	if err := db.Model(&entities.EvaluationRunCase{}).
		Where("run_id = ?", run.ID).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("deleting a run must delete its own cases")
	}

	if err := db.Model(&entities.EvaluationRunCase{}).
		Where("run_id = ?", otherRun.ID).
		Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("deleting one run must not affect another run's cases")
	}
}

func evaluationPersistenceRun(
	status entities.EvaluationRunStatus,
) entities.EvaluationRun {
	return entities.EvaluationRun{
		ID:          uuid.New(),
		DatasetName: "Stored dataset snapshot",
		CreatedBy:   uuid.NewString(),
		TotalCases:  2,
		Status:      status,
	}
}

func evaluationPersistenceCase(
	runID uuid.UUID,
	caseID string,
	position int,
) entities.EvaluationRunCase {
	return entities.EvaluationRunCase{
		ID:              uuid.New(),
		RunID:           runID,
		CaseID:          caseID,
		Position:        position,
		Title:           "Stored case snapshot",
		ExpectedGherkin: evaluationPersistenceGolden,
	}
}

func evaluationPersistenceDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_EVALUATION_REPORT_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("TEST_DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("set TEST_EVALUATION_REPORT_DATABASE_URL or TEST_DATABASE_URL")
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}

	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := pool.Close(); err != nil {
			t.Errorf("close test database connection: %v", err)
		}
	})

	pool.SetMaxOpenConns(1)
	pool.SetMaxIdleConns(1)

	var databaseName string
	if err := db.Raw("SELECT current_database()").
		Scan(&databaseName).Error; err != nil {
		t.Fatal(err)
	}

	if databaseName != "story2gherkin_migration_test" &&
		databaseName != "story2gherkin_test" {
		t.Fatalf(
			"expected a dedicated test database, got %q",
			databaseName,
		)
	}
	schema := "evaluation_report_test_" +
		strings.ReplaceAll(uuid.NewString(), "-", "")

	if err := db.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Exec(
			`DROP SCHEMA IF EXISTS "` + schema + `" CASCADE`,
		).Error; err != nil {
			t.Errorf("cleanup test schema: %v", err)
		}
	})

	if err := db.Exec(
		`SET search_path TO "` + schema + `"`,
	).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(
		&entities.GoldenDataset{},
		&entities.GoldenDatasetItem{},
		&entities.EvaluationRun{},
		&entities.EvaluationRunCase{},
	); err != nil {
		t.Fatal(err)
	}

	return db
}

func TestEvaluationReportPostgresReportCodes(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	repo := repository.NewEvaluationReportRepository(db)
	ctx := context.Background()

	first := evaluationPersistenceRun(entities.EvaluationRunCompleted)
	second := evaluationPersistenceRun(entities.EvaluationRunQueued)

	for _, row := range []*entities.EvaluationRun{&first, &second} {
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}

	if first.RunNumber != 1 || second.RunNumber != 2 {
		t.Fatalf(
			"unexpected generated numbers: first=%d second=%d",
			first.RunNumber,
			second.RunNumber,
		)
	}

	runs, total, err := repo.List(ctx, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(runs) != 2 {
		t.Fatalf("unexpected list: total=%d count=%d", total, len(runs))
	}

	expected := map[string]string{
		first.ID.String():  "ER-0001",
		second.ID.String(): "ER-0002",
	}

	for _, run := range runs {
		code, exists := expected[run.RunID]
		if !exists || run.ReportCode != code {
			t.Fatalf(
				"report code for %s = %q, want %q",
				run.RunID,
				run.ReportCode,
				code,
			)
		}
	}

	before, err := repo.GetByID(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	if before.ReportCode != "ER-0001" {
		t.Fatalf("detail code = %q", before.ReportCode)
	}

	if err := db.Model(&entities.EvaluationRun{}).
		Where("id = ?", first.ID).
		Update("status", entities.EvaluationRunFailed).Error; err != nil {
		t.Fatal(err)
	}

	after, err := repo.GetByID(ctx, first.ID)
	if err != nil {
		t.Fatal(err)
	}

	if after.Status != "FAILED" || after.ReportCode != before.ReportCode {
		t.Fatal("updating run status must preserve the report code")
	}
}

func TestEvaluationReportPostgresRunNumberBackfill(t *testing.T) {
	db := evaluationPersistenceDatabase(t)
	ctx := context.Background()

	// Simulate the earlier schema inside this disposable test schema.
	if err := db.Exec(
		"ALTER TABLE evaluation_runs DROP COLUMN run_number",
	).Error; err != nil {
		t.Fatal(err)
	}

	legacyIDs := []uuid.UUID{uuid.New(), uuid.New()}

	for _, id := range legacyIDs {
		if err := db.Exec(`
			INSERT INTO evaluation_runs (
				id,
				dataset_name,
				created_by,
				total_cases,
				status,
				created_at,
				updated_at
			)
			VALUES (?, ?, ?, ?, ?, NOW(), NOW())
		`,
			id,
			"Legacy dataset snapshot",
			uuid.NewString(),
			1,
			"QUEUED",
		).Error; err != nil {
			t.Fatal(err)
		}
	}

	if err := db.AutoMigrate(&entities.EvaluationRun{}); err != nil {
		t.Fatal(err)
	}

	var legacyRows []entities.EvaluationRun
	if err := db.Find(&legacyRows).Error; err != nil {
		t.Fatal(err)
	}
	if len(legacyRows) != 2 {
		t.Fatal("migration must preserve existing runs")
	}

	expectedIDs := map[uuid.UUID]bool{
		legacyIDs[0]: true,
		legacyIDs[1]: true,
	}

	var highestNumber int64
	seenNumbers := make(map[int64]bool)

	for _, row := range legacyRows {
		if !expectedIDs[row.ID] {
			t.Fatal("migration changed an existing run ID")
		}
		if row.RunNumber <= 0 || seenNumbers[row.RunNumber] {
			t.Fatal("existing runs must receive distinct positive numbers")
		}

		seenNumbers[row.RunNumber] = true

		if row.RunNumber > highestNumber {
			highestNumber = row.RunNumber
		}
	}

	next := evaluationPersistenceRun(entities.EvaluationRunQueued)
	if err := db.Create(&next).Error; err != nil {
		t.Fatal(err)
	}
	if next.RunNumber <= highestNumber {
		t.Fatal("the sequence must continue after backfilled numbers")
	}

	repo := repository.NewEvaluationReportRepository(db)
	runs, total, err := repo.List(ctx, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(runs) != 3 {
		t.Fatal("all existing and new runs must remain readable")
	}

	seenCodes := make(map[string]bool)

	for _, run := range runs {
		if !strings.HasPrefix(run.ReportCode, "ER-") ||
			seenCodes[run.ReportCode] {
			t.Fatal("report codes must be formatted and unique")
		}

		seenCodes[run.ReportCode] = true
	}
}
