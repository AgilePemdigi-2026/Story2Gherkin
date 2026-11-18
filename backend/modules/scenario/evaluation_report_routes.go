package scenario

import (
	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/middlewares"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/handler"
)

// RegisterEvaluationReportRoutes registers authenticated report endpoints.
// The server group supplies the API prefix, such as /api/v1.
func RegisterEvaluationReportRoutes(
	server *gin.RouterGroup,
	jwtService authService.JWTService,
	reports *handler.EvaluationReportHandler,
) {
	routes := server.Group(
		"/evaluation-runs",
		middlewares.Authenticate(jwtService),
	)

	routes.GET("", reports.List)
	routes.GET("/:run_id", reports.Detail)
	routes.GET("/:run_id/download", reports.Download)
}
