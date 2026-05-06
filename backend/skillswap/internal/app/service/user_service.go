package service

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/repository"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
)

type UserService interface {
	GetProfile(userID uuid.UUID) (*UserProfileResponse, error)
	GetPublicProfile(userID uuid.UUID) (*UserProfileResponse, error)
	UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) error
	SearchUsers(req *SearchUsersRequest) (*SearchUsersResponse, error)

	// E2EE key management
	SetE2EEKeys(userID uuid.UUID, publicKey, encryptedBackup string) error
	GetPublicKey(userID uuid.UUID) (string, error)
	GetKeyBackup(userID uuid.UUID) (string, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// DTOs for API responses
type UserProfileResponse struct {
	UserID        uuid.UUID       `json:"user_id"`
	Name          string          `json:"name"`
	Email         string          `json:"email"`
	Location      *string         `json:"location"`
	HasPhoto      bool            `json:"has_photo"`
	IsPublic      bool            `json:"is_public"`
	PublicKey     *string         `json:"public_key,omitempty"`
	HasKeyBackup  bool            `json:"has_key_backup"`
	SkillsOffered []SkillResponse `json:"skills_offered"`
	SkillsWanted  []SkillResponse `json:"skills_wanted"`
	CreatedAt     time.Time       `json:"created_at"`
}

type SkillResponse struct {
	SkillID uuid.UUID `json:"skill_id"`
	Name    string    `json:"name"`
}

type UpdateProfileRequest struct {
	Name     *string `json:"name,omitempty"`
	Location *string `json:"location,omitempty"`
	IsPublic *bool   `json:"is_public,omitempty"`
}

type SearchUsersRequest struct {
	Location   string `json:"location,omitempty"`
	SearchTerm string `json:"search_term,omitempty"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
}

type SearchUsersResponse struct {
	Users      []UserProfileResponse `json:"users"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"total_pages"`
}

func (s *userService) GetProfile(userID uuid.UUID) (*UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return s.toUserProfileResponse(user), nil
}

func (s *userService) GetPublicProfile(userID uuid.UUID) (*UserProfileResponse, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if !user.IsPublic {
		return nil, fmt.Errorf("user profile is not public: %w", apperrors.ErrForbidden)
	}

	return s.toUserProfileResponse(user), nil
}

func (s *userService) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		name := *req.Name
		if len(name) < 2 || len(name) > 100 {
			return fmt.Errorf("name must be between 2 and 100 characters: %w", apperrors.ErrValidation)
		}
		user.Name = name
	}
	if req.Location != nil {
		if len(*req.Location) > 200 {
			return fmt.Errorf("location must be at most 200 characters: %w", apperrors.ErrValidation)
		}
		user.Location = req.Location
	}
	if req.IsPublic != nil {
		user.IsPublic = *req.IsPublic
	}

	return s.userRepo.Update(user)
}

func (s *userService) SearchUsers(req *SearchUsersRequest) (*SearchUsersResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page < 1 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit

	filters := repository.UserFilters{
		IsPublic:   boolPtr(true), // Only show public profiles
		Location:   req.Location,
		SearchTerm: req.SearchTerm,
	}

	users, total, err := s.userRepo.List(req.Limit, offset, filters)
	if err != nil {
		return nil, err
	}

	userResponses := make([]UserProfileResponse, len(users))
	for i, user := range users {
		userResponses[i] = *s.toUserProfileResponse(user)
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &SearchUsersResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

// Helper functions
func (s *userService) toUserProfileResponse(user *models.User) *UserProfileResponse {
	skillsOffered := make([]SkillResponse, len(user.SkillsOffered))
	for i, skill := range user.SkillsOffered {
		skillsOffered[i] = SkillResponse{
			SkillID: skill.Skill.SkillID,
			Name:    skill.Skill.Name,
		}
	}

	skillsWanted := make([]SkillResponse, len(user.SkillsWanted))
	for i, skill := range user.SkillsWanted {
		skillsWanted[i] = SkillResponse{
			SkillID: skill.Skill.SkillID,
			Name:    skill.Skill.Name,
		}
	}

	return &UserProfileResponse{
		UserID:        user.UserID,
		Name:          user.Name,
		Email:         user.Email,
		Location:      user.Location,
		HasPhoto:      len(user.PhotoData) > 0,
		IsPublic:      user.IsPublic,
		PublicKey:     user.PublicKey,
		HasKeyBackup:  user.EncryptedKeyBackup != nil && *user.EncryptedKeyBackup != "",
		SkillsOffered: skillsOffered,
		SkillsWanted:  skillsWanted,
		CreatedAt:     user.CreatedAt,
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// ── E2EE key management ──────────────────────────────────────────────────────

func (s *userService) SetE2EEKeys(userID uuid.UUID, publicKey, encryptedBackup string) error {
	// Validate public key: must be standard base64 encoding of exactly 32 bytes (X25519)
	decoded, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return fmt.Errorf("invalid public key encoding: %w", apperrors.ErrValidation)
	}
	if len(decoded) != 32 {
		return fmt.Errorf("invalid public key length (expected 32 bytes, got %d): %w", len(decoded), apperrors.ErrValidation)
	}

	// Validate encrypted backup: non-empty, reasonable size
	if len(encryptedBackup) == 0 {
		return fmt.Errorf("encrypted key backup is required: %w", apperrors.ErrValidation)
	}
	if len(encryptedBackup) > 512 {
		return fmt.Errorf("encrypted key backup too large: %w", apperrors.ErrValidation)
	}

	return s.userRepo.UpdateE2EEKeys(userID, publicKey, encryptedBackup)
}

func (s *userService) GetPublicKey(userID uuid.UUID) (string, error) {
	return s.userRepo.GetPublicKey(userID)
}

func (s *userService) GetKeyBackup(userID uuid.UUID) (string, error) {
	return s.userRepo.GetKeyBackup(userID)
}
