package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/qa"
	"github.com/gin-gonic/gin"
)

// SetupQARoutes wires the topic Q&A endpoints. Read endpoints are public
// (browsing without an account is fine); mutations and votes need auth.
func SetupQARoutes(api *gin.RouterGroup, cfg *config.Config, h *qa.Handler) {
	pub := api.Group("/questions")
	{
		pub.GET("", h.List)
		pub.GET("/:id", h.Get)
	}

	auth := api.Group("")
	auth.Use(middleware.JWTAuth(*cfg))
	{
		auth.POST("/questions", h.Ask)
		auth.PUT("/questions/:id", h.Update)
		auth.DELETE("/questions/:id", h.Delete)
		auth.POST("/questions/:id/answers", h.Answer)

		auth.PUT("/answers/:id", h.UpdateAnswer)
		auth.DELETE("/answers/:id", h.DeleteAnswer)
		auth.POST("/answers/:id/accept", h.Accept)

		auth.POST("/votes", h.Vote)
	}
}
