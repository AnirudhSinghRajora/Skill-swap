package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/repository"
	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// upgrader negotiates the WebSocket handshake.
// CheckOrigin allows all origins because authentication is token-based
// (passed via query parameter), not cookie-based, so CSRF is not a risk.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// WSHandler manages WebSocket upgrade, authentication, and message routing.
type WSHandler struct {
	hub         *Hub
	chatService appservice.ChatService
	chatRepo    repository.ChatRepository
	cfg         *config.Config
}

// NewWSHandler creates a new WebSocket handler.
func NewWSHandler(hub *Hub, chatService appservice.ChatService, chatRepo repository.ChatRepository, cfg *config.Config) *WSHandler {
	return &WSHandler{
		hub:         hub,
		chatService: chatService,
		chatRepo:    chatRepo,
		cfg:         cfg,
	}
}

// ── WebSocket protocol types ─────────────────────────────────────────────────

// wsIncoming is the envelope for every client → server message.
type wsIncoming struct {
	Type           string   `json:"type"`
	ConversationID string   `json:"conversation_id"`
	Content        string   `json:"content"`
	TempID         string   `json:"temp_id"`
	Encrypted      bool     `json:"encrypted"`
	ImageIDs       []string `json:"image_ids,omitempty"`
}

// wsOutgoing is the envelope for every server → client message.
type wsOutgoing struct {
	Type           string      `json:"type"`
	ConversationID string      `json:"conversation_id,omitempty"`
	UserID         string      `json:"user_id,omitempty"`
	Message        interface{} `json:"message,omitempty"`
	TempID         string      `json:"temp_id,omitempty"`
	Error          string      `json:"error,omitempty"`
	RefType        string      `json:"ref_type,omitempty"`
}

// ── HTTP → WebSocket upgrade ─────────────────────────────────────────────────

// HandleWebSocket upgrades an HTTP request to a WebSocket connection.
// Authentication is performed via the "token" query parameter because the
// browser WebSocket API does not support custom request headers.
//
// Route: GET /api/v1/ws?token=<jwt>
func (wh *WSHandler) HandleWebSocket(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token query parameter"})
		return
	}

	userID, err := wh.validateToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade writes its own HTTP error; just log.
		log.Printf("[WS] Upgrade failed for user %s: %v", userID, err)
		return
	}

	client := newClient(wh.hub, conn, userID, wh.processMessage)
	wh.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// validateToken parses and validates a JWT, returning the user_id claim.
// It mirrors the same validation logic used by the JWTAuth middleware
// (HMAC signing method, same secret).
func (wh *WSHandler) validateToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(wh.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("missing user_id claim")
	}

	return userID, nil
}

// ── Message dispatch ─────────────────────────────────────────────────────────

// processMessage is the callback invoked by a Client's readPump for every
// incoming WebSocket frame.
func (wh *WSHandler) processMessage(client *Client, raw []byte) {
	var msg wsIncoming
	if err := json.Unmarshal(raw, &msg); err != nil {
		wh.sendError(client, "Invalid message format", "", "")
		return
	}

	switch msg.Type {
	case "send_message":
		wh.handleSendMessage(client, msg)
	case "typing":
		wh.handleTyping(client, msg)
	case "stop_typing":
		wh.handleStopTyping(client, msg)
	case "mark_read":
		wh.handleMarkRead(client, msg)
	case "call_invite", "call_accept", "call_reject", "call_end":
		wh.handleCallSignal(client, msg)
	default:
		wh.sendError(client, "Unknown message type: "+msg.Type, msg.Type, msg.TempID)
	}
}

// ── send_message ─────────────────────────────────────────────────────────────

func (wh *WSHandler) handleSendMessage(client *Client, msg wsIncoming) {
	if !client.checkRateLimit() {
		wh.sendError(client, "Rate limit exceeded. Maximum 30 messages per minute.", "send_message", msg.TempID)
		return
	}

	convID, err := uuid.Parse(msg.ConversationID)
	if err != nil {
		wh.sendError(client, "Invalid conversation_id", "send_message", msg.TempID)
		return
	}

	content := strings.TrimSpace(msg.Content)
	// Encrypted messages are base64-encoded ciphertext which is ~33% larger than plaintext.
	// Allow a higher limit for encrypted content to accommodate the encoding overhead.
	maxLen := 10000
	if msg.Encrypted {
		maxLen = 20000
	}
	if len(content) == 0 || len(content) > maxLen {
		wh.sendError(client, fmt.Sprintf("Message content must be 1\u2013%d characters", maxLen), "send_message", msg.TempID)
		return
	}

	senderID, err := uuid.Parse(client.UserID)
	if err != nil {
		wh.sendError(client, "Invalid session", "send_message", msg.TempID)
		return
	}

	// Parse explicit image IDs (for encrypted messages with embedded images)
	var imageIDs []uuid.UUID
	for _, idStr := range msg.ImageIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			imageIDs = append(imageIDs, id)
		}
	}

	// Persist via the shared ChatService (includes participant validation).
	savedMsg, err := wh.chatService.SendMessage(convID, senderID, content, msg.Encrypted, imageIDs)
	if err != nil {
		log.Printf("[WS] SendMessage error user=%s conv=%s: %v", client.UserID, msg.ConversationID, err)
		wh.sendError(client, "Failed to send message", "send_message", msg.TempID)
		return
	}

	out := wsOutgoing{
		Type:    "new_message",
		TempID:  msg.TempID,
		Message: toMessageResponse(savedMsg),
	}
	data, err := json.Marshal(out)
	if err != nil {
		log.Printf("[WS] Marshal error: %v", err)
		return
	}

	// Deliver to both participants (sender gets confirmation, recipient gets message).
	wh.sendToConversationParticipants(convID, data)
}

