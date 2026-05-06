package chat

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler handles HTTP requests for the chat feature.
type Handler struct {
	chatService appservice.ChatService
}

// NewHandler creates a new chat handler.
func NewHandler(chatService appservice.ChatService) *Handler {
	return &Handler{chatService: chatService}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func (h *Handler) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return uuid.Nil, false
	}
	return userID, true
}

func (h *Handler) parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid " + param + " format"})
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrNotParticipant):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrWrongStatus):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrInvalidFileType):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		resp.InternalError(c, err)
	}
}

func toMessageResponse(msg *models.Message) models.MessageResponse {
	return models.MessageResponse{
		MessageID:      msg.MessageID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		Content:        msg.Content,
		Encrypted:      msg.Encrypted,
		HasImages:      msg.HasImages,
		IsEdited:       msg.IsEdited,
		CreatedAt:      msg.CreatedAt,
		UpdatedAt:      msg.UpdatedAt,
	}
}

// ── Conversations ────────────────────────────────────────────────────────────

// GetConversations returns all conversations for the authenticated user.
// GET /chat/conversations
func (h *Handler) GetConversations(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	conversations, err := h.chatService.GetUserConversations(userID)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

// GetConversationBySwap looks up (or creates) the conversation for a swap,
// then returns its full detail.
// GET /chat/conversations/by-swap/:swapId
func (h *Handler) GetConversationBySwap(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	swapID, ok := h.parseUUIDParam(c, "swapId")
	if !ok {
		return
	}

	conv, err := h.chatService.GetOrCreateConversation(swapID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	fullConv, err := h.chatService.GetConversation(conv.ConversationID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.buildConversationDetail(fullConv, userID))
}

// GetConversation returns full details of a single conversation.
// GET /chat/conversations/:id
func (h *Handler) GetConversation(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	convID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	conv, err := h.chatService.GetConversation(convID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.buildConversationDetail(conv, userID))
}

func (h *Handler) buildConversationDetail(conv *models.Conversation, currentUserID uuid.UUID) gin.H {
	requester := conv.Swap.Requester
	responder := conv.Swap.Responder

	requesterJSON := gin.H{
		"user_id":   requester.UserID.String(),
		"name":      requester.Name,
		"has_photo": len(requester.PhotoData) > 0,
	}
	responderJSON := gin.H{
		"user_id":   responder.UserID.String(),
		"name":      responder.Name,
		"has_photo": len(responder.PhotoData) > 0,
	}

	var currentUser, otherUser gin.H
	if conv.Swap.RequesterID == currentUserID {
		currentUser = requesterJSON
		otherUser = responderJSON
	} else {
		currentUser = responderJSON
		otherUser = requesterJSON
	}

	return gin.H{
		"conversation_id":     conv.ConversationID.String(),
		"swap_id":             conv.SwapID.String(),
		"current_user":        currentUser,
		"other_user":          otherUser,
		"offered_skill":       gin.H{"skill_id": conv.Swap.OfferedSkillID.String(), "name": conv.Swap.OfferedSkill.Name},
		"wanted_skill":        gin.H{"skill_id": conv.Swap.WantedSkillID.String(), "name": conv.Swap.WantedSkill.Name},
		"swap_status":         string(conv.Swap.Status),
		"requester_completed": conv.Swap.RequesterCompleted,
		"responder_completed": conv.Swap.ResponderCompleted,
		"created_at":          conv.CreatedAt,
	}
}

// ── Messages ─────────────────────────────────────────────────────────────────

// GetMessages returns paginated messages for a conversation.
// GET /chat/conversations/:id/messages?before=<RFC3339>&limit=<int>
func (h *Handler) GetMessages(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	convID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	var before *time.Time
	if b := c.Query("before"); b != "" {
		if t, err := time.Parse(time.RFC3339Nano, b); err == nil {
			before = &t
		} else if t, err := time.Parse(time.RFC3339, b); err == nil {
			before = &t
		}
	}

	messages, err := h.chatService.GetMessages(convID, userID, before, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	msgResponses := make([]models.MessageResponse, len(messages))
	for i := range messages {
		msgResponses[i] = toMessageResponse(&messages[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": msgResponses,
		"has_more": len(messages) == limit,
	})
}

// SendMessage creates a new message in a conversation.
// POST /chat/conversations/:id/messages
func (h *Handler) SendMessage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	convID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse explicit image IDs for encrypted messages
	var imageIDs []uuid.UUID
	for _, idStr := range req.ImageIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			imageIDs = append(imageIDs, id)
		}
	}

	msg, err := h.chatService.SendMessage(convID, userID, req.Content, req.Encrypted, imageIDs)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toMessageResponse(msg))
}

