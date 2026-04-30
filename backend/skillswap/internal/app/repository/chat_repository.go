package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatRepository defines data access operations for the chat feature.
type ChatRepository interface {
	// Conversations
	CreateConversation(conv *models.Conversation) error
	GetConversationByID(id uuid.UUID) (*models.Conversation, error)
	GetConversationBySwapID(swapID uuid.UUID) (*models.Conversation, error)
	GetUserConversations(userID uuid.UUID) ([]models.Conversation, error)

	// Messages
	CreateMessage(msg *models.Message) error
	GetMessageByID(id uuid.UUID) (*models.Message, error)
	GetMessages(conversationID uuid.UUID, before *time.Time, limit int) ([]models.Message, error)
	UpdateMessageContent(messageID uuid.UUID, content string) error
	DeleteMessage(messageID uuid.UUID) error

	// Read status
	UpsertReadStatus(userID, conversationID uuid.UUID) error
	GetUnreadCount(userID, conversationID uuid.UUID) (int64, error)
	GetTotalUnreadCount(userID uuid.UUID) (int64, error)

	// Chat images
	CreateChatImage(img *models.ChatImage) error
	GetChatImage(imageID uuid.UUID) (*models.ChatImage, error)
	LinkImagesToMessage(imageIDs []uuid.UUID, messageID uuid.UUID) error

	// Chat audio (voice notes)
	CreateChatAudio(a *models.ChatAudio) error
	GetChatAudio(audioID uuid.UUID) (*models.ChatAudio, error)

	// Authorization
	IsParticipant(userID, conversationID uuid.UUID) (bool, error)
	GetConversationParticipantIDs(conversationID uuid.UUID) (uuid.UUID, uuid.UUID, error)

	// Conversation metadata
	UpdateLastMessageAt(conversationID uuid.UUID, t time.Time) error
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

// --- Conversations ---

func (r *chatRepository) CreateConversation(conv *models.Conversation) error {
	return r.db.Create(conv).Error
}

func (r *chatRepository) GetConversationByID(id uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Preload("Swap.Requester").
		Preload("Swap.Responder").
		Preload("Swap.OfferedSkill").
		Preload("Swap.WantedSkill").
		Where("conversation_id = ?", id).
		First(&conv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("conversation %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &conv, nil
}

func (r *chatRepository) GetConversationBySwapID(swapID uuid.UUID) (*models.Conversation, error) {
	var conv models.Conversation
	err := r.db.Where("swap_id = ?", swapID).First(&conv).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("conversation %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &conv, nil
}

// GetUserConversations returns all conversations the user participates in,
// ordered by most recent activity. Each result includes the swap context
// for building the conversation list UI.
func (r *chatRepository) GetUserConversations(userID uuid.UUID) ([]models.Conversation, error) {
	var conversations []models.Conversation
	err := r.db.
		Preload("Swap.Requester").
		Preload("Swap.Responder").
		Preload("Swap.OfferedSkill").
		Preload("Swap.WantedSkill").
		Joins("JOIN swap_requests ON swap_requests.swap_id = conversations.swap_id").
		Where("swap_requests.requester_id = ? OR swap_requests.responder_id = ?", userID, userID).
		Where("conversations.deleted_at IS NULL").
		Order("COALESCE(conversations.last_message_at, conversations.created_at) DESC").
		Find(&conversations).Error
	if err != nil {
		return nil, err
	}
	return conversations, nil
}

// --- Messages ---

func (r *chatRepository) CreateMessage(msg *models.Message) error {
	return r.db.Create(msg).Error
}

func (r *chatRepository) GetMessageByID(id uuid.UUID) (*models.Message, error) {
	var msg models.Message
	err := r.db.Where("message_id = ?", id).First(&msg).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("message %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &msg, nil
}

// GetMessages returns paginated messages for a conversation using cursor-based pagination.
// Messages are returned in ascending order (oldest first) so the UI can append naturally.
// Pass `before` to load older messages (scroll-up pagination).
func (r *chatRepository) GetMessages(conversationID uuid.UUID, before *time.Time, limit int) ([]models.Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := r.db.Where("conversation_id = ?", conversationID)
	if before != nil {
		query = query.Where("created_at < ?", *before)
	}

	var messages []models.Message
	err := query.
		Preload("Sender").
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse to ascending order for the client
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *chatRepository) UpdateMessageContent(messageID uuid.UUID, content string) error {
	now := time.Now()
	return r.db.Model(&models.Message{}).
		Where("message_id = ?", messageID).
		Updates(map[string]interface{}{
			"content":    content,
			"is_edited":  true,
			"updated_at": now,
		}).Error
}

func (r *chatRepository) DeleteMessage(messageID uuid.UUID) error {
	return r.db.Where("message_id = ?", messageID).Delete(&models.Message{}).Error
}

// --- Read Status ---

// UpsertReadStatus marks a conversation as read by updating last_read_at to now.
// Uses PostgreSQL ON CONFLICT to insert or update in one statement.
func (r *chatRepository) UpsertReadStatus(userID, conversationID uuid.UUID) error {
	now := time.Now()
	sql := `
		INSERT INTO message_read_status (user_id, conversation_id, last_read_at)
		VALUES (?, ?, ?)
		ON CONFLICT (user_id, conversation_id)
		DO UPDATE SET last_read_at = EXCLUDED.last_read_at
	`
	return r.db.Exec(sql, userID, conversationID, now).Error
}

// GetUnreadCount returns the number of unread messages in a specific conversation for a user.
func (r *chatRepository) GetUnreadCount(userID, conversationID uuid.UUID) (int64, error) {
	var count int64
	// Messages are unread if they were created after the user's last_read_at,
	// or if the user has no read status entry (never opened the conversation).
	sql := `
		SELECT COUNT(*)
		FROM messages m
		WHERE m.conversation_id = ?
		  AND m.sender_id != ?
		  AND m.deleted_at IS NULL
		  AND m.created_at > COALESCE(
		      (SELECT last_read_at FROM message_read_status
		       WHERE user_id = ? AND conversation_id = ?),
		      '1970-01-01'::timestamptz
		  )
	`
	err := r.db.Raw(sql, conversationID, userID, userID, conversationID).Scan(&count).Error
	return count, err
}

// GetTotalUnreadCount returns the total unread message count across all conversations for a user.
func (r *chatRepository) GetTotalUnreadCount(userID uuid.UUID) (int64, error) {
	var count int64
	sql := `
		SELECT COUNT(*)
		FROM messages m
		JOIN conversations c ON c.conversation_id = m.conversation_id
		JOIN swap_requests s ON s.swap_id = c.swap_id
		WHERE (s.requester_id = ? OR s.responder_id = ?)
		  AND m.sender_id != ?
		  AND m.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		  AND m.created_at > COALESCE(
		      (SELECT last_read_at FROM message_read_status
		       WHERE user_id = ? AND conversation_id = m.conversation_id),
		      '1970-01-01'::timestamptz
		  )
	`
	err := r.db.Raw(sql, userID, userID, userID, userID).Scan(&count).Error
	return count, err
}

// --- Chat Images ---

func (r *chatRepository) CreateChatImage(img *models.ChatImage) error {
	return r.db.Create(img).Error
}

// LinkImagesToMessage sets the message_id on the given chat images.
func (r *chatRepository) LinkImagesToMessage(imageIDs []uuid.UUID, messageID uuid.UUID) error {
	if len(imageIDs) == 0 {
		return nil
	}
	return r.db.Model(&models.ChatImage{}).Where("image_id IN ?", imageIDs).Update("message_id", messageID).Error
}

// GetChatImage retrieves a chat image by ID. The image_data (BYTEA) is included
// because this is used by the image-serving endpoint.
func (r *chatRepository) GetChatImage(imageID uuid.UUID) (*models.ChatImage, error) {
	var img models.ChatImage
	err := r.db.Where("image_id = ?", imageID).First(&img).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("chat image %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &img, nil
}

// --- Authorization ---

// IsParticipant checks whether a user is a participant (requester or responder)
// of the swap associated with a conversation.
func (r *chatRepository) IsParticipant(userID, conversationID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Raw(`
		SELECT COUNT(*)
		FROM conversations c
		JOIN swap_requests s ON s.swap_id = c.swap_id
		WHERE c.conversation_id = ?
		  AND (s.requester_id = ? OR s.responder_id = ?)
		  AND c.deleted_at IS NULL
	`, conversationID, userID, userID).Scan(&count).Error
	return count > 0, err
}

// GetConversationParticipantIDs returns the two user IDs in a conversation.
func (r *chatRepository) GetConversationParticipantIDs(conversationID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	var result struct {
		RequesterID uuid.UUID
		ResponderID uuid.UUID
	}
	err := r.db.Raw(`
		SELECT s.requester_id, s.responder_id
		FROM conversations c
		JOIN swap_requests s ON s.swap_id = c.swap_id
		WHERE c.conversation_id = ?
	`, conversationID).Scan(&result).Error
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return result.RequesterID, result.ResponderID, nil
}

// --- Metadata ---

func (r *chatRepository) UpdateLastMessageAt(conversationID uuid.UUID, t time.Time) error {
	return r.db.Model(&models.Conversation{}).
		Where("conversation_id = ?", conversationID).
		Update("last_message_at", t).Error
}

// CreateChatAudio inserts a new chat_audio row. The blob lives in BYTEA.
func (r *chatRepository) CreateChatAudio(a *models.ChatAudio) error {
	if err := r.db.Create(a).Error; err != nil {
		return fmt.Errorf("create chat audio: %w", err)
	}
	return nil
}

// GetChatAudio retrieves an audio row by id, including the blob.
func (r *chatRepository) GetChatAudio(audioID uuid.UUID) (*models.ChatAudio, error) {
	var a models.ChatAudio
	if err := r.db.Where("audio_id = ?", audioID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get chat audio: %w", err)
	}
	return &a, nil
}
