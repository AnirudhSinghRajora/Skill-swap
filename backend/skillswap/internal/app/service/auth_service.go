package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
	"unicode"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/repository"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2api "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

type AuthService interface {
	Register(req *RegisterRequest) (*AuthResponse, error)
	Login(req *LoginRequest) (*AuthResponse, error)
	RefreshToken(refreshToken string) (*AuthResponse, error)
	ValidateToken(tokenString string) (*TokenClaims, error)
	GetGoogleAuthURL(state string) string
	GoogleCallback(code string) (*AuthResponse, error)
	ForgotPassword(email string) (string, error)
	ResetPassword(token, newPassword string) error
	SendVerificationEmail(userID uuid.UUID) (string, error)
	VerifyEmail(token string) error
	GetMeInfo(userID uuid.UUID) (*MeInfo, error)
}

// MeInfo is the lightweight self-profile used by GET /auth/me. The boolean
// flags let the client decide whether E2EE setup is required without
// fetching any sensitive material.
type MeInfo struct {
	UserID        uuid.UUID `json:"user_id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	HasPublicKey  bool      `json:"has_public_key"`
	HasKeyBackup  bool      `json:"has_key_backup"`
	EmailVerified bool      `json:"email_verified"`
}

type authService struct {
	userRepo    repository.UserRepository
	cfg         config.Config
	db          *gorm.DB
	oauthConfig *oauth2.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg config.Config) AuthService {
	svc := &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		svc.oauthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	return svc
}

// NewAuthServiceWithDB creates an auth service with direct DB access (for token tables)
func NewAuthServiceWithDB(userRepo repository.UserRepository, cfg config.Config, db *gorm.DB) AuthService {
	svc := &authService{
		userRepo: userRepo,
		cfg:      cfg,
		db:       db,
	}

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		svc.oauthConfig = &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	return svc
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
	Name      string    `json:"name,omitempty"`
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

	// OAuth users can't login with password
	if user.PasswordHash == "" {
		return nil, fmt.Errorf("please use %s to sign in: %w", user.AuthProvider, apperrors.ErrInvalidCredentials)
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

// GetMeInfo returns the lightweight self-profile used by GET /auth/me.
// The HasPublicKey / HasKeyBackup flags drive client-side E2EE setup
// gating without leaking any sensitive material.
func (s *authService) GetMeInfo(userID uuid.UUID) (*MeInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
	}
	return &MeInfo{
		UserID:        user.UserID,
		Name:          user.Name,
		Email:         user.Email,
		HasPublicKey:  user.PublicKey != nil && *user.PublicKey != "",
		HasKeyBackup:  user.EncryptedKeyBackup != nil && *user.EncryptedKeyBackup != "",
		EmailVerified: user.EmailVerified,
	}, nil
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
		Name:      user.Name,
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
		Name:      user.Name,
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

// GetGoogleAuthURL returns the Google OAuth consent URL
func (s *authService) GetGoogleAuthURL(state string) string {
	if s.oauthConfig == nil {
		return ""
	}
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// GoogleCallback exchanges the auth code for tokens and creates/logs in a user
func (s *authService) GoogleCallback(code string) (*AuthResponse, error) {
	if s.oauthConfig == nil {
		return nil, fmt.Errorf("google oauth not configured: %w", apperrors.ErrValidation)
	}

	ctx := context.Background()

	// Exchange code for token
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", apperrors.ErrInvalidCredentials)
	}

	// Get user info from Google
	oauth2Service, err := oauth2api.NewService(ctx, option.WithTokenSource(s.oauthConfig.TokenSource(ctx, token)))
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 service: %w", err)
	}

	googleUser, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if googleUser.Email == "" {
		return nil, fmt.Errorf("no email returned from Google: %w", apperrors.ErrValidation)
	}

	// Find or create user
	user, err := s.userRepo.GetByEmail(googleUser.Email)
	if err != nil {
		// User doesn't exist — create new account
		user = &models.User{
			Name:          googleUser.Name,
			Email:         googleUser.Email,
			AuthProvider:  "google",
			GoogleID:      &googleUser.Id,
			EmailVerified: true,
			IsPublic:      true,
		}
		if err := s.userRepo.Create(user); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// User exists — link Google account if not already linked
		if user.GoogleID == nil || *user.GoogleID == "" {
			user.GoogleID = &googleUser.Id
			user.AuthProvider = "google"
			user.EmailVerified = true
			if err := s.userRepo.Update(user); err != nil {
				return nil, fmt.Errorf("failed to link google account: %w", err)
			}
		}
	}

	return s.generateAuthResponse(user)
}

// ForgotPassword creates a password reset token and returns it
func (s *authService) ForgotPassword(email string) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not configured: %w", apperrors.ErrValidation)
	}

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Don't reveal if the user exists — return success silently
		return "", nil
	}

	// Generate secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	resetToken := &models.PasswordResetToken{
		UserID:    user.UserID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.db.Create(resetToken).Error; err != nil {
		return "", fmt.Errorf("failed to save reset token: %w", err)
	}

	return tokenStr, nil
}

// ResetPassword validates a reset token and sets a new password
func (s *authService) ResetPassword(token, newPassword string) error {
	if s.db == nil {
		return fmt.Errorf("database not configured: %w", apperrors.ErrValidation)
	}

	if err := validatePassword(newPassword); err != nil {
		return err
	}

	var resetToken models.PasswordResetToken
	if err := s.db.Where("token = ? AND used = false AND expires_at > ?", token, time.Now()).First(&resetToken).Error; err != nil {
		return fmt.Errorf("invalid or expired reset token: %w", apperrors.ErrInvalidToken)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user, err := s.userRepo.GetByID(resetToken.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Clear the E2EE key backup. It was encrypted with the OLD password
	// (PBKDF2-derived AES-GCM) and is now undecryptable. Leaving it in
	// place would cause the client to silently overwrite it on next
	// login with a fresh key pair, but only after a confusing
	// "decryption failed" fallback. Clearing it makes the next login
	// deterministically generate fresh E2EE keys bound to the new
	// password. (Old ciphertext in past conversations is unrecoverable
	// — that's a fundamental property of password-derived key escrow.)
	if err := s.db.Model(&models.User{}).
		Where("user_id = ?", user.UserID).
		Updates(map[string]any{
			"public_key":           nil,
			"encrypted_key_backup": nil,
			"updated_at":           time.Now(),
		}).Error; err != nil {
		// Non-fatal: log via returned error path would mask the password
		// reset success. The client's E2EE init will still work — it
		// will just attempt decryption first, fail, then regenerate.
		_ = err
	}

	// Mark token as used
	resetToken.Used = true
	s.db.Save(&resetToken)

	return nil
}

// SendVerificationEmail creates a verification token and returns it
func (s *authService) SendVerificationEmail(userID uuid.UUID) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("database not configured: %w", apperrors.ErrValidation)
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
	}

	if user.EmailVerified {
		return "", fmt.Errorf("email already verified: %w", apperrors.ErrValidation)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	verifyToken := &models.EmailVerificationToken{
		UserID:    user.UserID,
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.db.Create(verifyToken).Error; err != nil {
		return "", fmt.Errorf("failed to save verification token: %w", err)
	}

	return tokenStr, nil
}

// VerifyEmail validates a verification token and marks the email as verified
func (s *authService) VerifyEmail(token string) error {
	if s.db == nil {
		return fmt.Errorf("database not configured: %w", apperrors.ErrValidation)
	}

	var verifyToken models.EmailVerificationToken
	if err := s.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&verifyToken).Error; err != nil {
		return fmt.Errorf("invalid or expired verification token: %w", apperrors.ErrInvalidToken)
	}

	user, err := s.userRepo.GetByID(verifyToken.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
	}

	user.EmailVerified = true
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Delete the used token
	s.db.Delete(&verifyToken)

	return nil
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