// EditMessage updates a message's content (sender only, within edit window).
// PUT /chat/messages/:id
func (h *Handler) EditMessage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	msgID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req models.EditMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatService.EditMessage(msgID, userID, req.Content); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message updated"})
}

// DeleteMessage soft-deletes a message (sender only).
// DELETE /chat/messages/:id
func (h *Handler) DeleteMessage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	msgID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.chatService.DeleteMessage(msgID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

// ── Read Status ──────────────────────────────────────────────────────────────

// MarkRead marks all messages in a conversation as read.
// PUT /chat/conversations/:id/read
func (h *Handler) MarkRead(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	convID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.chatService.MarkConversationRead(convID, userID); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation marked as read"})
}

// GetUnreadCount returns the total unread message count across all conversations.
// GET /chat/unread-count
func (h *Handler) GetUnreadCount(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	count, err := h.chatService.GetTotalUnreadCount(userID)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// ── Images ───────────────────────────────────────────────────────────────────

// UploadImage handles multipart image upload for chat messages.
// POST /chat/images
func (h *Handler) UploadImage(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, int64(appservice.MaxImageSize)+1))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read image data"})
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}

	// Check if this is an encrypted image upload (E2EE — opaque ciphertext)
	encrypted := c.PostForm("encrypted") == "true"

	img, err := h.chatService.UploadChatImage(userID, data, mimeType, encrypted)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"image_id":  img.ImageID.String(),
		"mime_type": img.MimeType,
		"file_size": img.FileSize,
	})
}

// ServeImage serves a chat image's binary data.
// GET /chat/images/:id — public (no auth) since <img> tags can't send JWT headers.
// Image IDs are unguessable UUIDs.
func (h *Handler) ServeImage(c *gin.Context) {
	imageID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	img, err := h.chatService.GetChatImagePublic(imageID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Header("Content-Type", img.MimeType)
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, img.MimeType, img.ImageData)
}

// ── Swap Completion ──────────────────────────────────────────────────────────

// MarkSwapComplete marks the current user's side of a swap as complete.
// PUT /chat/swaps/:swapId/complete
func (h *Handler) MarkSwapComplete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	swapID, ok := h.parseUUIDParam(c, "swapId")
	if !ok {
		return
	}

	swap, err := h.chatService.MarkSwapComplete(swapID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.buildSwapCompletionResponse(swap))
}

// UndoSwapComplete retracts the current user's completion mark.
// DELETE /chat/swaps/:swapId/complete
func (h *Handler) UndoSwapComplete(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}
	swapID, ok := h.parseUUIDParam(c, "swapId")
	if !ok {
		return
	}

	swap, err := h.chatService.UndoSwapComplete(swapID, userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.buildSwapCompletionResponse(swap))
}

func (h *Handler) buildSwapCompletionResponse(swap *models.SwapRequest) gin.H {
	return gin.H{
		"swap_id":             swap.SwapID.String(),
		"status":              string(swap.Status),
		"requester_completed": swap.RequesterCompleted,
		"responder_completed": swap.ResponderCompleted,
		"requester":           gin.H{"user_id": swap.Requester.UserID.String(), "name": swap.Requester.Name},
		"responder":           gin.H{"user_id": swap.Responder.UserID.String(), "name": swap.Responder.Name},
		"offered_skill":       gin.H{"skill_id": swap.OfferedSkill.SkillID.String(), "name": swap.OfferedSkill.Name},
		"wanted_skill":        gin.H{"skill_id": swap.WantedSkill.SkillID.String(), "name": swap.WantedSkill.Name},
	}
}
