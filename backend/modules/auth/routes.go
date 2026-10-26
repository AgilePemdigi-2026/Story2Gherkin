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

		// Refresh Token
		authRoutes.POST("/refresh", authHandler.RefreshToken)

		// Send Password Reset
		authRoutes.POST("/password-reset/send", authHandler.SendPasswordReset)

		// Reset Password
		authRoutes.POST("/password-reset/reset", authHandler.ResetPassword)
	}
}
