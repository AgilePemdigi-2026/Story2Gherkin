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
	h := do.MustInvoke[handler.ScenarioHandler](injector)
	review := do.MustInvoke[*handler.ReviewHandler](injector)
	golden := do.MustInvoke[*handler.GoldenDatasetHandler](injector)
	jwt := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)
	routes := server.Group("/scenarios", middlewares.Authenticate(jwt))
	routes.POST("/generate", h.Generate)
	routes.GET("", review.List)
	routes.GET("/:id", review.Detail)
	routes.PUT("/:id", review.Save)
	routes.GET("/:id/download", review.Download)

	datasets := server.Group("/golden-datasets", middlewares.Authenticate(jwt))
	RegisterGoldenDatasetRoutes(datasets, golden)
}

func RegisterGoldenDatasetRoutes(
	group *gin.RouterGroup,
	golden *handler.GoldenDatasetHandler,
) {
	group.POST("", golden.Upload)
	group.GET("", golden.List)
	group.GET("/:id", golden.Detail)
	group.PUT("/:id", golden.Update)
	group.DELETE("/:id", golden.Delete)
}
