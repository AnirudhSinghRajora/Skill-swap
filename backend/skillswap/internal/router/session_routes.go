package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/session"
	"github.com/gin-gonic/gin"
)

// SetupSessionRoutes wires session-scheduling endpoints. Sessions are scoped
// to a parent swap for proposal/listing, and addressed by their own id for
// state transitions (accept / cancel / complete).
func SetupSessionRoutes(api *gin.RouterGroup, cfg *config.Config, h *session.Handler) {
	swaps := api.Group("/swaps")
	swaps.Use(middleware.JWTAuth(*cfg))
	{
		swaps.POST("/:id/sessions", h.Propose) // POST /api/v1/swaps/:id/sessions
		swaps.GET("/:id/sessions", h.List)     // GET  /api/v1/swaps/:id/sessions
	}

	sessions := api.Group("/sessions")
	sessions.Use(middleware.JWTAuth(*cfg))
	{
		sessions.PUT("/:id/accept", h.Accept)         // PUT /api/v1/sessions/:id/accept
		sessions.PUT("/:id/cancel", h.Cancel)         // PUT /api/v1/sessions/:id/cancel
		sessions.PUT("/:id/complete", h.Complete)     // PUT /api/v1/sessions/:id/complete
		sessions.GET("/:id/calendar.ics", h.ICS)      // GET /api/v1/sessions/:id/calendar.ics
		sessions.GET("/:id/notes", h.GetNotes)        // GET /api/v1/sessions/:id/notes
		sessions.PUT("/:id/notes", h.UpdateNotes)     // PUT /api/v1/sessions/:id/notes
	}
}
