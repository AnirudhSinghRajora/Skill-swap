package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationTypeSwapRequest   NotificationType = "swap_request"
	NotificationTypeSwapAccepted  NotificationType = "swap_accepted"
	NotificationTypeSwapRejected  NotificationType = "swap_rejected"
	NotificationTypeSwapCompleted NotificationType = "swap_completed"
	NotificationTypeNewRating     NotificationType = "new_rating"
	NotificationTypeSkillMatched  NotificationType = "skill_matched"
	NotificationTypeSystemAlert   NotificationType = "system_alert"
	NotificationTypeAdminNotice   NotificationType = "admin_notice"
)

type Notification struct {
	NotificationID uuid.UUID        `gorm:"type:uuid;primaryKey;column:notification_id;default:gen_random_uuid()"`
	UserID         uuid.UUID        `gorm:"type:uuid;not null;index"`
	Type           NotificationType `gorm:"column:type;not null"`
	Title          string           `gorm:"column:title;not null"`
	Message        string           `gorm:"column:message;not null"`
	IsRead         bool             `gorm:"column:is_read;default:false"`
	RelatedID      *uuid.UUID       `gorm:"type:uuid;column:related_id"` // ID of related entity (swap, rating, etc.)
	CreatedAt      time.Time        `gorm:"column:created_at;autoCreateTime"`
	DeletedAt      gorm.DeletedAt   `gorm:"column:deleted_at;index"`

	// Relations
	User User `gorm:"foreignKey:UserID;references:UserID"`
}

// BeforeCreate is called by GORM before creating a Notification record
func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if n.NotificationID == uuid.Nil {
		n.NotificationID = uuid.New()
	}
	return
}

func (Notification) TableName() string { return "notifications" }

// NotificationRequest represents a request to create a notification
type NotificationRequest struct {
	UserID    uuid.UUID        `json:"user_id" binding:"required"`
	Type      NotificationType `json:"type" binding:"required"`
	Title     string           `json:"title" binding:"required,min=1,max=255"`
	Message   string           `json:"message" binding:"required,min=1,max=1000"`
	RelatedID *uuid.UUID       `json:"related_id,omitempty"`
}

// NotificationResponse represents a notification response
type NotificationResponse struct {
	NotificationID uuid.UUID        `json:"notification_id"`
	Type           NotificationType `json:"type"`
	Title          string           `json:"title"`
	Message        string           `json:"message"`
	IsRead         bool             `json:"is_read"`
	RelatedID      *uuid.UUID       `json:"related_id,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
}

// MarkAsReadRequest represents a request to mark notifications as read
type MarkAsReadRequest struct {
	NotificationIDs []uuid.UUID `json:"notification_ids" binding:"required"`
}

// NotificationStatsResponse represents notification statistics
type NotificationStatsResponse struct {
	TotalNotifications int64 `json:"total_notifications"`
	UnreadCount        int64 `json:"unread_count"`
	ReadCount          int64 `json:"read_count"`
}
