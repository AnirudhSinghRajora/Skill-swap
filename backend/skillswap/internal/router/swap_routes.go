package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/swap"
	"github.com/gin-gonic/gin"
)

// SetupSwapRoutes configures all swap request related routes
func SetupSwapRoutes(api *gin.RouterGroup, cfg *config.Config, swapHandler *swap.Handler) {
	// Protected swap routes (authentication required)
	swaps := api.Group("/swaps")
	swaps.Use(middleware.JWTAuth(*cfg))
	{
		swaps.POST("", swapHandler.CreateSwapRequest)          // POST /api/v1/swaps
		swaps.GET("", swapHandler.GetUserSwapRequests)         // GET /api/v1/swaps
		swaps.GET("/matches", swapHandler.GetPotentialMatches) // GET /api/v1/swaps/matches
		swaps.GET("/:id", swapHandler.GetSwapRequest)          // GET /api/v1/swaps/:id
		swaps.PUT("/:id/status", swapHandler.UpdateSwapStatus) // PUT /api/v1/swaps/:id/status
		swaps.DELETE("/:id", swapHandler.DeleteSwapRequest)    // DELETE /api/v1/swaps/:id
		swaps.POST("/:id/no-show", swapHandler.ReportNoShow)   // POST /api/v1/swaps/:id/no-show
		swaps.POST("/:id/dispute", swapHandler.RaiseDispute)   // POST /api/v1/swaps/:id/dispute
	}

	// Public-ish reliability stats (auth required for consistency with other user endpoints)
	users := api.Group("/users")
	users.Use(middleware.JWTAuth(*cfg))
	{
		users.GET("/:id/reliability", swapHandler.GetReliability) // GET /api/v1/users/:id/reliability
	}
}
