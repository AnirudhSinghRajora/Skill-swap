package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/user"
	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures all user-related routes
func SetupUserRoutes(api *gin.RouterGroup, userService service.UserService, cfg *config.Config) {
	userHandler := user.NewHandler(userService)

	// Public user routes (no authentication required)
	public := api.Group("/public")
	{
		public.GET("/users/search", userHandler.SearchUsers)
		public.GET("/users/:id", userHandler.GetPublicProfile)
	}

	// Public slug-based profile: /api/v1/u/:slug
	api.GET("/u/:slug", userHandler.GetPublicProfileBySlug)

	// Protected user routes (authentication required)
	protected := api.Group("/users")
	protected.Use(middleware.JWTAuth(*cfg))
	{
		protected.GET("/profile", userHandler.GetProfile)
		protected.PUT("/profile", userHandler.UpdateProfile)
		protected.PUT("/profile/slug", userHandler.UpdateMySlug)

		// E2EE key management
		protected.PUT("/me/e2ee-keys", userHandler.SetE2EEKeys)
		protected.GET("/me/key-backup", userHandler.GetKeyBackup)
		protected.GET("/:id/public-key", userHandler.GetUserPublicKey)
	}
}
