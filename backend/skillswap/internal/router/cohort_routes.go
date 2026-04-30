package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/cohort"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupCohortRoutes wires cohort + room endpoints. Listing and detail are
// public (anonymous browsing of public cohorts is fine), all mutations
// require auth.
func SetupCohortRoutes(api *gin.RouterGroup, cfg *config.Config, h *cohort.Handler) {
	pub := api.Group("/cohorts")
	{
		pub.GET("", h.List)
		pub.GET("/:id", h.Get)
		pub.GET("/:id/sessions", h.ListSessions)
	}

	auth := api.Group("/cohorts")
	auth.Use(middleware.JWTAuth(*cfg))
	{
		auth.POST("", h.Create)
		auth.POST("/:id/join", h.Join)
		auth.POST("/:id/leave", h.Leave)
		auth.POST("/:id/sessions", h.AddSession)
		auth.DELETE("/:id/members/:user_id", h.RemoveMember)
	}

	csAuth := api.Group("/cohort-sessions")
	csAuth.Use(middleware.JWTAuth(*cfg))
	{
		csAuth.POST("/:id/token", h.JoinToken) // POST /api/v1/cohort-sessions/:id/token
	}
}
