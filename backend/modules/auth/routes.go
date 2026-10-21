package auth

import (
	"github.com/story2gherkin/modules/auth/handler"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.RouterGroup, injector *do.Injector) {
	authHandler := do.MustInvoke[handler.AuthHandler](injector)

	authRoutes := server.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)
		// Login, Logout, sama Forgot Password
	}
}