package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Conversation represents a chat linked to an accepted swap request.
// One conversation per swap — created automatically when a swap is accepted.
type Conversation struct {
	ConversationID uuid.UUID      `gorm:"type:uuid;primaryKey;column:conversation_id;default:gen_random_uuid()"`
	SwapID         uuid.UUID      `gorm:"type:uuid;column:swap_id;uniqueIndex;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	LastMessageAt  *time.Time     `gorm:"column:last_message_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relations
	Swap     SwapRequest `gorm:"foreignKey:SwapID;references:SwapID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Messages []Message   `gorm:"foreignKey:ConversationID;references:ConversationID"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ConversationID == uuid.Nil {
		c.ConversationID = uuid.New()
	}
	return
}

func (Conversation) TableName() string { return "conversations" }

// Message represents a single chat message within a conversation.
type Message struct {
	MessageID      uuid.UUID      `gorm:"type:uuid;primaryKey;column:message_id;default:gen_random_uuid()"`
	ConversationID uuid.UUID      `gorm:"type:uuid;column:conversation_id;index;not null"`
	SenderID       uuid.UUID      `gorm:"type:uuid;column:sender_id;index;not null"`
	Content        string         `gorm:"column:content;type:text;not null"`
	Encrypted      bool           `gorm:"column:encrypted;default:false"`
	HasImages      bool           `gorm:"column:has_images;default:false"`
	IsEdited       bool           `gorm:"column:is_edited;default:false"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      *time.Time     `gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relations
	Conversation Conversation `gorm:"foreignKey:ConversationID;references:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Sender       User         `gorm:"foreignKey:SenderID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Images       []ChatImage  `gorm:"foreignKey:MessageID;references:MessageID"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.MessageID == uuid.Nil {
		m.MessageID = uuid.New()
	}
	return
}

func (Message) TableName() string { return "messages" }

// MessageReadStatus tracks the last time a user read a conversation.
// Any messages created after last_read_at are considered unread.
type MessageReadStatus struct {
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	ConversationID uuid.UUID `gorm:"type:uuid;primaryKey;column:conversation_id"`
	LastReadAt     time.Time `gorm:"column:last_read_at;not null;default:CURRENT_TIMESTAMP"`

	// Relations
	User         User         `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Conversation Conversation `gorm:"foreignKey:ConversationID;references:ConversationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (MessageReadStatus) TableName() string { return "message_read_status" }

// ChatImage stores an image uploaded within a chat conversation.
type ChatImage struct {
	ImageID    uuid.UUID  `gorm:"type:uuid;primaryKey;column:image_id;default:gen_random_uuid()"`
	MessageID  *uuid.UUID `gorm:"type:uuid;column:message_id;index"`
	UploaderID uuid.UUID  `gorm:"type:uuid;column:uploader_id;not null"`
	ImageData  []byte     `gorm:"column:image_data;type:bytea;not null"`
	MimeType   string     `gorm:"column:mime_type;type:varchar(50);not null"`
	FileSize   int        `gorm:"column:file_size;not null"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`

	// Relations
	Message  *Message `gorm:"foreignKey:MessageID;references:MessageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Uploader User     `gorm:"foreignKey:UploaderID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (i *ChatImage) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ImageID == uuid.Nil {
		i.ImageID = uuid.New()
	}
	return
}

func (ChatImage) TableName() string { return "chat_images" }

// --- Response / DTO types ---

// ConversationResponse is the API response for a conversation listing.
type ConversationResponse struct {
	ConversationID uuid.UUID        `json:"conversation_id"`
	SwapID         uuid.UUID        `json:"swap_id"`
	OtherUser      ConversationUser `json:"other_user"`
	OfferedSkill   string           `json:"offered_skill"`
	WantedSkill    string           `json:"wanted_skill"`
	LastMessage    *MessagePreview  `json:"last_message"`
	UnreadCount    int64            `json:"unread_count"`
	CreatedAt      time.Time        `json:"created_at"`
}

// ConversationUser is a minimal user representation for chat context.
type ConversationUser struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	HasPhoto bool      `json:"has_photo"`
}

// MessagePreview is a truncated message for conversation lists.
type MessagePreview struct {
	Content   string    `json:"content"`
	SenderID  uuid.UUID `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
}

// MessageResponse is the API response for a single message.
type MessageResponse struct {
	MessageID      uuid.UUID  `json:"message_id"`
	ConversationID uuid.UUID  `json:"conversation_id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	Content        string     `json:"content"`
	Encrypted      bool       `json:"encrypted"`
	HasImages      bool       `json:"has_images"`
	IsEdited       bool       `json:"is_edited"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

// SendMessageRequest is the incoming request body for sending a message.
type SendMessageRequest struct {
	Content   string   `json:"content" binding:"required,min=1"`
	Encrypted bool     `json:"encrypted"`
	ImageIDs  []string `json:"image_ids,omitempty"`
}

// EditMessageRequest is the incoming request body for editing a message.
type EditMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}
