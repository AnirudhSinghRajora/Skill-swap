package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserBlock records that one user has blocked another. Composite PK
// (blocker_id, blocked_id). Either side existing in the table breaks
// communication and visibility in both directions.
type UserBlock struct {
	BlockerID uuid.UUID `gorm:"type:uuid;primaryKey;column:blocker_id"`
	BlockedID uuid.UUID `gorm:"type:uuid;primaryKey;column:blocked_id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (UserBlock) TableName() string { return "user_blocks" }

type ReportStatus string

const (
	ReportStatusOpen      ReportStatus = "open"
	ReportStatusReviewing ReportStatus = "reviewing"
	ReportStatusResolved  ReportStatus = "resolved"
	ReportStatusDismissed ReportStatus = "dismissed"
)

type ReportTargetKind string

const (
	ReportTargetUser    ReportTargetKind = "user"
	ReportTargetMessage ReportTargetKind = "message"
	ReportTargetSwap    ReportTargetKind = "swap"
)

// Report is an abuse report filed by one user against another (or against
// a specific message/swap). Admins resolve via the moderation queue.
type Report struct {
	ReportID       uuid.UUID        `gorm:"type:uuid;primaryKey;column:report_id;default:gen_random_uuid()"`
	ReporterID     uuid.UUID        `gorm:"type:uuid;column:reporter_id;not null;index"`
	TargetUserID   uuid.UUID        `gorm:"type:uuid;column:target_user_id;not null;index"`
	TargetKind     ReportTargetKind `gorm:"column:target_kind;not null"`
	TargetID       *uuid.UUID       `gorm:"type:uuid;column:target_id"`
	Reason         string           `gorm:"column:reason;type:text;not null"`
	Status         ReportStatus     `gorm:"column:status;not null;default:'open'"`
	CreatedAt      time.Time        `gorm:"column:created_at;autoCreateTime"`
	ResolvedAt     *time.Time       `gorm:"column:resolved_at"`
	ResolverID     *uuid.UUID       `gorm:"type:uuid;column:resolver_id"`
	ResolutionNote *string          `gorm:"column:resolution_note;type:text"`
}

func (r *Report) BeforeCreate(tx *gorm.DB) error {
	if r.ReportID == uuid.Nil {
		r.ReportID = uuid.New()
	}
	return nil
}

func (Report) TableName() string { return "reports" }
