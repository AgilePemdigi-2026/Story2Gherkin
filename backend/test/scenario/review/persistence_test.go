package review_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story2gherkin/database/entities"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/repository"
	"github.com/story2gherkin/modules/scenario/service"
	userRepo "github.com/story2gherkin/modules/user/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Uses a disposable schema, never modifies existing application tables.
// Set both TEST_REVIEW_DATABASE_URL and TEST_REVIEW_VALIDATOR_URL to run.
func TestReviewPersistenceWithPostgresAndPython(t *testing.T) {
	dsn, validatorURL := os.Getenv("TEST_REVIEW_DATABASE_URL"), os.Getenv("TEST_REVIEW_VALIDATOR_URL")
	if dsn == "" || validatorURL == "" {
		t.Skip("set review integration database and validator URLs")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	connection.SetMaxOpenConns(1)
	schema := "review_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	defer db.Exec("DROP SCHEMA " + schema + " CASCADE")
	if err := db.Exec("SET search_path TO " + schema + ", public").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entities.User{}, &entities.GenerationRequest{}, &entities.GherkinScenario{}); err != nil {
		t.Fatal(err)
	}
	owner := entities.User{ID: uuid.New(), Name: "Review analyst", Email: "analyst@example.test", Password: "not-a-login-password", Role: entities.RoleSystemAnalyst}
	other := entities.User{ID: uuid.New(), Name: "Other analyst", Email: "other@example.test", Password: "not-a-login-password", Role: entities.RoleSystemAnalyst}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	request := entities.GenerationRequest{ID: uuid.New(), UserID: owner.ID, UserStory: "As a user I want to login", GuiElements: []string{"Login button", "Missing field"}, FreeScenario: "Click Login button", Status: dto.RequestSuccess}
	if err := db.Create(&request).Error; err != nil {
		t.Fatal(err)
	}
	scenario := entities.GherkinScenario{ID: uuid.New(), GenerationRequestID: &request.ID, FeatureName: "Original", GherkinText: original, SourceType: dto.SourceLLM, Status: dto.ScenarioValid}
	if err := db.Create(&scenario).Error; err != nil {
		t.Fatal(err)
	}
	client, err := service.NewLLMClient(validatorURL, "review-verification-service-key-1234567890")
	if err != nil {
		t.Fatal(err)
	}
	review := service.NewReviewService(client, userRepo.NewUserRepository(db), repository.NewReviewRepository(db))
	ctx := context.Background()
	loaded, err := review.Detail(ctx, owner.ID.String(), scenario.ID.String())
	if err != nil || !loaded.Validation.Valid || loaded.Validation.CoveragePercent != 50 {
		t.Fatalf("load result=%+v err=%v", loaded, err)
	}
	_, err = review.Detail(ctx, other.ID.String(), scenario.ID.String())
	if !errors.Is(err, service.ErrScenarioNotFound) {
		t.Fatal("ownership check failed", err)
	}
	invalid := strings.Replace(original, "When I click", "Then I click", 1)
	_, err = review.Save(ctx, owner.ID.String(), scenario.ID.String(), dto.SaveCorrectionRequest{Gherkin: invalid, UpdatedAt: loaded.UpdatedAt})
	var invalidErr *service.InvalidCorrectionError
	if !errors.As(err, &invalidErr) {
		t.Fatal("invalid text accepted", err)
	}
	var persisted entities.GherkinScenario
	if err := db.First(&persisted, "id = ?", scenario.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.GherkinText != original || !persisted.UpdatedAt.Equal(loaded.UpdatedAt) {
		t.Fatal("rejected draft changed database")
	}
	corrected := strings.Replace(original, "Original", "Corrected", 1)
	saved, err := review.Save(ctx, owner.ID.String(), scenario.ID.String(), dto.SaveCorrectionRequest{Gherkin: corrected, UpdatedAt: loaded.UpdatedAt})
	if err != nil {
		t.Fatal(err)
	}
	if saved.SourceType != "CORRECTED" || saved.Validation.CoveragePercent != 50 || saved.UpdatedAt.Equal(loaded.UpdatedAt) {
		t.Fatalf("incorrect save %+v", saved)
	}
	_, err = review.Save(ctx, owner.ID.String(), scenario.ID.String(), dto.SaveCorrectionRequest{Gherkin: original, UpdatedAt: loaded.UpdatedAt})
	if !errors.Is(err, repository.ErrCorrectionConflict) {
		t.Fatal("stale update was not rejected", err)
	}
	download, err := review.Download(ctx, owner.ID.String(), scenario.ID.String())
	if err != nil || download.Gherkin != corrected {
		t.Fatalf("export changed bytes: %q err=%v", download.Gherkin, err)
	}
	// Zero-valued coverage and an empty parser message must be persisted too.
	zero := strings.Replace(corrected, "Login button", "an action", 1)
	zeroSaved, err := review.Save(ctx, owner.ID.String(), scenario.ID.String(), dto.SaveCorrectionRequest{Gherkin: zero, UpdatedAt: saved.UpdatedAt})
	if err != nil || zeroSaved.Validation.CoveragePercent != 0 {
		t.Fatalf("zero coverage save=%+v err=%v", zeroSaved, err)
	}
	if err := db.First(&persisted, "id = ?", scenario.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.CoveragePercent != 0 || persisted.CoveredGuiCount != 0 || persisted.ParserMessage != "" {
		t.Fatal("zero fields were not updated")
	}
}
