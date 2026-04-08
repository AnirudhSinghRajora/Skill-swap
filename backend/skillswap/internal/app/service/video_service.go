package service

import (
	"fmt"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
)

type VideoService struct {
	apiKey    string
	apiSecret string
	url       string
}

func NewVideoService(cfg config.Config) *VideoService {
	return &VideoService{
		apiKey:    cfg.LiveKitAPIKey,
		apiSecret: cfg.LiveKitAPISecret,
		url:       cfg.LiveKitURL,
	}
}

// IsConfigured returns true if LiveKit credentials are set
func (s *VideoService) IsConfigured() bool {
	return s.apiKey != "" && s.apiSecret != ""
}

// GetConnectionURL returns the LiveKit WebSocket URL for clients
func (s *VideoService) GetConnectionURL() string {
	return s.url
}

// GenerateToken creates a LiveKit access token for a user to join a room
func (s *VideoService) GenerateToken(userID uuid.UUID, userName, roomName string) (string, error) {
	if !s.IsConfigured() {
		return "", fmt.Errorf("livekit not configured")
	}

	at := auth.NewAccessToken(s.apiKey, s.apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	at.AddGrant(grant).
		SetIdentity(userID.String()).
		SetName(userName).
		SetValidFor(4 * time.Hour)

	return at.ToJWT()
}

// GetRoomName generates a deterministic room name for a conversation
func GetRoomName(conversationID uuid.UUID) string {
	return fmt.Sprintf("skillswap-call-%s", conversationID.String())
}
