package scenario

import (
	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/middlewares"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/handler"
)

// RegisterEvaluationRunRoutes registers QA endpoints that create and monitor runs.
func RegisterEvaluationRunRoutes(
	server *gin.RouterGroup,
	jwtService authService.JWTService,
	runs *handler.EvaluationRunHandler,
) {
	routes := server.Group(
		"/evaluation-runs",
		middlewares.Authenticate(jwtService),
	)

	routes.POST("", runs.Create)
	routes.GET("/:run_id/progress", runs.Progress)
}
