package service

import (
	"errors"
	"testing"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
)

// ── Mock chat repository (minimal — only what SendMessage needs) ─────────────

type mockChatRepo struct {
	participants  map[uuid.UUID]bool // conversationID → allows
	messages      []*models.Message
	linkedImages  []uuid.UUID
	lastMessageAt time.Time
}

func newMockChatRepo() *mockChatRepo {
	return &mockChatRepo{
		participants: make(map[uuid.UUID]bool),
	}
}

func (m *mockChatRepo) IsParticipant(userID, conversationID uuid.UUID) (bool, error) {
	return m.participants[conversationID], nil
}

func (m *mockChatRepo) CreateMessage(msg *models.Message) error {
	msg.MessageID = uuid.New()
	msg.CreatedAt = time.Now()
	m.messages = append(m.messages, msg)
	return nil
}

func (m *mockChatRepo) GetMessageByID(id uuid.UUID) (*models.Message, error) {
	for _, msg := range m.messages {
		if msg.MessageID == id {
			return msg, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockChatRepo) LinkImagesToMessage(imageIDs []uuid.UUID, messageID uuid.UUID) error {
	m.linkedImages = append(m.linkedImages, imageIDs...)
	return nil
}

func (m *mockChatRepo) UpdateLastMessageAt(conversationID uuid.UUID, t time.Time) error {
	m.lastMessageAt = t
	return nil
}

// Unused methods — satisfy interface
func (m *mockChatRepo) CreateConversation(_ *models.Conversation) error { return nil }
func (m *mockChatRepo) GetConversationByID(_ uuid.UUID) (*models.Conversation, error) {
	return nil, nil
}
func (m *mockChatRepo) GetConversationBySwapID(_ uuid.UUID) (*models.Conversation, error) {
	return nil, nil
}
func (m *mockChatRepo) GetUserConversations(_ uuid.UUID) ([]models.Conversation, error) {
	return nil, nil
}
func (m *mockChatRepo) GetMessages(_ uuid.UUID, _ *time.Time, _ int) ([]models.Message, error) {
	return nil, nil
}
func (m *mockChatRepo) UpdateMessageContent(_ uuid.UUID, _ string) error { return nil }
func (m *mockChatRepo) DeleteMessage(_ uuid.UUID) error                  { return nil }
func (m *mockChatRepo) UpsertReadStatus(_, _ uuid.UUID) error            { return nil }
func (m *mockChatRepo) GetUnreadCount(_, _ uuid.UUID) (int64, error)     { return 0, nil }
func (m *mockChatRepo) GetTotalUnreadCount(_ uuid.UUID) (int64, error)   { return 0, nil }
func (m *mockChatRepo) CreateChatImage(_ *models.ChatImage) error        { return nil }
func (m *mockChatRepo) GetChatImage(_ uuid.UUID) (*models.ChatImage, error) {
	return nil, nil
}
func (m *mockChatRepo) GetConversationParticipantIDs(_ uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	return uuid.Nil, uuid.Nil, nil
}

// ── Tests ────────────────────────────────────────────────────────────────────

func TestSendMessage_Encrypted(t *testing.T) {
	repo := newMockChatRepo()
	svc := NewChatService(repo, nil, nil)

	convID := uuid.New()
	userID := uuid.New()
	repo.participants[convID] = true

	// Encrypted message with base64 ciphertext content
	msg, err := svc.SendMessage(convID, userID, "base64encodedciphertext==", true, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !msg.Encrypted {
		t.Error("expected Encrypted flag to be true")
	}
}

func TestSendMessage_EncryptedSkipsImageLinking(t *testing.T) {
	repo := newMockChatRepo()
	svc := NewChatService(repo, nil, nil)

	convID := uuid.New()
	userID := uuid.New()
	repo.participants[convID] = true

	// Content that would match image URL pattern in plaintext mode
	fakeImageURL := `<img src="/chat/images/` + uuid.New().String() + `">`
	_, err := svc.SendMessage(convID, userID, fakeImageURL, true, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// No images should have been linked because encrypted=true
	if len(repo.linkedImages) > 0 {
		t.Errorf("expected no linked images for encrypted message, got %d", len(repo.linkedImages))
	}
}

func TestSendMessage_PlaintextLinksImages(t *testing.T) {
	repo := newMockChatRepo()
	svc := NewChatService(repo, nil, nil)

	convID := uuid.New()
	userID := uuid.New()
	repo.participants[convID] = true

	imageID := uuid.New()
	content := `<p>Check this out: <img src="/chat/images/` + imageID.String() + `"></p>`
	_, err := svc.SendMessage(convID, userID, content, false, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Image should have been linked in plaintext mode
	if len(repo.linkedImages) != 1 {
		t.Fatalf("expected 1 linked image, got %d", len(repo.linkedImages))
	}
	if repo.linkedImages[0] != imageID {
		t.Errorf("expected linked image %s, got %s", imageID, repo.linkedImages[0])
	}
}

func TestSendMessage_NotParticipant(t *testing.T) {
	repo := newMockChatRepo()
	svc := NewChatService(repo, nil, nil)

	convID := uuid.New()
	userID := uuid.New()
	// NOT adding convID to participants

	_, err := svc.SendMessage(convID, userID, "test", false, nil)
	if err == nil {
		t.Fatal("expected error for non-participant")
	}
	if !errors.Is(err, apperrors.ErrNotParticipant) {
		t.Errorf("expected ErrNotParticipant, got: %v", err)
	}
}

func TestSendMessage_EncryptedPreservesContent(t *testing.T) {
	repo := newMockChatRepo()
	svc := NewChatService(repo, nil, nil)

	convID := uuid.New()
	userID := uuid.New()
	repo.participants[convID] = true

	// Simulate actual ciphertext (random-looking base64)
	ciphertext := "SGVsbG8gV29ybGQhISEhISEhISEhISEhISEh" // base64-ish content
	msg, err := svc.SendMessage(convID, userID, ciphertext, true, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Content should be stored as-is (no server-side processing)
	if msg.Content != ciphertext {
		t.Errorf("expected content preserved as %q, got %q", ciphertext, msg.Content)
	}
}

func TestSendMessage_EncryptedExplicitImageIDs(t *testing.T) {
	repo := newMockChatRepo()
	svc := NewChatService(repo, nil, nil)

	convID := uuid.New()
	userID := uuid.New()
	repo.participants[convID] = true

	imgID1 := uuid.New()
	imgID2 := uuid.New()
	msg, err := svc.SendMessage(convID, userID, "encrypted-blob-with-images", true, []uuid.UUID{imgID1, imgID2})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !msg.HasImages {
		t.Error("expected HasImages to be true")
	}
	if len(repo.linkedImages) != 2 {
		t.Fatalf("expected 2 linked images, got %d", len(repo.linkedImages))
	}
	if repo.linkedImages[0] != imgID1 || repo.linkedImages[1] != imgID2 {
		t.Errorf("expected linked images [%s, %s], got %v", imgID1, imgID2, repo.linkedImages)
	}
}

func TestSendMessage_PlaintextAndEncryptedFlagsPersist(t *testing.T) {
	repo := newMockChatRepo()
	svc := NewChatService(repo, nil, nil)

	convID := uuid.New()
	userID := uuid.New()
	repo.participants[convID] = true

	// Send plaintext
	plain, err := svc.SendMessage(convID, userID, "<p>Hello</p>", false, nil)
	if err != nil {
		t.Fatalf("plaintext send error: %v", err)
	}
	if plain.Encrypted {
		t.Error("plaintext message should not be encrypted")
	}

	// Send encrypted
	enc, err := svc.SendMessage(convID, userID, "encrypted-blob", true, nil)
	if err != nil {
		t.Fatalf("encrypted send error: %v", err)
	}
	if !enc.Encrypted {
		t.Error("encrypted message should have Encrypted=true")
	}

	// Verify both messages stored with correct flags
	if len(repo.messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(repo.messages))
	}
	if repo.messages[0].Encrypted {
		t.Error("first message (plaintext) should have Encrypted=false")
	}
	if !repo.messages[1].Encrypted {
		t.Error("second message (encrypted) should have Encrypted=true")
	}
}
