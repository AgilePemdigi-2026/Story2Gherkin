package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"github.com/story2gherkin/modules/auth/handler"
)

func RegisterRoutes(
	server *gin.RouterGroup,
	injector *do.Injector,
) {
	authHandler := do.MustInvoke[handler.AuthHandler](injector)

	authRoutes := server.Group("/auth")
	{
		authRoutes.POST("/register", authHandler.Register)

		// Forgot Password
		authRoutes.POST(
			"/forgot-password",
			authHandler.ForgotPassword,
		)

		authRoutes.POST("/reset-password", authHandler.ResetPassword)
	}
}
