package bdd_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/modules/scenario/dto"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/modules/scenario/service"
)

type stubScenarioSvc struct {
	mode  string
	calls int
}

func (s *stubScenarioSvc) Generate(_ context.Context, _ string, _ dto.GenerateRequest) (dto.GenerateResponse, error) {
	s.calls++
	switch s.mode {
	case "sibuk":
		return dto.GenerateResponse{}, service.ErrUnavailable
	default:
		return dto.GenerateResponse{
			Feature: dto.Feature{FeatureName: "Hapus User"},
			Gherkin: "Feature: Hapus User\n",
		}, nil
	}
}

type generateFeature struct {
	svc  *stubScenarioSvc
	rec  *httptest.ResponseRecorder
	body string
}

func (f *generateFeature) generatorValid() error {
	f.svc.mode = "valid"
	return nil
}

func (f *generateFeature) generatorSibuk() error {
	f.svc.mode = "sibuk"
	return nil
}

func (f *generateFeature) generateLengkap() error {
	f.body = `{"user_story":"Sebagai admin saya ingin menghapus data user agar akun basi hilang.","gui_elements":["Tombol Hapus"],"free_scenario":"Admin login lalu klik hapus."}`
	return f.kirim()
}

func (f *generateFeature) generateTanpaStory() error {
	f.body = `{"gui_elements":["Tombol Hapus"],"free_scenario":"Admin klik hapus."}`
	return f.kirim()
}

func (f *generateFeature) kirim() error {
	gin.SetMode(gin.TestMode)
	h := handler.NewScenarioHandler(f.svc)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "00000000-0000-4000-8000-000000000001")
		c.Next()
	})
	router.POST("/generate", h.Generate)

	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader(f.body))
	req.Header.Set("Content-Type", "application/json")
	f.rec = httptest.NewRecorder()
	router.ServeHTTP(f.rec, req)
	return nil
}

func (f *generateFeature) statusHarus(code int) error {
	if f.rec.Code != code {
		return fmt.Errorf("status = %d, ingin %d, body=%s", f.rec.Code, code, f.rec.Body.String())
	}
	return nil
}

func (f *generateFeature) generatorDipanggil(n int) error {
	if f.svc.calls != n {
		return fmt.Errorf("generator dipanggil %d kali, ingin %d", f.svc.calls, n)
	}
	return nil
}

func InitializeGenerateScenario(ctx *godog.ScenarioContext) {
	f := &generateFeature{}

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		f.svc = &stubScenarioSvc{}
		f.rec = nil
		f.body = ""
		return ctx, nil
	})

	ctx.Step(`^generator mengembalikan skenario valid$`, f.generatorValid)
	ctx.Step(`^generator sedang sibuk$`, f.generatorSibuk)
	ctx.Step(`^saya generate skenario lengkap$`, f.generateLengkap)
	ctx.Step(`^saya generate tanpa user story$`, f.generateTanpaStory)
	ctx.Step(`^status respons harus (\d+)$`, f.statusHarus)
	ctx.Step(`^generator dipanggil (\d+) kali$`, f.generatorDipanggil)
	ctx.Step(`^generator tidak dipanggil$`, func() error { return f.generatorDipanggil(0) })
}

func TestGenerateFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeGenerateScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/scenario_generate.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
