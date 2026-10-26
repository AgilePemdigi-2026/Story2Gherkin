package scenario

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	authService "github.com/story2gherkin/modules/auth/service"
	"github.com/story2gherkin/modules/scenario/handler"
	"github.com/story2gherkin/pkg/constants"
	"github.com/story2gherkin/pkg/utils"
)

func RequireJWT(jwt authService.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		fields := strings.Fields(c.GetHeader("Authorization"))
		if len(fields) != 2 || !strings.EqualFold(fields[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.BuildResponseFailed("unauthorized", "A valid Bearer token is required.", nil))
			return
		}
		userID, err := jwt.GetUserIDByToken(fields[1])
		if err != nil || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.BuildResponseFailed("unauthorized", "A valid Bearer token is required.", nil))
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	h := do.MustInvoke[handler.ScenarioHandler](injector)
	jwt := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)
	routes := server.Group("/scenarios", RequireJWT(jwt))
	routes.POST("/generate", h.Generate)
}
