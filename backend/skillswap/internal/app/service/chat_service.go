package service

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/repository"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// MaxEditWindow is how long after sending a user can edit their message.
	MaxEditWindow = 15 * time.Minute
	// MaxImageSize is the maximum allowed chat image size in bytes (5MB).
	MaxImageSize = 5 * 1024 * 1024
)

// ChatService defines the business logic for the chat feature.
type ChatService interface {
	// Conversations
	GetOrCreateConversation(swapID uuid.UUID) (*models.Conversation, error)
	GetConversation(conversationID, userID uuid.UUID) (*models.Conversation, error)
	GetUserConversations(userID uuid.UUID) ([]models.ConversationResponse, error)

	// Messages
	SendMessage(conversationID, senderID uuid.UUID, content string, encrypted bool, imageIDs []uuid.UUID) (*models.Message, error)
	GetMessages(conversationID, userID uuid.UUID, before *time.Time, limit int) ([]models.Message, error)
	EditMessage(messageID, userID uuid.UUID, content string) error
	DeleteMessage(messageID, userID uuid.UUID) error

	// Read status
	MarkConversationRead(conversationID, userID uuid.UUID) error
	GetTotalUnreadCount(userID uuid.UUID) (int64, error)

	// Images
	UploadChatImage(uploaderID uuid.UUID, data []byte, mimeType string, encrypted bool) (*models.ChatImage, error)
	GetChatImage(imageID, userID uuid.UUID) (*models.ChatImage, error)
	GetChatImagePublic(imageID uuid.UUID) (*models.ChatImage, error)

	// Swap completion
	MarkSwapComplete(swapID, userID uuid.UUID) (*models.SwapRequest, error)
	UndoSwapComplete(swapID, userID uuid.UUID) (*models.SwapRequest, error)
}

type chatService struct {
	repo                repository.ChatRepository
	db                  *gorm.DB
	notificationService *NotificationService
}

func NewChatService(repo repository.ChatRepository, db *gorm.DB, notificationService *NotificationService) ChatService {
	return &chatService{
		repo:                repo,
		db:                  db,
		notificationService: notificationService,
	}
}

// --- Conversations ---

// GetOrCreateConversation returns the conversation for a swap, creating it if it doesn't exist.
// The swap must be in accepted or completed status.
func (s *chatService) GetOrCreateConversation(swapID uuid.UUID) (*models.Conversation, error) {
	// Check if conversation already exists
	conv, err := s.repo.GetConversationBySwapID(swapID)
	if err == nil {
		return conv, nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	// Verify swap exists and is in an eligible status
	var swap models.SwapRequest
	if err := s.db.Where("swap_id = ?", swapID).First(&swap).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("swap request %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	if swap.Status != models.StatusAccepted && swap.Status != models.StatusCompleted {
		return nil, fmt.Errorf("swap must be accepted to start a conversation: %w", apperrors.ErrWrongStatus)
	}

	conv = &models.Conversation{SwapID: swapID}
	if err := s.repo.CreateConversation(conv); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}
	return conv, nil
}

// GetConversation retrieves a conversation after verifying the user is a participant.
func (s *chatService) GetConversation(conversationID, userID uuid.UUID) (*models.Conversation, error) {
	ok, err := s.repo.IsParticipant(userID, conversationID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("user is not a participant of this conversation: %w", apperrors.ErrNotParticipant)
	}
	return s.repo.GetConversationByID(conversationID)
}

