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
		// Register
		authRoutes.POST("/register", authHandler.Register)

		// Login
		authRoutes.POST("/login", authHandler.Login)

		// Logout
		authRoutes.POST("/logout", authHandler.Logout)

		// Forgot Password
		authRoutes.POST("/forgot-password", authHandler.ForgotPassword)

		// Reset Password
		authRoutes.POST("/reset-password", authHandler.ResetPassword)
	}
}
