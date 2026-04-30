package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChatAudio stores a voice note uploaded inside a chat conversation.
//
// The shape mirrors ChatImage on purpose: the client uploads the audio
// blob first, gets back an audio_id, and then includes that id in the
// chat message it sends. For E2EE conversations the blob is opaque
// ciphertext (the same per-conversation key encrypts both messages and
// attachments) and the server stores it as application/octet-stream.
//
// DurationMs is sent by the client purely as a UX hint (so the receiver
// can render a placeholder waveform/timestamp before the blob arrives);
// we don't trust or recompute it server-side.
type ChatAudio struct {
	AudioID    uuid.UUID  `gorm:"column:audio_id;type:uuid;default:gen_random_uuid();primaryKey" json:"audio_id"`
	MessageID  *uuid.UUID `gorm:"column:message_id;type:uuid;index" json:"message_id,omitempty"`
	UploaderID uuid.UUID  `gorm:"column:uploader_id;type:uuid;not null" json:"uploader_id"`
	AudioData  []byte     `gorm:"column:audio_data;type:bytea;not null" json:"-"`
	MimeType   string     `gorm:"column:mime_type;type:varchar(60);not null" json:"mime_type"`
	DurationMs int32      `gorm:"column:duration_ms;not null;default:0" json:"duration_ms"`
	Encrypted  bool       `gorm:"column:encrypted;not null;default:false" json:"encrypted"`
	FileSize   int        `gorm:"column:file_size;not null" json:"file_size"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	Message  *Message `gorm:"foreignKey:MessageID;references:MessageID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
	Uploader User     `gorm:"foreignKey:UploaderID;references:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (a *ChatAudio) BeforeCreate(tx *gorm.DB) (err error) {
	if a.AudioID == uuid.Nil {
		a.AudioID = uuid.New()
	}
	return
}

func (ChatAudio) TableName() string { return "chat_audio" }