// GetUserConversations returns all conversations for a user, enriched with
// unread counts, last message preview, and the other participant's info.
func (s *chatService) GetUserConversations(userID uuid.UUID) ([]models.ConversationResponse, error) {
	conversations, err := s.repo.GetUserConversations(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.ConversationResponse, 0, len(conversations))
	for _, conv := range conversations {
		// Determine the other user
		var otherUser models.User
		if conv.Swap.RequesterID == userID {
			otherUser = conv.Swap.Responder
		} else {
			otherUser = conv.Swap.Requester
		}

		resp := models.ConversationResponse{
			ConversationID: conv.ConversationID,
			SwapID:         conv.SwapID,
			OtherUser: models.ConversationUser{
				UserID:   otherUser.UserID,
				Name:     otherUser.Name,
				HasPhoto: otherUser.PhotoData != nil && len(otherUser.PhotoData) > 0,
			},
			OfferedSkill: conv.Swap.OfferedSkill.Name,
			WantedSkill:  conv.Swap.WantedSkill.Name,
			CreatedAt:    conv.CreatedAt,
		}

		// Unread count
		unread, err := s.repo.GetUnreadCount(userID, conv.ConversationID)
		if err == nil {
			resp.UnreadCount = unread
		}

		// Last message preview
		msgs, err := s.repo.GetMessages(conv.ConversationID, nil, 1)
		if err == nil && len(msgs) > 0 {
			last := msgs[len(msgs)-1]
			var preview string
			if last.Encrypted {
				preview = "\U0001f510 Encrypted message"
			} else {
				preview = last.Content
				if len(preview) > 100 {
					preview = preview[:100] + "\u2026"
				}
			}
			resp.LastMessage = &models.MessagePreview{
				Content:   preview,
				SenderID:  last.SenderID,
				CreatedAt: last.CreatedAt,
			}
		}

		responses = append(responses, resp)
	}

	return responses, nil
}

// --- Messages ---

// chatImageURLPattern extracts image IDs from chat image URLs embedded in message HTML.
var chatImageURLPattern = regexp.MustCompile(`/chat/images/([0-9a-fA-F-]{36})`)

// SendMessage creates a new message in a conversation. Validates participation
// and updates the conversation's last_message_at timestamp.
// imageIDs allows explicit image linking for encrypted messages where content is opaque.
func (s *chatService) SendMessage(conversationID, senderID uuid.UUID, content string, encrypted bool, imageIDs []uuid.UUID) (*models.Message, error) {
	ok, err := s.repo.IsParticipant(senderID, conversationID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("user is not a participant of this conversation: %w", apperrors.ErrNotParticipant)
	}

	// Block check: if either party has blocked the other, reject the message.
	// Two-participant model — fetch the other side via the swap_request row.
	var otherID uuid.UUID
	if err := s.db.Raw(`
		SELECT CASE WHEN sr.requester_id = ? THEN sr.responder_id ELSE sr.requester_id END
		FROM conversations c
		JOIN swap_requests sr ON sr.swap_id = c.swap_id
		WHERE c.conversation_id = ?
	`, senderID, conversationID).Scan(&otherID).Error; err == nil && otherID != uuid.Nil {
		var blockCount int64
		if err := s.db.Model(&models.UserBlock{}).
			Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
				senderID, otherID, otherID, senderID).
			Count(&blockCount).Error; err == nil && blockCount > 0 {
			return nil, fmt.Errorf("cannot send: conversation is blocked: %w", apperrors.ErrForbidden)
		}
	}

	msg := &models.Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		Encrypted:      encrypted,
		HasImages:      len(imageIDs) > 0,
	}
	if err := s.repo.CreateMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	if encrypted {
		// For encrypted messages, use the explicitly provided image IDs
		if len(imageIDs) > 0 {
			_ = s.repo.LinkImagesToMessage(imageIDs, msg.MessageID)
		}
	} else {
		// For plaintext HTML, extract image IDs from content via regex
		if matches := chatImageURLPattern.FindAllStringSubmatch(content, -1); len(matches) > 0 {
			var extractedIDs []uuid.UUID
			for _, m := range matches {
				if id, err := uuid.Parse(m[1]); err == nil {
					extractedIDs = append(extractedIDs, id)
				}
			}
			if len(extractedIDs) > 0 {
				_ = s.repo.LinkImagesToMessage(extractedIDs, msg.MessageID)
				msg.HasImages = true
			}
		}
	}

	// Update conversation timestamp
	_ = s.repo.UpdateLastMessageAt(conversationID, msg.CreatedAt)

	// Reload with sender info
	reloaded, _ := s.repo.GetMessageByID(msg.MessageID)
	if reloaded != nil {
		msg = reloaded
	}

	return msg, nil
}

// GetMessages retrieves paginated messages for a conversation after verifying participation.
func (s *chatService) GetMessages(conversationID, userID uuid.UUID, before *time.Time, limit int) ([]models.Message, error) {
	ok, err := s.repo.IsParticipant(userID, conversationID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("user is not a participant of this conversation: %w", apperrors.ErrNotParticipant)
	}
	return s.repo.GetMessages(conversationID, before, limit)
}

