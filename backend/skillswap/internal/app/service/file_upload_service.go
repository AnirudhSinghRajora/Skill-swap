package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileUploadService struct {
	db           *gorm.DB
	maxFileSize  int64
	allowedTypes map[string]bool
}

func NewFileUploadService(db *gorm.DB) *FileUploadService {
	// Set allowed image types
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	return &FileUploadService{
		db:           db,
		maxFileSize:  5 * 1024 * 1024, // 5MB default
		allowedTypes: allowedTypes,
	}
}

// UploadUserPhoto uploads a user's profile photo to database
func (s *FileUploadService) UploadUserPhoto(userID uuid.UUID, file *multipart.FileHeader) (*models.FileUploadResponse, error) {
	// Validate file
	if err := s.validateFile(file); err != nil {
		return nil, err
	}

	// Read file data
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	photoData, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	// Get MIME type
	mimeType := file.Header.Get("Content-Type")

	// Update user's photo in database
	result := s.db.Model(&models.User{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"photo_data":      photoData,
			"photo_mime_type": mimeType,
		})

	if result.Error != nil {
		return nil, fmt.Errorf("failed to save photo to database: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
	}

	return &models.FileUploadResponse{
		Filename: file.Filename,
		URL:      fmt.Sprintf("/api/v1/files/users/%s/photo", userID.String()),
		Size:     file.Size,
		MimeType: mimeType,
	}, nil
}

// DeleteUserPhoto deletes a user's profile photo from database
func (s *FileUploadService) DeleteUserPhoto(userID uuid.UUID) error {
	// Check if user exists and has photo
	var user models.User
	if err := s.db.Select("photo_data").First(&user, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user has photo
	if len(user.PhotoData) == 0 {
		return fmt.Errorf("user has no photo to delete: %w", apperrors.ErrNoPhoto)
	}

	// Clear photo data in database
	result := s.db.Model(&models.User{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"photo_data":      nil,
			"photo_mime_type": nil,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to delete photo: %w", result.Error)
	}

	return nil
}

// GetUserPhoto returns user's photo data and MIME type
func (s *FileUploadService) GetUserPhoto(userID uuid.UUID) ([]byte, string, error) {
	var user models.User
	if err := s.db.Select("photo_data, photo_mime_type").First(&user, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, "", fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
		}
		return nil, "", fmt.Errorf("failed to get user photo: %w", err)
	}

	if len(user.PhotoData) == 0 {
		return nil, "", fmt.Errorf("user has no photo: %w", apperrors.ErrNoPhoto)
	}

	mimeType := "image/jpeg" // default
	if user.PhotoMimeType != nil {
		mimeType = *user.PhotoMimeType
	}

	return user.PhotoData, mimeType, nil
}

// GetFileInfo returns information about a user's photo
func (s *FileUploadService) GetFileInfo(userID uuid.UUID) (*models.FileInfo, error) {
	var user models.User
	if err := s.db.Select("photo_data, photo_mime_type, updated_at").First(&user, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if len(user.PhotoData) == 0 {
		return nil, fmt.Errorf("user has no photo: %w", apperrors.ErrNoPhoto)
	}

	return &models.FileInfo{
		Filename: "profile_photo",
		Size:     int64(len(user.PhotoData)),
		ModTime:  user.UpdatedAt,
		URL:      fmt.Sprintf("/api/v1/files/users/%s/photo", userID.String()),
	}, nil
}

// validateFile validates the uploaded file
func (s *FileUploadService) validateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > s.maxFileSize {
		return fmt.Errorf("file too large: %d bytes (max: %d bytes): %w", file.Size, s.maxFileSize, apperrors.ErrFileTooLarge)
	}

	// Check file type
	contentType := file.Header.Get("Content-Type")
	if !s.allowedTypes[contentType] {
		return fmt.Errorf("invalid file type: %s: %w", contentType, apperrors.ErrInvalidFileType)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !validExts[ext] {
		return fmt.Errorf("invalid file extension: %s: %w", ext, apperrors.ErrInvalidFileType)
	}

	return nil
}
