package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UserID             uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id;default:gen_random_uuid()"`
	Name               string    `gorm:"not null"`
	Email              string    `gorm:"uniqueIndex;not null"`
	PasswordHash       string    `gorm:"column:password_hash"`
	Location           *string
	PhotoData          []byte         `gorm:"column:photo_data;type:bytea"`
	PhotoMimeType      *string        `gorm:"column:photo_mime_type"`
	IsPublic           bool           `gorm:"column:is_public;default:true"`
	IsAdmin            bool           `gorm:"column:is_admin;default:false"`
	IsBanned           bool           `gorm:"column:is_banned;default:false"`
	AuthProvider       string         `gorm:"column:auth_provider;default:'local'"`
	GoogleID           *string        `gorm:"column:google_id;uniqueIndex"`
	EmailVerified      bool           `gorm:"column:email_verified;default:false"`
	PublicKey          *string        `gorm:"column:public_key"`
	EncryptedKeyBackup *string        `gorm:"column:encrypted_key_backup"`
	CreatedAt          time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at;index"`

	// Relations - restored
	SkillsOffered     []UserSkillOffered `gorm:"foreignKey:UserID;references:UserID"`
	SkillsWanted      []UserSkillWanted  `gorm:"foreignKey:UserID;references:UserID"`
	AvailabilitySlots []AvailabilitySlot `gorm:"foreignKey:UserID;references:UserID"`
}

// BeforeCreate is called by GORM before creating a User record
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UserID == uuid.Nil {
		u.UserID = uuid.New()
	}
	return
}

func (User) TableName() string { return "users" }
