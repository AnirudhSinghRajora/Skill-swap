package file

import (
	"errors"
	"net/http"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	fileUploadService *service.FileUploadService
}

func NewHandler(fileUploadService *service.FileUploadService) *Handler {
	return &Handler{
		fileUploadService: fileUploadService,
	}
}

// UploadUserPhoto uploads a user's profile photo
// @Summary Upload user profile photo
// @Description Upload a profile photo for the authenticated user
// @Tags files
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Profile photo file (max 5MB, jpg/png/gif/webp)"
// @Security BearerAuth
// @Success 200 {object} models.FileUploadResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 413 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/users/photo [post]
func (h *Handler) UploadUserPhoto(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided or invalid file"})
		return
	}

	// Upload file
	response, err := h.fileUploadService.UploadUserPhoto(uid, file)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrFileTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		case errors.Is(err, apperrors.ErrInvalidFileType):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

// DeleteUserPhoto deletes the authenticated user's profile photo
// @Summary Delete user profile photo
// @Description Delete the profile photo for the authenticated user
// @Tags files
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.FileDeleteResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/users/photo [delete]
func (h *Handler) DeleteUserPhoto(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.fileUploadService.DeleteUserPhoto(uid); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound), errors.Is(err, apperrors.ErrNoPhoto):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, models.FileDeleteResponse{
		Message: "Photo deleted successfully",
		Success: true,
	})
}

// GetUserPhoto serves a user's profile photo from database
// @Summary Get user profile photo
// @Description Serve a user's profile photo stored in database
// @Tags files
// @Produce image/jpeg,image/png,image/gif,image/webp
// @Param user_id path string true "User ID"
// @Success 200 {file} image
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/users/{user_id}/photo [get]
func (h *Handler) GetUserPhoto(c *gin.Context) {
	userIDStr := c.Param("user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get photo data from database
	photoData, mimeType, err := h.fileUploadService.GetUserPhoto(userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, apperrors.ErrNoPhoto) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found"})
			return
		}
		resp.InternalError(c, err)
		return
	}

	// Set appropriate headers
	c.Header("Content-Type", mimeType)
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Serve the image data
	c.Data(http.StatusOK, mimeType, photoData)
}

// GetUserPhotoInfo gets information about a user's own profile photo
// @Summary Get user photo info
// @Description Get information about the authenticated user's profile photo. Only the photo owner can access this.
// @Tags files
// @Produce json
// @Param user_id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} models.FileInfo
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/files/users/{user_id}/info [get]
func (h *Handler) GetUserPhotoInfo(c *gin.Context) {
	// Only the photo owner can access their photo info
	authUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr := c.Param("user_id")
	if userIDStr != authUserID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	fileInfo, err := h.fileUploadService.GetFileInfo(userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, apperrors.ErrNoPhoto) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found"})
			return
		}
		resp.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, fileInfo)
}
