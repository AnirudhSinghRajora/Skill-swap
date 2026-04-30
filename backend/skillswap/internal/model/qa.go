package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VoteTargetKind enumerates the entity types that can be voted on.
// Stored as a varchar in the votes table so we can grow it later
// (comments, answers-on-answers, ...) without an enum migration.
type VoteTargetKind string

const (
	VoteTargetQuestion VoteTargetKind = "question"
	VoteTargetAnswer   VoteTargetKind = "answer"
)

// Question is a top-level Q&A entry, optionally tagged to a skill.
type Question struct {
	QuestionID   uuid.UUID      `gorm:"column:question_id;type:uuid;default:gen_random_uuid();primaryKey" json:"question_id"`
	AskerID      uuid.UUID      `gorm:"column:asker_id;type:uuid;not null;index" json:"asker_id"`
	SkillID      *uuid.UUID     `gorm:"column:skill_id;type:uuid;index" json:"skill_id,omitempty"`
	Title        string         `gorm:"column:title;type:varchar(200);not null" json:"title"`
	Body         string         `gorm:"column:body;type:text;not null" json:"body"`
	UpvoteCount  int32          `gorm:"column:upvote_count;not null;default:0" json:"upvote_count"`
	AnswerCount  int32          `gorm:"column:answer_count;not null;default:0" json:"answer_count"`
	AcceptedID   *uuid.UUID     `gorm:"column:accepted_answer_id;type:uuid" json:"accepted_answer_id,omitempty"`
	CreatedAt    time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Question) TableName() string { return "questions" }

// Answer is a reply to a question. The single accepted answer per question
// is tracked on the Question via AcceptedID rather than denormalised here,
// so unsetting "accepted" only updates one row.
type Answer struct {
	AnswerID    uuid.UUID      `gorm:"column:answer_id;type:uuid;default:gen_random_uuid();primaryKey" json:"answer_id"`
	QuestionID  uuid.UUID      `gorm:"column:question_id;type:uuid;not null;index" json:"question_id"`
	AuthorID    uuid.UUID      `gorm:"column:author_id;type:uuid;not null;index" json:"author_id"`
	Body        string         `gorm:"column:body;type:text;not null" json:"body"`
	UpvoteCount int32          `gorm:"column:upvote_count;not null;default:0" json:"upvote_count"`
	CreatedAt   time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

func (Answer) TableName() string { return "answers" }

// Vote is one user's upvote on either a Question or an Answer. We rely on
// a composite UNIQUE(voter_id, target_kind, target_id) constraint plus the
// service layer's "delete-if-exists" logic to give us toggle semantics
// without a separate vote_state column.
type Vote struct {
	VoteID     uuid.UUID      `gorm:"column:vote_id;type:uuid;default:gen_random_uuid();primaryKey" json:"vote_id"`
	VoterID    uuid.UUID      `gorm:"column:voter_id;type:uuid;not null;index" json:"voter_id"`
	TargetKind VoteTargetKind `gorm:"column:target_kind;type:varchar(16);not null" json:"target_kind"`
	TargetID   uuid.UUID      `gorm:"column:target_id;type:uuid;not null" json:"target_id"`
	CreatedAt  time.Time      `gorm:"column:created_at" json:"created_at"`
}

func (Vote) TableName() string { return "votes" }
