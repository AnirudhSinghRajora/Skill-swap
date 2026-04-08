package router

import (
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/video"
	"github.com/gin-gonic/gin"
)

// SetupVideoRoutes configures video call routes
func SetupVideoRoutes(api *gin.RouterGroup, cfg *config.Config, videoHandler *video.Handler) {
	videoGroup := api.Group("/video")
	videoGroup.Use(middleware.JWTAuth(*cfg))
	{
		videoGroup.POST("/token", videoHandler.GetToken)
	}
}
