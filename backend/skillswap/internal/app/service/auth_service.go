package service

import (
	"fmt"
	"time"
	"unicode"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/repository"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	RefreshToken(refreshToken string) (*AuthResponse, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// DTOs for authentication
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Location string `json:"location,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Location     *string   `json:"location"`
	HasPhoto     bool      `json:"has_photo"`
	IsPublic     bool      `json:"is_public"`
	PublicKey    *string   `json:"public_key,omitempty"`
	HasKeyBackup bool      `json:"has_key_backup"`
}

type TokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	IsAdmin   bool      `json:"is_admin"`
	TokenType string    `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// Register creates a new user account
func (s *authService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, _ := s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists: %w", apperrors.ErrEmailTaken)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		IsPublic:     true, // Default to public profile
	}

	if req.Location != "" {
		user.Location = &req.Location
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// Login authenticates a user
func (s *authService) Login(req *LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password: %w", apperrors.ErrInvalidCredentials)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password: %w", apperrors.ErrInvalidCredentials)
	}

	// Generate tokens
	return s.generateAuthResponse(user)
}

// RefreshToken generates new access token from refresh token
func (s *authService) RefreshToken(refreshToken string) (*AuthResponse, error) {
	// Parse and validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", apperrors.ErrInvalidToken)
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type: %w", apperrors.ErrInvalidToken)
	}

	// Get user to generate new tokens
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
	}

	return s.generateAuthResponse(user)
}

// ValidateToken validates and parses JWT token
func (s *authService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %w", apperrors.ErrInvalidToken)
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token: %w", apperrors.ErrInvalidToken)
}

// Helper function to generate auth response with tokens
func (s *authService) generateAuthResponse(user *models.User) (*AuthResponse, error) {
	accessTokenExp := time.Now().Add(15 * time.Minute)
	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour)

	// Generate access token
	accessClaims := TokenClaims{
		UserID:    user.UserID,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin, // Use the user's actual admin status
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessTokenExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.UserID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := TokenClaims{
		UserID:    user.UserID,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin, // Use the user's actual admin status
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshTokenExp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.UserID.String(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(accessTokenExp).Seconds()),
		User: UserInfo{
			UserID:       user.UserID,
			Name:         user.Name,
			Email:        user.Email,
			Location:     user.Location,
			HasPhoto:     len(user.PhotoData) > 0,
			IsPublic:     user.IsPublic,
			PublicKey:    user.PublicKey,
			HasKeyBackup: user.EncryptedKeyBackup != nil && *user.EncryptedKeyBackup != "",
		},
	}, nil
}

// validatePassword checks password meets minimum strength requirements
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters: %w", apperrors.ErrValidation)
	}
	var hasLetter, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsDigit(ch):
			hasDigit = true
		default:
			hasSpecial = true
		}
	}
	if !hasLetter {
		return fmt.Errorf("password must contain at least one letter: %w", apperrors.ErrValidation)
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one number: %w", apperrors.ErrValidation)
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character: %w", apperrors.ErrValidation)
	}
	return nil
}
