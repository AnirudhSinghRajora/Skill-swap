package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/rating"
	"github.com/gin-gonic/gin"
)

// SetupRatingRoutes sets up all rating-related routes
func SetupRatingRoutes(api *gin.RouterGroup, cfg *config.Config, ratingHandler *rating.Handler) {
	// Public rating routes (viewing ratings)
	ratingsGroup := api.Group("/ratings")
	{
		ratingsGroup.GET("/:id", ratingHandler.GetRating)
		ratingsGroup.GET("/swap/:swap_id", ratingHandler.GetSwapRatings)
	}

	// Protected rating routes (requires authentication)
	protectedRatings := api.Group("/ratings")
	protectedRatings.Use(middleware.JWTAuth(*cfg))
	{
		protectedRatings.POST("", ratingHandler.CreateRating)
		protectedRatings.PUT("/:id", ratingHandler.UpdateRating)
		protectedRatings.DELETE("/:id", ratingHandler.DeleteRating)
	}

	// User-specific rating routes
	usersGroup := api.Group("/users")
	{
		usersGroup.GET("/:id/ratings", ratingHandler.GetUserRatings)
		usersGroup.GET("/:id/ratings/stats", ratingHandler.GetUserRatingStats)
	}
}
