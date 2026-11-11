package scenario

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/story2gherkin/middlewares"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/pkg/constants"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	// Existing Scenario Handlers
	h := do.MustInvoke[handler.ScenarioHandler](injector)
	review := do.MustInvoke[*handler.ReviewHandler](injector)

	// Golden Dataset Handler
	golden := do.MustInvoke[*handler.GoldenDatasetHandler](injector)

	// Generation History Handler
	historyHandler := do.MustInvoke[*handler.HistoryHandler](injector)

	// JWT Service
	jwt := do.MustInvokeNamed[authService.JWTService](
		injector,
		constants.JWTService,
	)

	// Generate and Review routes
	routes := server.Group(
		"/scenarios",
		middlewares.Authenticate(jwt),
	)

	routes.POST("/generate", h.Generate)
	routes.GET("", review.List)
	routes.GET("/:id", review.Detail)
	routes.PUT("/:id", review.Save)
	routes.GET("/:id/download", review.Download)

	// Golden Dataset routes
	goldenRoutes := server.Group(
		"/golden-datasets",
		middlewares.Authenticate(jwt),
	)

	goldenRoutes.POST("", golden.Upload)
	goldenRoutes.GET("", golden.List)
	goldenRoutes.GET("/:id", golden.Detail)

	// Generation History routes
	historyRoutes := server.Group(
		"/history",
		middlewares.Authenticate(jwt),
	)

	historyRoutes.GET("", historyHandler.List)
	historyRoutes.GET("/:id", historyHandler.Detail)
}
