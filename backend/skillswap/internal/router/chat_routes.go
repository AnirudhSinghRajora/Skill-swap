package router

import (
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/chat"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupChatRoutes configures all chat-related routes.
// All endpoints require JWT authentication.
func SetupChatRoutes(api *gin.RouterGroup, cfg *config.Config, chatHandler *chat.Handler) {
	// Public: serve chat images (browser <img> tags can't send JWT headers)
	api.GET("/chat/images/:id", chatHandler.ServeImage)

	chatGroup := api.Group("/chat")
	chatGroup.Use(middleware.JWTAuth(*cfg))
	{
		// ── Conversations ────────────────────────────────────────────
		chatGroup.GET("/conversations", chatHandler.GetConversations)
		chatGroup.GET("/conversations/by-swap/:swapId", chatHandler.GetConversationBySwap)
		chatGroup.GET("/conversations/:id", chatHandler.GetConversation)

		// ── Messages ─────────────────────────────────────────────────
		chatGroup.GET("/conversations/:id/messages", chatHandler.GetMessages)
		chatGroup.POST("/conversations/:id/messages",
			middleware.RateLimit(middleware.RateLimitConfig{
				Max:      30,
				Duration: time.Minute,
				Message:  "Message rate limit exceeded. Please slow down.",
				KeyFunc: func(c *gin.Context) string {
					uid, _ := c.Get("user_id")
					return "chat_msg:" + uid.(string)
				},
			}),
			chatHandler.SendMessage,
		)

		// ── Read status ──────────────────────────────────────────────
		chatGroup.PUT("/conversations/:id/read", chatHandler.MarkRead)
		chatGroup.GET("/unread-count", chatHandler.GetUnreadCount)

		// ── Message editing / deletion ───────────────────────────────
		chatGroup.PUT("/messages/:id", chatHandler.EditMessage)
		chatGroup.DELETE("/messages/:id", chatHandler.DeleteMessage)

		// ── Images ───────────────────────────────────────────────────
		chatGroup.POST("/images",
			middleware.RateLimit(middleware.RateLimitConfig{
				Max:      10,
				Duration: time.Minute,
				Message:  "Image upload rate limit exceeded.",
				KeyFunc: func(c *gin.Context) string {
					uid, _ := c.Get("user_id")
					return "chat_img:" + uid.(string)
				},
			}),
			chatHandler.UploadImage,
		)
		// ── Swap completion ──────────────────────────────────────────
		chatGroup.PUT("/swaps/:swapId/complete", chatHandler.MarkSwapComplete)
		chatGroup.DELETE("/swaps/:swapId/complete", chatHandler.UndoSwapComplete)
	}
}