// ── typing / stop_typing ─────────────────────────────────────────────────────

func (wh *WSHandler) handleTyping(client *Client, msg wsIncoming) {
	wh.forwardPresenceEvent(client, msg, "user_typing")
}

func (wh *WSHandler) handleStopTyping(client *Client, msg wsIncoming) {
	wh.forwardPresenceEvent(client, msg, "user_stop_typing")
}

// forwardPresenceEvent validates participation and sends a presence
// indicator (typing / stop_typing) to the other participant.
func (wh *WSHandler) forwardPresenceEvent(client *Client, msg wsIncoming, eventType string) {
	convID, err := uuid.Parse(msg.ConversationID)
	if err != nil {
		return
	}

	userID, err := uuid.Parse(client.UserID)
	if err != nil {
		return
	}

	// Lightweight authorization check — only participants may send presence.
	ok, err := wh.chatRepo.IsParticipant(userID, convID)
	if err != nil || !ok {
		return
	}

	out := wsOutgoing{
		Type:           eventType,
		ConversationID: msg.ConversationID,
		UserID:         client.UserID,
	}
	data, _ := json.Marshal(out)
	wh.sendToOtherParticipant(convID, client.UserID, data)
}

// ── mark_read ────────────────────────────────────────────────────────────────

func (wh *WSHandler) handleMarkRead(client *Client, msg wsIncoming) {
	convID, err := uuid.Parse(msg.ConversationID)
	if err != nil {
		return
	}

	userID, err := uuid.Parse(client.UserID)
	if err != nil {
		return
	}

	// ChatService.MarkConversationRead validates participation internally.
	if err := wh.chatService.MarkConversationRead(convID, userID); err != nil {
		log.Printf("[WS] MarkRead error user=%s conv=%s: %v", client.UserID, msg.ConversationID, err)
		return
	}

	out := wsOutgoing{
		Type:           "message_read",
		ConversationID: msg.ConversationID,
		UserID:         client.UserID,
	}
	data, _ := json.Marshal(out)
	wh.sendToOtherParticipant(convID, client.UserID, data)
}

// ── Delivery helpers ─────────────────────────────────────────────────────────

// sendToConversationParticipants delivers data to both participants.
func (wh *WSHandler) sendToConversationParticipants(conversationID uuid.UUID, data []byte) {
	requesterID, responderID, err := wh.chatRepo.GetConversationParticipantIDs(conversationID)
	if err != nil {
		log.Printf("[WS] GetConversationParticipantIDs error conv=%s: %v", conversationID, err)
		return
	}
	wh.hub.SendToUser(requesterID.String(), data)
	wh.hub.SendToUser(responderID.String(), data)
}

// sendToOtherParticipant delivers data to the participant who is NOT senderUserID.
func (wh *WSHandler) sendToOtherParticipant(conversationID uuid.UUID, senderUserID string, data []byte) {
	requesterID, responderID, err := wh.chatRepo.GetConversationParticipantIDs(conversationID)
	if err != nil {
		return
	}
	if requesterID.String() == senderUserID {
		wh.hub.SendToUser(responderID.String(), data)
	} else {
		wh.hub.SendToUser(requesterID.String(), data)
	}
}

// sendError delivers a typed error message to a single client.
func (wh *WSHandler) sendError(client *Client, message, refType, tempID string) {
	out := wsOutgoing{
		Type:    "error",
		Error:   message,
		RefType: refType,
		TempID:  tempID,
	}
	data, _ := json.Marshal(out)
	select {
	case client.send <- data:
	default:
		// Buffer full — client will be evicted by the hub.
	}
}

// ── Call signaling ───────────────────────────────────────────────────────────

// handleCallSignal forwards call_invite / call_accept / call_reject / call_end
// to the other participant in the conversation via WebSocket.
func (wh *WSHandler) handleCallSignal(client *Client, msg wsIncoming) {
	convID, err := uuid.Parse(msg.ConversationID)
	if err != nil {
		wh.sendError(client, "Invalid conversation_id", msg.Type, msg.TempID)
		return
	}

	userID, err := uuid.Parse(client.UserID)
	if err != nil {
		return
	}

	ok, err := wh.chatRepo.IsParticipant(userID, convID)
	if err != nil || !ok {
		wh.sendError(client, "Not a participant", msg.Type, msg.TempID)
		return
	}

	out := wsOutgoing{
		Type:           msg.Type,
		ConversationID: msg.ConversationID,
		UserID:         client.UserID,
	}
	data, _ := json.Marshal(out)
	wh.sendToOtherParticipant(convID, client.UserID, data)
}
