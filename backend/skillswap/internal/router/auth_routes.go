package router

import (
	"time"

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

		// Email verification (consume token)
		authGroup.GET("/verify-email", authHandler.VerifyEmail)
		// Friendlier alias used by the frontend /auth/verify page
		authGroup.POST("/verify-email", authHandler.VerifyEmail)
	}

	// Protected auth routes (authentication required)
	authProtected := api.Group("/auth")
	authProtected.Use(middleware.JWTAuth(*cfg))
	{
		authProtected.POST("/logout", authHandler.Logout)
		authProtected.GET("/me", authHandler.GetMe)

		// Resend verification email — strict per-user rate limit so a
		// compromised account cannot mailbomb a victim.
		authProtected.POST("/resend-verification",
			middleware.RateLimit(middleware.RateLimitConfig{
				Max:      3,
				Duration: time.Hour,
				Message:  "Too many verification email requests. Please try again later.",
				KeyFunc: func(c *gin.Context) string {
					uid, _ := c.Get("user_id")
					return "resend_verify:" + uid.(string)
				},
			}),
			authHandler.ResendVerificationEmail,
		)
	}
}
