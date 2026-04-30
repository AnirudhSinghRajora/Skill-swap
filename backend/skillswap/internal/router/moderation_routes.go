package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/moderation"
	"github.com/gin-gonic/gin"
)

// SetupModerationRoutes wires the user-block / report endpoints (auth-required)
// and admin moderation queue (admin-only).
func SetupModerationRoutes(api *gin.RouterGroup, cfg *config.Config, h *moderation.Handler) {
	mod := api.Group("/moderation")
	mod.Use(middleware.JWTAuth(*cfg))
	{
		mod.POST("/blocks", h.Block)
		mod.GET("/blocks", h.ListBlocks)
		mod.DELETE("/blocks/:user_id", h.Unblock)
		mod.POST("/reports", h.CreateReport)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.JWTAuth(*cfg), middleware.AdminAuth())
	{
		admin.GET("/abuse-reports", h.AdminListReports)
		admin.POST("/abuse-reports/:id/resolve", h.AdminResolveReport)
	}
}
