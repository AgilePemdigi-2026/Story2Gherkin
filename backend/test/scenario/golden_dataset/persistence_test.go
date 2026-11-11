package golden_dataset_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func goldenTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("TEST_GOLDEN_DATASET_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_GOLDEN_DATASET_DATABASE_URL")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}

	// Prevent accidental testing against the main database.
	var databaseName string
	if err := db.Raw(
		"SELECT current_database()",
	).Scan(&databaseName).Error; err != nil {
		pool.Close()
		t.Fatal(err)
	}

	if databaseName != "story2gherkin_migration_test" {
		pool.Close()
		t.Fatalf(
			"unsafe database: %s",
			databaseName,
		)
	}

	// Keep search_path on the same connection.
	pool.SetMaxOpenConns(1)

	schema := "golden_test_" +
		strings.ReplaceAll(uuid.NewString(), "-", "")

	if err := db.Exec(
		`CREATE SCHEMA "` + schema + `"`,
	).Error; err != nil {
		pool.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Exec("SET search_path TO public")

		if err := db.Exec(
			`DROP SCHEMA IF EXISTS "` + schema + `" CASCADE`,
		).Error; err != nil {
			t.Errorf("cleanup schema: %v", err)
		}

		pool.Close()
	})

	if err := db.Exec(
		`SET search_path TO "` + schema + `", public`,
	).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(
		&entities.GoldenDataset{},
		&entities.GoldenDatasetItem{},
	); err != nil {
		t.Fatal(err)
	}

	return db
}

func goldenTestItems() []entities.GoldenDatasetItem {
	return []entities.GoldenDatasetItem{
		{
			ItemTitle: "Successful Login",
			UserStoryText: "As a user, I want to log in " +
				"to access my dashboard.",
			GUIElements: []string{
				"Email field",
				"Password field",
				"Login button",
			},
			FreeScenarioText: "User logs in with valid credentials.",
			ExpectedGherkin: "Feature: Login\n" +
				"  Scenario: Successful Login\n" +
				"    Given a login page\n" +
				"    When valid credentials are entered\n" +
				"    Then the dashboard is displayed",
		},
		{
			ItemTitle: "Invalid Password",
			UserStoryText: "As a user, I want to see an error " +
				"when my password is incorrect.",
			GUIElements: []string{
				"Password field",
				"Login button",
				"Error message",
			},
			FreeScenarioText: "User enters an invalid password.",
			ExpectedGherkin: "Feature: Login\n" +
				"  Scenario: Invalid Password\n" +
				"    Given a login page\n" +
				"    When an invalid password is entered\n" +
				"    Then an error message is displayed",
		},
	}
}

func TestGoldenDatasetPostgresPersistence(t *testing.T) {
	db := goldenTestDatabase(t)
	repo := repository.NewGoldenDatasetRepository(db)
	ctx := context.Background()

	var successfulDatasetID uuid.UUID

	t.Run("commit dataset and all items", func(t *testing.T) {
		dataset := &entities.GoldenDataset{
			ID:        uuid.New(),
			Name:      "Integration Test Dataset",
			CreatedBy: uuid.NewString(),
		}

		items := goldenTestItems()

		err := repo.CreateWithItems(ctx, dataset, items)
		if err != nil {
			t.Fatalf("create failed: %v", err)
		}

		successfulDatasetID = dataset.ID

		loaded, err := repo.GetByID(ctx, dataset.ID)
		if err != nil {
			t.Fatal(err)
		}

		if loaded.ItemCount != 2 || len(loaded.Items) != 2 {
			t.Fatalf(
				"expected 2 items, count=%d loaded=%d",
				loaded.ItemCount,
				len(loaded.Items),
			)
		}

		for _, item := range loaded.Items {
			if item.GoldenDatasetID != dataset.ID {
				t.Fatal("incorrect foreign key")
			}
			if len(item.GUIElements) == 0 {
				t.Fatal("GUI elements were not persisted")
			}
		}
	})

	t.Run("rollback when item insert fails", func(t *testing.T) {
		dataset := &entities.GoldenDataset{
			ID:        uuid.New(),
			Name:      "Rollback Test Dataset",
			CreatedBy: uuid.NewString(),
		}

		items := goldenTestItems()

		// Exceeds the varchar(200) database constraint.
		items[1].ItemTitle = strings.Repeat("X", 201)

		err := repo.CreateWithItems(ctx, dataset, items)
		if err == nil {
			t.Fatal("expected database insert failure")
		}

		var parentCount int64
		if err := db.Model(&entities.GoldenDataset{}).
			Where("id = ?", dataset.ID).
			Count(&parentCount).Error; err != nil {
			t.Fatal(err)
		}

		var itemCount int64
		if err := db.Model(&entities.GoldenDatasetItem{}).
			Where("golden_dataset_id = ?", dataset.ID).
			Count(&itemCount).Error; err != nil {
			t.Fatal(err)
		}

		if parentCount != 0 || itemCount != 0 {
			t.Fatalf(
				"rollback failed: parents=%d items=%d",
				parentCount,
				itemCount,
			)
		}
	})

	t.Run("foreign key rejects orphan items", func(t *testing.T) {
		item := goldenTestItems()[0]
		item.ID = uuid.New()
		item.GoldenDatasetID = uuid.New()

		err := db.Create(&item).Error
		if err == nil {
			t.Fatal("orphan item was accepted")
		}
	})

	t.Run("dataset with items cannot be deleted", func(t *testing.T) {
		if successfulDatasetID == uuid.Nil {
			t.Skip("successful dataset not available")
		}

		err := db.Exec(
			"DELETE FROM golden_datasets WHERE id = ?",
			successfulDatasetID,
		).Error

		if err == nil {
			t.Fatal("dataset with items was deleted")
		}

		var count int64
		if err := db.Model(&entities.GoldenDataset{}).
			Where("id = ?", successfulDatasetID).
			Count(&count).Error; err != nil {
			t.Fatal(err)
		}

		if count != 1 {
			t.Fatal("original dataset no longer exists")
		}
	})
}
