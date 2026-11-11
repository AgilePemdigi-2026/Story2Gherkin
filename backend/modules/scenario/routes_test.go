package scenario_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/pkg/constants"
)

// Guards the production wiring: handlers that exist but are never
// registered return 404 in the running app, which unit tests miss.
func TestRegisterRoutesRegistersGoldenDataset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	injector := do.New()
	do.ProvideValue(injector, handler.NewScenarioHandler(nil))
	do.ProvideValue(injector, handler.NewReviewHandler(nil))
	do.ProvideValue(injector, handler.NewGoldenDatasetHandler(nil))
	do.ProvideNamedValue[authService.JWTService](injector, constants.JWTService, authService.NewJWTService())

	engine := gin.New()
	scenario.RegisterRoutes(engine.Group("/api/v1"), injector)

	registered := make(map[string]bool, len(engine.Routes()))
	for _, route := range engine.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, want := range []string{
		"POST /api/v1/golden-datasets",
		"GET /api/v1/golden-datasets",
		"GET /api/v1/golden-datasets/:id",
	} {
		if !registered[want] {
			t.Errorf("route %q is not registered", want)
		}
	}
}
