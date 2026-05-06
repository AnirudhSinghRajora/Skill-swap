package service

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/repository"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
)

// ── Mock repository ──────────────────────────────────────────────────────────

type mockUserRepo struct {
	users map[uuid.UUID]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[uuid.UUID]*models.User)}
}

func (m *mockUserRepo) addUser(u *models.User) {
	m.users[u.UserID] = u
}

func (m *mockUserRepo) Create(user *models.User) error {
	m.users[user.UserID] = user
	return nil
}

func (m *mockUserRepo) GetByID(id uuid.UUID) (*models.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByEmail(email string) (*models.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockUserRepo) Update(user *models.User) error {
	m.users[user.UserID] = user
	return nil
}

func (m *mockUserRepo) Delete(id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) List(limit, offset int, filters repository.UserFilters) ([]*models.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepo) UpdateE2EEKeys(userID uuid.UUID, publicKey, encryptedBackup string) error {
	u, ok := m.users[userID]
	if !ok {
		return apperrors.ErrNotFound
	}
	u.PublicKey = &publicKey
	u.EncryptedKeyBackup = &encryptedBackup
	return nil
}

func (m *mockUserRepo) GetPublicKey(userID uuid.UUID) (string, error) {
	u, ok := m.users[userID]
	if !ok {
		return "", apperrors.ErrNotFound
	}
	if u.PublicKey == nil || *u.PublicKey == "" {
		return "", apperrors.ErrNotFound
	}
	return *u.PublicKey, nil
}

func (m *mockUserRepo) GetKeyBackup(userID uuid.UUID) (string, error) {
	u, ok := m.users[userID]
	if !ok {
		return "", apperrors.ErrNotFound
	}
	if u.EncryptedKeyBackup == nil || *u.EncryptedKeyBackup == "" {
		return "", apperrors.ErrNotFound
	}
	return *u.EncryptedKeyBackup, nil
}

// ── SetE2EEKeys tests ────────────────────────────────────────────────────────

func TestSetE2EEKeys_ValidPublicKey(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Alice", Email: "alice@test.com", PasswordHash: "hash"})

	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	backup := "some-encrypted-backup-blob"

	err := svc.SetE2EEKeys(userID, validKey, backup)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	stored, err := svc.GetPublicKey(userID)
	if err != nil {
		t.Fatalf("expected no error fetching public key, got: %v", err)
	}
	if stored != validKey {
		t.Errorf("expected public key %q, got %q", validKey, stored)
	}
}

func TestSetE2EEKeys_InvalidBase64(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Bob", Email: "bob@test.com", PasswordHash: "hash"})

	err := svc.SetE2EEKeys(userID, "!!not-valid-base64!!", "backup")
	if err == nil {
		t.Fatal("expected validation error for invalid base64")
	}
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Errorf("expected ErrValidation, got: %v", err)
	}
}

func TestSetE2EEKeys_WrongKeyLength(t *testing.T) {
	tests := []struct {
		name   string
		keyLen int
	}{
		{"too short (16 bytes)", 16},
		{"too long (64 bytes)", 64},
		{"empty key (0 bytes)", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockUserRepo()
			svc := NewUserService(repo)

			userID := uuid.New()
			repo.addUser(&models.User{UserID: userID, Name: "Test", Email: "test@test.com", PasswordHash: "hash"})

			key := base64.StdEncoding.EncodeToString(make([]byte, tt.keyLen))
			err := svc.SetE2EEKeys(userID, key, "backup")
			if err == nil {
				t.Fatal("expected validation error for wrong key length")
			}
			if !errors.Is(err, apperrors.ErrValidation) {
				t.Errorf("expected ErrValidation, got: %v", err)
			}
		})
	}
}

func TestSetE2EEKeys_EmptyBackup(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Test", Email: "test@test.com", PasswordHash: "hash"})

	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	err := svc.SetE2EEKeys(userID, validKey, "")
	if err == nil {
		t.Fatal("expected validation error for empty backup")
	}
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Errorf("expected ErrValidation, got: %v", err)
	}
}

