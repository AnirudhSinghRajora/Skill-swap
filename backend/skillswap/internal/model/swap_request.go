package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SwapStatus string

const (
	StatusPending   SwapStatus = "pending"
	StatusAccepted  SwapStatus = "accepted"
	StatusRejected  SwapStatus = "rejected"
	StatusCancelled SwapStatus = "cancelled"
	StatusCompleted SwapStatus = "completed"
)

type SwapRequest struct {
	SwapID              uuid.UUID      `gorm:"type:uuid;primaryKey;column:swap_id;default:gen_random_uuid()"`
	RequesterID         uuid.UUID      `gorm:"type:uuid;column:requester_id;index"`
	ResponderID         uuid.UUID      `gorm:"type:uuid;column:responder_id;index"`
	OfferedSkillID      uuid.UUID      `gorm:"type:uuid;column:offered_skill_id"`
	WantedSkillID       uuid.UUID      `gorm:"type:uuid;column:wanted_skill_id"`
	Status              SwapStatus     `gorm:"type:swap_status;default:'pending'"`
	RequesterCompleted  bool           `gorm:"column:requester_completed;default:false"`
	ResponderCompleted  bool           `gorm:"column:responder_completed;default:false"`
	NoShowFlag          bool           `gorm:"column:no_show_flag;default:false"`
	NoShowReporterID    *uuid.UUID     `gorm:"type:uuid;column:no_show_reporter_id"`
	NoShowReason        *string        `gorm:"column:no_show_reason;type:text"`
	DisputeReason       *string        `gorm:"column:dispute_reason;type:text"`
	CreatedAt           time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt           time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relations - restored
	Requester    User         `gorm:"foreignKey:RequesterID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Responder    User         `gorm:"foreignKey:ResponderID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	OfferedSkill Skill        `gorm:"foreignKey:OfferedSkillID;references:SkillID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	WantedSkill  Skill        `gorm:"foreignKey:WantedSkillID;references:SkillID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	Ratings      []SwapRating `gorm:"foreignKey:SwapID;references:SwapID"`
}

// BeforeCreate is called by GORM before creating a SwapRequest record
func (s *SwapRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if s.SwapID == uuid.Nil {
		s.SwapID = uuid.New()
	}
	return
}

func (SwapRequest) TableName() string { return "swap_requests" }
