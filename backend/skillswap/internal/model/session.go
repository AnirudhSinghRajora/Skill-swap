package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionStatus is the lifecycle state of a scheduled session attached to
// an accepted swap. Sessions are created in `proposed` state by either
// participant; the other party either `accept`s or `cancel`s.
type SessionStatus string

const (
	SessionProposed  SessionStatus = "proposed"
	SessionAccepted  SessionStatus = "accepted"
	SessionCancelled SessionStatus = "cancelled"
	SessionCompleted SessionStatus = "completed"
)

// Session is a single planned meet-up between the two parties of a swap.
// LiveKitRoomName is opaque to this package; the video service derives it
// from the session id when joining is requested.
type Session struct {
	SessionID        uuid.UUID      `gorm:"type:uuid;primaryKey;column:session_id;default:gen_random_uuid()"`
	SwapID           uuid.UUID      `gorm:"type:uuid;column:swap_id;index;not null"`
	ProposerID       uuid.UUID      `gorm:"type:uuid;column:proposer_id;index;not null"`
	ScheduledStart   time.Time      `gorm:"column:scheduled_start;not null"`
	ScheduledEnd     time.Time      `gorm:"column:scheduled_end;not null"`
	Status           SessionStatus  `gorm:"column:status;type:varchar(16);not null;default:'proposed'"`
	LiveKitRoomName  *string        `gorm:"column:livekit_room_name;type:varchar(64)"`
	ReminderSentAt   *time.Time     `gorm:"column:reminder_sent_at"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Session) TableName() string { return "sessions" }
