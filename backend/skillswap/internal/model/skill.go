package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Skill struct {
	SkillID    uuid.UUID  `gorm:"type:uuid;primaryKey;column:skill_id;default:gen_random_uuid()"`
	Name       string     `gorm:"uniqueIndex;not null"`
	CategoryID *uuid.UUID `gorm:"type:uuid;column:category_id"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
}

// BeforeCreate is called by GORM before creating a Skill record
func (s *Skill) BeforeCreate(tx *gorm.DB) (err error) {
	if s.SkillID == uuid.Nil {
		s.SkillID = uuid.New()
	}
	return
}

func (Skill) TableName() string { return "skills" }

// SkillCategory groups skills hierarchically.
type SkillCategory struct {
	CategoryID uuid.UUID  `gorm:"type:uuid;primaryKey;column:category_id;default:gen_random_uuid()"`
	Name       string     `gorm:"column:name;uniqueIndex;not null"`
	Slug       string     `gorm:"column:slug;uniqueIndex;not null"`
	ParentID   *uuid.UUID `gorm:"type:uuid;column:parent_id"`
	SortOrder  int        `gorm:"column:sort_order;default:0"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
}

func (c *SkillCategory) BeforeCreate(tx *gorm.DB) (err error) {
	if c.CategoryID == uuid.Nil {
		c.CategoryID = uuid.New()
	}
	return
}

func (SkillCategory) TableName() string { return "skill_categories" }

// SkillAlias maps an alternative name to a canonical skill.
type SkillAlias struct {
	AliasID          uuid.UUID `gorm:"type:uuid;primaryKey;column:alias_id;default:gen_random_uuid()"`
	AliasName        string    `gorm:"column:alias_name;uniqueIndex;not null"`
	CanonicalSkillID uuid.UUID `gorm:"type:uuid;column:canonical_skill_id;not null"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (a *SkillAlias) BeforeCreate(tx *gorm.DB) (err error) {
	if a.AliasID == uuid.Nil {
		a.AliasID = uuid.New()
	}
	return
}

func (SkillAlias) TableName() string { return "skill_aliases" }