// EditMessage updates a message's content. Only the sender can edit, and only
// within MaxEditWindow of the original send time.
func (s *chatService) EditMessage(messageID, userID uuid.UUID, content string) error {
	msg, err := s.repo.GetMessageByID(messageID)
	if err != nil {
		return err
	}
	if msg.SenderID != userID {
		return fmt.Errorf("only the sender can edit a message: %w", apperrors.ErrForbidden)
	}
	if time.Since(msg.CreatedAt) > MaxEditWindow {
		return fmt.Errorf("messages can only be edited within %v of sending: %w", MaxEditWindow, apperrors.ErrForbidden)
	}
	return s.repo.UpdateMessageContent(messageID, content)
}

// DeleteMessage soft-deletes a message. Only the sender can delete their own messages.
func (s *chatService) DeleteMessage(messageID, userID uuid.UUID) error {
	msg, err := s.repo.GetMessageByID(messageID)
	if err != nil {
		return err
	}
	if msg.SenderID != userID {
		return fmt.Errorf("only the sender can delete a message: %w", apperrors.ErrForbidden)
	}
	return s.repo.DeleteMessage(messageID)
}

// --- Read Status ---

func (s *chatService) MarkConversationRead(conversationID, userID uuid.UUID) error {
	ok, err := s.repo.IsParticipant(userID, conversationID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("user is not a participant of this conversation: %w", apperrors.ErrNotParticipant)
	}
	return s.repo.UpsertReadStatus(userID, conversationID)
}

func (s *chatService) GetTotalUnreadCount(userID uuid.UUID) (int64, error) {
	return s.repo.GetTotalUnreadCount(userID)
}

// --- Images ---

// UploadChatImage validates and stores an image uploaded in chat.
// The image is not yet linked to a message — that happens when the message is sent.
// When encrypted=true, the data is opaque ciphertext — skip MIME validation and store as application/octet-stream.
func (s *chatService) UploadChatImage(uploaderID uuid.UUID, data []byte, mimeType string, encrypted bool) (*models.ChatImage, error) {
	if len(data) > MaxImageSize {
		return nil, fmt.Errorf("image exceeds maximum size of %dMB: %w", MaxImageSize/(1024*1024), apperrors.ErrFileTooLarge)
	}

	if encrypted {
		// Encrypted images are opaque ciphertext — don't validate content
		mimeType = "application/octet-stream"
	} else {
		// Validate MIME type for plaintext uploads
		allowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
		}
		if !allowedTypes[mimeType] {
			return nil, fmt.Errorf("unsupported image type %s: %w", mimeType, apperrors.ErrInvalidFileType)
		}
	}

	img := &models.ChatImage{
		UploaderID: uploaderID,
		ImageData:  data,
		MimeType:   mimeType,
		FileSize:   len(data),
	}
	if err := s.repo.CreateChatImage(img); err != nil {
		return nil, fmt.Errorf("failed to store chat image: %w", err)
	}

	return img, nil
}

// GetChatImage retrieves a chat image. Access is granted if the image's uploader
// is the requesting user, or if the image belongs to a conversation the user participates in.
func (s *chatService) GetChatImage(imageID, userID uuid.UUID) (*models.ChatImage, error) {
	img, err := s.repo.GetChatImage(imageID)
	if err != nil {
		return nil, err
	}

	// Uploader can always access their own images
	if img.UploaderID == userID {
		return img, nil
	}

	// Otherwise, check that the image is in a conversation the user participates in
	if img.MessageID != nil {
		msg, err := s.repo.GetMessageByID(*img.MessageID)
		if err != nil {
			return nil, err
		}
		ok, err := s.repo.IsParticipant(userID, msg.ConversationID)
		if err != nil {
			return nil, err
		}
		if ok {
			return img, nil
		}
	}

	return nil, fmt.Errorf("access denied to this image: %w", apperrors.ErrForbidden)
}

// GetChatImagePublic retrieves a chat image without auth checks.
// Used by the public image-serving endpoint; image IDs are unguessable UUIDs.
func (s *chatService) GetChatImagePublic(imageID uuid.UUID) (*models.ChatImage, error) {
	return s.repo.GetChatImage(imageID)
}

// --- Swap Completion ---

