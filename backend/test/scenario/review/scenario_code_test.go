package review_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/story2gherkin/database"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func scenarioCodeDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to a disposable PostgreSQL database")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	name := "scenario_code_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := admin.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		t.Fatal(err)
	}
	if err := admin.Exec("CREATE SCHEMA " + name).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec("DROP SCHEMA " + name + " CASCADE"); connection.Close() })
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		parsed, err := url.Parse(dsn)
		if err != nil {
			t.Fatal(err)
		}
		query := parsed.Query()
		query.Set("search_path", name+",public")
		parsed.RawQuery = query.Encode()
		dsn = parsed.String()
	} else {
		dsn += " search_path=" + name + ",public"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(12)
	t.Cleanup(func() { pool.Close() })
	return db
}

func TestScenarioCodesAreReturnedAndRemainStableAfterCorrection(t *testing.T) {
	svc, store, _, _ := fixture()
	for number, want := range map[int64]string{1: "SCN-0001", 42: "SCN-0042", 9999: "SCN-9999", 10000: "SCN-10000"} {
		store.row.ScenarioNumber = number
		items, err := svc.List(context.Background(), store.owner.String())
		if err != nil || len(items) != 1 {
			t.Fatalf("list: %v", err)
		}
		if items[0].ScenarioCode != want || items[0].ID != store.row.ID.String() {
			t.Fatalf("wrong code or UUID: %+v", items[0])
		}
		loaded, err := svc.Detail(context.Background(), store.owner.String(), store.row.ID.String())
		if err != nil || loaded.ScenarioCode != want {
			t.Fatalf("detail lost code: %+v %v", loaded, err)
		}
		saved, err := svc.Save(context.Background(), store.owner.String(), store.row.ID.String(), dto.SaveCorrectionRequest{Gherkin: original, UpdatedAt: store.row.UpdatedAt})
		if err != nil || saved.ScenarioCode != want || store.row.ScenarioNumber != number {
			t.Fatalf("correction changed code: %+v %v", saved, err)
		}
	}
}

func TestHTTPScenarioCodesKeepUUIDForLookup(t *testing.T) {
	svc, store, _, _ := fixture()
	for _, path := range []string{"/scenarios", "/scenarios/" + store.row.ID.String()} {
		response := httptest.NewRecorder()
		routerFor(svc, store.owner.String()).ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 200 || !strings.Contains(response.Body.String(), `"scenario_code":"SCN-0001"`) || !strings.Contains(response.Body.String(), store.row.ID.String()) {
			t.Fatalf("missing human code or UUID: %d %s", response.Code, response.Body.String())
		}
	}
}

func TestScenarioNumberMigrationBackfillsExistingRowsAndDoesNotRenumber(t *testing.T) {
	db := scenarioCodeDatabase(t)
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	// Reconstruct the old schema without the new column, inside this test schema.
	if err := db.Migrator().DropColumn(&entities.GherkinScenario{}, "ScenarioNumber"); err != nil {
		t.Fatal(err)
	}
	ids := []uuid.UUID{uuid.New(), uuid.New()}
	stamp := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	for _, id := range ids {
		if err := db.Exec(`INSERT INTO gherkin_scenarios (id, feature_name, gherkin_text, source_type, status, created_at, updated_at) VALUES (?, ?, ?, 'LLM', 'VALID', ?, ?)`, id, "Original title", original, stamp, stamp).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	var rows []entities.GherkinScenario
	if err := db.Order("scenario_number").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].ScenarioNumber <= 0 || rows[1].ScenarioNumber <= rows[0].ScenarioNumber {
		t.Fatalf("legacy rows were not numbered: %+v", rows)
	}
	assigned := map[uuid.UUID]int64{}
	for _, row := range rows {
		assigned[row.ID] = row.ScenarioNumber
		if row.GherkinText != original || row.FeatureName != "Original title" || !row.CreatedAt.Equal(stamp) || !row.UpdatedAt.Equal(stamp) {
			t.Fatal("migration changed original content/timestamps")
		}
	}
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		if assigned[row.ID] != row.ScenarioNumber {
			t.Fatal("repeated migration changed assigned numbers")
		}
	}
	repo := repository.NewGherkinScenarioRepository(db)
	created, err := repo.Create(context.Background(), nil, entities.GherkinScenario{ID: uuid.New(), FeatureName: "New", GherkinText: original, SourceType: dto.SourceLLM, Status: dto.ScenarioValid})
	if err != nil || created.ScenarioNumber <= 2 {
		t.Fatalf("new code does not continue after backfill: %+v %v", created, err)
	}
	last := created.ScenarioNumber
	if err := db.Delete(&created).Error; err != nil {
		t.Fatal(err)
	}
	created, err = repo.Create(context.Background(), nil, entities.GherkinScenario{ID: uuid.New(), FeatureName: "After delete", GherkinText: original, SourceType: dto.SourceLLM, Status: dto.ScenarioValid})
	if err != nil || created.ScenarioNumber <= last {
		t.Fatalf("deleted number reused: %+v %v", created, err)
	}
}

func TestScenarioNumbersAreUniqueForConcurrentCreatesAndSelectedInOwnedList(t *testing.T) {
	db := scenarioCodeDatabase(t)
	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	owner := entities.User{ID: uuid.New(), Name: "Analyst", Email: "scenario-codes@example.test", Password: "test-only", Role: entities.RoleSystemAnalyst}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	request := entities.GenerationRequest{ID: uuid.New(), UserID: owner.ID, UserStory: "Test", GuiElements: []string{"Login button"}, FreeScenario: "Test", Status: dto.RequestSuccess}
	if err := db.Create(&request).Error; err != nil {
		t.Fatal(err)
	}
	repo := repository.NewGherkinScenarioRepository(db)
	const count = 20
	numbers, failures := make(chan int64, count), make(chan error, count)
	var workers sync.WaitGroup
	for range count {
		workers.Add(1)
		go func() {
			defer workers.Done()
			row, err := repo.Create(context.Background(), nil, entities.GherkinScenario{ID: uuid.New(), GenerationRequestID: &request.ID, FeatureName: "Same title", GherkinText: original, SourceType: dto.SourceLLM, Status: dto.ScenarioValid})
			if err != nil {
				failures <- err
				return
			}
			numbers <- row.ScenarioNumber
		}()
	}
	workers.Wait()
	close(numbers)
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	seen := map[int64]bool{}
	for number := range numbers {
		if number <= 0 || seen[number] {
			t.Fatalf("non-unique scenario number: %d", number)
		}
		seen[number] = true
	}
	if len(seen) != count {
		t.Fatalf("expected %d numbers, got %d", count, len(seen))
	}
	review := repository.NewReviewRepository(db)
	listed, err := review.ListOwned(context.Background(), owner.ID)
	if err != nil || len(listed) != count {
		t.Fatalf("owned list: %v", err)
	}
	for _, row := range listed {
		if !seen[row.ScenarioNumber] {
			t.Fatal("owned list omitted assigned number")
		}
	}
	row := listed[0]
	want := row.ScenarioNumber
	row.FeatureName = "Renamed feature"
	row.ScenarioNumber = 999999
	saved, err := review.SaveCorrectionOwned(context.Background(), owner.ID, row, row.UpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := review.GetOwned(context.Background(), owner.ID, row.ID)
	if err != nil || loaded.ScenarioNumber != want || loaded.FeatureName != saved.FeatureName {
		t.Fatalf("rename changed stored code: %+v %v", loaded, err)
	}
}
