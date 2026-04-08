package video

import (
	"net/http"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles video call API endpoints.
type Handler struct {
	videoService *service.VideoService
}

// NewHandler creates a new video handler.
func NewHandler(videoService *service.VideoService) *Handler {
	return &Handler{videoService: videoService}
}

// TokenRequest is the request body for generating a video call token.
type TokenRequest struct {
	ConversationID string `json:"conversation_id" binding:"required"`
}

// TokenResponse is the response for a video call token request.
type TokenResponse struct {
	Token string `json:"token"`
	URL   string `json:"url"`
	Room  string `json:"room"`
}

// GetToken generates a LiveKit token for the authenticated user to join a call.
func (h *Handler) GetToken(c *gin.Context) {
	if !h.videoService.IsConfigured() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Video calls are not configured"})
		return
	}

	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	userName := "User"
	if name, ok := c.Get("user_name"); ok {
		userName = name.(string)
	}

	roomName := service.GetRoomName(conversationID)
	token, err := h.videoService.GenerateToken(userID, userName, roomName)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Token: token,
		URL:   h.videoService.GetConnectionURL(),
		Room:  roomName,
	})
}