// MarkSwapComplete marks the calling user's side of a swap as complete.
// When both participants have marked complete, the swap status changes to "completed"
// and a notification is sent.
func (s *chatService) MarkSwapComplete(swapID, userID uuid.UUID) (*models.SwapRequest, error) {
	var swap models.SwapRequest
	err := s.db.Preload("OfferedSkill").Preload("WantedSkill").
		Where("swap_id = ?", swapID).First(&swap).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("swap request %w", apperrors.ErrNotFound)
		}
		return nil, err
	}

	if swap.Status != models.StatusAccepted {
		return nil, fmt.Errorf("can only complete accepted swaps: %w", apperrors.ErrWrongStatus)
	}

	isRequester := swap.RequesterID == userID
	isResponder := swap.ResponderID == userID
	if !isRequester && !isResponder {
		return nil, fmt.Errorf("only swap participants can mark completion: %w", apperrors.ErrNotParticipant)
	}

	// Set this user's completion flag
	updates := map[string]interface{}{}
	if isRequester {
		if swap.RequesterCompleted {
			return &swap, nil // Already marked
		}
		updates["requester_completed"] = true
	} else {
		if swap.ResponderCompleted {
			return &swap, nil
		}
		updates["responder_completed"] = true
	}

	// Check if both sides are now complete
	bothComplete := false
	if isRequester && swap.ResponderCompleted {
		bothComplete = true
	} else if isResponder && swap.RequesterCompleted {
		bothComplete = true
	}

	if bothComplete {
		updates["status"] = models.StatusCompleted
	}

	if err := s.db.Model(&models.SwapRequest{}).Where("swap_id = ?", swapID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update swap completion: %w", err)
	}

	// Reload
	if err := s.db.Preload("Requester").Preload("Responder").
		Preload("OfferedSkill").Preload("WantedSkill").
		Where("swap_id = ?", swapID).First(&swap).Error; err != nil {
		return nil, err
	}

	// Notify the other participant
	if s.notificationService != nil {
		otherUserID := swap.ResponderID
		if isResponder {
			otherUserID = swap.RequesterID
		}

		if bothComplete {
			_ = s.notificationService.CreateSwapStatusNotification(
				swap.RequesterID, swap.SwapID, "completed", swap.OfferedSkill.Name,
			)
			_ = s.notificationService.CreateSwapStatusNotification(
				swap.ResponderID, swap.SwapID, "completed", swap.OfferedSkill.Name,
			)
		} else {
			// Notify the other user that their partner marked complete
			var marker models.User
			if err := s.db.Select("name").First(&marker, "user_id = ?", userID).Error; err == nil {
				_, _ = s.notificationService.CreateNotification(&models.NotificationRequest{
					UserID:    otherUserID,
					Type:      models.NotificationTypeSwapCompleted,
					Title:     "Partner Marked Swap Complete",
					Message:   fmt.Sprintf("%s marked your skill swap as complete. Mark your side to finish!", marker.Name),
					RelatedID: &swap.SwapID,
				})
			}
		}
	}

	return &swap, nil
}

// UndoSwapComplete allows a user to retract their completion mark
// before both parties have confirmed.
func (s *chatService) UndoSwapComplete(swapID, userID uuid.UUID) (*models.SwapRequest, error) {
	var swap models.SwapRequest
	err := s.db.Where("swap_id = ?", swapID).First(&swap).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("swap request %w", apperrors.ErrNotFound)
		}
		return nil, err
	}

	if swap.Status != models.StatusAccepted {
		return nil, fmt.Errorf("can only undo completion on accepted swaps: %w", apperrors.ErrWrongStatus)
	}

	isRequester := swap.RequesterID == userID
	isResponder := swap.ResponderID == userID
	if !isRequester && !isResponder {
		return nil, fmt.Errorf("only swap participants can undo completion: %w", apperrors.ErrNotParticipant)
	}

	updates := map[string]interface{}{}
	if isRequester {
		updates["requester_completed"] = false
	} else {
		updates["responder_completed"] = false
	}

	if err := s.db.Model(&models.SwapRequest{}).Where("swap_id = ?", swapID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to undo swap completion: %w", err)
	}

	// Reload
	if err := s.db.Preload("Requester").Preload("Responder").
		Preload("OfferedSkill").Preload("WantedSkill").
		Where("swap_id = ?", swapID).First(&swap).Error; err != nil {
		return nil, err
	}

	return &swap, nil
}
