package scenario

import (
	"github.com/gin-gonic/gin"
	"github.com/story2gherkin/middlewares"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/handler"
)

// RegisterBehatRunRoutes registers end-to-end test endpoints for saved scenarios.
func RegisterBehatRunRoutes(
	server *gin.RouterGroup,
	jwt authService.JWTService,
	runs *handler.BehatRunHandler,
) {
	routes := server.Group("/scenarios", middlewares.Authenticate(jwt))
	routes.POST("/:id/behat-runs", runs.Run)
	routes.GET("/:id/behat-runs", runs.History)
}
