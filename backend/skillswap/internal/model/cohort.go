package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CohortKind distinguishes a small private learning cohort from a larger
// public drop-in room. Both share the same schema; the kind drives the
// default capacity and discoverability rules in the service layer.
type CohortKind string

const (
	CohortKindCohort CohortKind = "cohort" // small, host-curated
	CohortKindRoom   CohortKind = "room"   // public, drop-in
)

// CohortStatus is the lifecycle of a cohort.
type CohortStatus string

const (
	CohortOpen       CohortStatus = "open"
	CohortInProgress CohortStatus = "in_progress"
	CohortCompleted  CohortStatus = "completed"
	CohortCancelled  CohortStatus = "cancelled"
)

// CohortRole is the membership role inside a cohort.
type CohortRole string

const (
	CohortRoleHost   CohortRole = "host"
	CohortRoleMember CohortRole = "member"
)

// Cohort is a multi-participant learning group hosted by one user around a
// single skill. Sessions within a cohort live in cohort_sessions.
type Cohort struct {
	CohortID        uuid.UUID      `gorm:"type:uuid;primaryKey;column:cohort_id;default:gen_random_uuid()"`
	HostID          uuid.UUID      `gorm:"type:uuid;column:host_id;index;not null"`
	SkillID         uuid.UUID      `gorm:"type:uuid;column:skill_id;index;not null"`
	Title           string         `gorm:"column:title;type:varchar(120);not null"`
	Description     string         `gorm:"column:description;type:text"`
	Capacity        int16          `gorm:"column:capacity;not null;default:4"`
	Kind            CohortKind     `gorm:"column:kind;type:varchar(16);not null;default:'cohort'"`
	IsPublic        bool           `gorm:"column:is_public;not null;default:true"`
	Status          CohortStatus   `gorm:"column:status;type:varchar(16);not null;default:'open'"`
	SchedulePattern *string        `gorm:"column:schedule_pattern;type:varchar(120)"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Cohort) TableName() string { return "cohorts" }

// CohortMember is a (cohort, user) row. Composite primary key prevents
// double-joins. The host appears here too, with role=host.
type CohortMember struct {
	CohortID uuid.UUID  `gorm:"type:uuid;primaryKey;column:cohort_id"`
	UserID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:user_id"`
	Role     CohortRole `gorm:"column:role;type:varchar(16);not null;default:'member'"`
	JoinedAt time.Time  `gorm:"column:joined_at;autoCreateTime"`
}

func (CohortMember) TableName() string { return "cohort_members" }

// CohortSession is one scheduled meet-up of a cohort. Mirrors the
// individual `sessions` shape but keyed to a cohort instead of a swap.
type CohortSession struct {
	CohortSessionID uuid.UUID      `gorm:"type:uuid;primaryKey;column:cohort_session_id;default:gen_random_uuid()"`
	CohortID        uuid.UUID      `gorm:"type:uuid;column:cohort_id;index;not null"`
	ScheduledStart  time.Time      `gorm:"column:scheduled_start;not null"`
	ScheduledEnd    time.Time      `gorm:"column:scheduled_end;not null"`
	LiveKitRoomName *string        `gorm:"column:livekit_room_name;type:varchar(64)"`
	Status          string         `gorm:"column:status;type:varchar(16);not null;default:'scheduled'"`
	CreatedAt       time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (CohortSession) TableName() string { return "cohort_sessions" }
