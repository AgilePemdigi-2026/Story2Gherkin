package bdd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/story2gherkin/modules/auth/handler"
	"github.com/story2gherkin/modules/auth/service"
	userDto "github.com/story2gherkin/modules/user/dto"
	"github.com/story2gherkin/pkg/constants"
	"gorm.io/gorm"

	"github.com/cucumber/godog"
)

type fakeAuthService struct {
	service.AuthService
	emails map[string]bool
}

func (f *fakeAuthService) Register(_ context.Context, req userDto.UserCreateRequest) (userDto.UserResponse, error) {
	if f.emails[req.Email] {
		return userDto.UserResponse{}, userDto.ErrEmailAlreadyExists
	}
	f.emails[req.Email] = true
	return userDto.UserResponse{ID: "user-bdd", Name: req.Name, Email: req.Email, Role: req.Role}, nil
}

type authFeature struct {
	emails map[string]bool
	rec    *httptest.ResponseRecorder
}

func (f *authFeature) tidakAdaPengguna(email string) error {
	delete(f.emails, email)
	return nil
}

func (f *authFeature) sudahAdaPengguna(email string) error {
	f.emails[email] = true
	return nil
}

func (f *authFeature) sayaMendaftar(name, email, password, role string) error {
	gin.SetMode(gin.TestMode)
	injector := do.New()
	do.ProvideNamed(injector, constants.DB, func(*do.Injector) (*gorm.DB, error) {
		return &gorm.DB{}, nil
	})
	h := handler.NewAuthController(injector, &fakeAuthService{emails: f.emails})

	router := gin.New()
	router.POST("/register", h.Register)

	body := fmt.Sprintf(`{"name":%q,"email":%q,"password":%q,"role":%q}`, name, email, password, role)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	f.rec = httptest.NewRecorder()
	router.ServeHTTP(f.rec, req)
	return nil
}

func (f *authFeature) statusHarus(code int) error {
	if f.rec.Code != code {
		return fmt.Errorf("status = %d, ingin %d, body=%s", f.rec.Code, code, f.rec.Body.String())
	}
	return nil
}

func (f *authFeature) responsBerisiEmail(email string) error {
	var res struct {
		Data struct {
			Email string `json:"email"`
		} `json:"data"`
	}
	if err := json.Unmarshal(f.rec.Body.Bytes(), &res); err != nil {
		return fmt.Errorf("gagal decode respons: %w", err)
	}
	if res.Data.Email != email {
		return fmt.Errorf("email = %q, ingin %q", res.Data.Email, email)
	}
	return nil
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/auth_register.feature"},
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	f := &authFeature{}

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		f.emails = map[string]bool{}
		f.rec = nil
		return ctx, nil
	})

	ctx.Step(`^tidak ada pengguna beremail "([^"]*)"$`, f.tidakAdaPengguna)
	ctx.Step(`^sudah ada pengguna beremail "([^"]*)"$`, f.sudahAdaPengguna)
	ctx.Step(`^saya mendaftar dengan nama "([^"]*)", email "([^"]*)", password "([^"]*)", role "([^"]*)"$`, f.sayaMendaftar)
	ctx.Step(`^status respons harus (\d+)$`, f.statusHarus)
	ctx.Step(`^respons berisi email "([^"]*)"$`, f.responsBerisiEmail)
}
