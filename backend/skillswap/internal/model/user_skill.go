package models

import "github.com/google/uuid"

// UserSkillOffered represents table user_skills_offered
type UserSkillOffered struct {
	UserID  uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	SkillID uuid.UUID `gorm:"type:uuid;primaryKey;column:skill_id"`
	Level   int16     `gorm:"column:level;not null;default:2"` // 1=novice 2=intermediate 3=advanced 4=expert

	// Relations - restored
	User  User  `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Skill Skill `gorm:"foreignKey:SkillID;references:SkillID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserSkillOffered) TableName() string { return "user_skills_offered" }

// UserSkillWanted represents table user_skills_wanted
type UserSkillWanted struct {
	UserID  uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	SkillID uuid.UUID `gorm:"type:uuid;primaryKey;column:skill_id"`
	Level   int16     `gorm:"column:level;not null;default:2"`

	// Relations - restored
	User  User  `gorm:"foreignKey:UserID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Skill Skill `gorm:"foreignKey:SkillID;references:SkillID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserSkillWanted) TableName() string { return "user_skills_wanted" }
