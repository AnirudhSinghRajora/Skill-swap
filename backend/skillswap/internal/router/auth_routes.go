package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/auth"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures all authentication-related routes
func SetupAuthRoutes(api *gin.RouterGroup, authService service.AuthService, cfg *config.Config) {
	authHandler := auth.NewHandler(authService, cfg)

	// Public auth routes (no authentication required)
	authGroup := api.Group("/auth")
	authGroup.Use(middleware.AuthRateLimit())
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)

		// Google OAuth
		authGroup.GET("/google", authHandler.GoogleRedirect)
		authGroup.GET("/google/callback", authHandler.GoogleCallback)

		// Password reset
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)

		// Email verification
		authGroup.GET("/verify-email", authHandler.VerifyEmail)
	}

	// Protected auth routes (authentication required)
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.JWTAuth(*cfg))
	{
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.GET("/me", authHandler.GetMe)
	}
}