func TestSetE2EEKeys_BackupTooLarge(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Test", Email: "test@test.com", PasswordHash: "hash"})

	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	hugeBackup := strings.Repeat("A", 513)
	err := svc.SetE2EEKeys(userID, validKey, hugeBackup)
	if err == nil {
		t.Fatal("expected validation error for oversized backup")
	}
	if !errors.Is(err, apperrors.ErrValidation) {
		t.Errorf("expected ErrValidation, got: %v", err)
	}
}

func TestSetE2EEKeys_NonexistentUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	err := svc.SetE2EEKeys(uuid.New(), validKey, "backup")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSetE2EEKeys_KeyRotation(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Alice", Email: "alice@test.com", PasswordHash: "hash"})

	key1 := base64.StdEncoding.EncodeToString(make([]byte, 32))
	if err := svc.SetE2EEKeys(userID, key1, "backup1"); err != nil {
		t.Fatalf("first SetE2EEKeys failed: %v", err)
	}

	newKeyBytes := make([]byte, 32)
	newKeyBytes[0] = 0x42
	key2 := base64.StdEncoding.EncodeToString(newKeyBytes)
	if err := svc.SetE2EEKeys(userID, key2, "backup2"); err != nil {
		t.Fatalf("key rotation SetE2EEKeys failed: %v", err)
	}

	stored, err := svc.GetPublicKey(userID)
	if err != nil {
		t.Fatalf("GetPublicKey failed: %v", err)
	}
	if stored != key2 {
		t.Errorf("expected rotated key %q, got %q", key2, stored)
	}
}

// ── GetPublicKey tests ───────────────────────────────────────────────────────

func TestGetPublicKey_NoKeySet(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Test", Email: "test@test.com", PasswordHash: "hash"})

	_, err := svc.GetPublicKey(userID)
	if err == nil {
		t.Fatal("expected error when no public key is set")
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetPublicKey_NonexistentUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.GetPublicKey(uuid.New())
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

// ── GetKeyBackup tests ───────────────────────────────────────────────────────

func TestGetKeyBackup_Success(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Alice", Email: "alice@test.com", PasswordHash: "hash"})

	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))
	if err := svc.SetE2EEKeys(userID, validKey, "my-backup-blob"); err != nil {
		t.Fatalf("SetE2EEKeys failed: %v", err)
	}

	backup, err := svc.GetKeyBackup(userID)
	if err != nil {
		t.Fatalf("GetKeyBackup failed: %v", err)
	}
	if backup != "my-backup-blob" {
		t.Errorf("expected backup %q, got %q", "my-backup-blob", backup)
	}
}

func TestGetKeyBackup_NoBackup(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Test", Email: "test@test.com", PasswordHash: "hash"})

	_, err := svc.GetKeyBackup(userID)
	if err == nil {
		t.Fatal("expected error when no backup exists")
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestGetKeyBackup_NonexistentUser(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	_, err := svc.GetKeyBackup(uuid.New())
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestSetE2EEKeys_ExactBoundaryBackupLength(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewUserService(repo)

	userID := uuid.New()
	repo.addUser(&models.User{UserID: userID, Name: "Test", Email: "test@test.com", PasswordHash: "hash"})

	validKey := base64.StdEncoding.EncodeToString(make([]byte, 32))

	// Exactly 512 characters — should succeed
	backup512 := strings.Repeat("A", 512)
	if err := svc.SetE2EEKeys(userID, validKey, backup512); err != nil {
		t.Fatalf("expected 512-char backup to succeed, got: %v", err)
	}

	// 513 characters — should fail
	backup513 := strings.Repeat("A", 513)
	if err := svc.SetE2EEKeys(userID, validKey, backup513); err == nil {
		t.Fatal("expected 513-char backup to fail")
	}
}
