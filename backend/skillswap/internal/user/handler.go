package user

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	userService appservice.UserService
}

func NewHandler(userService appservice.UserService) *Handler {
	return &Handler{
		userService: userService,
	}
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} appservice.UserProfileResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			response.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body appservice.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req appservice.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate field lengths
	if req.Name != nil {
		name := *req.Name
		if len(name) < 2 || len(name) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be between 2 and 100 characters"})
			return
		}
	}
	if req.Location != nil && len(*req.Location) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Location must be at most 200 characters"})
		return
	}

	if err := h.userService.UpdateProfile(userID, &req); err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			response.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// SearchUsers godoc
// @Summary Search users
// @Description Search for users by location or search term
// @Tags users
// @Accept json
// @Produce json
// @Param location query string false "Filter by location"
// @Param search_term query string false "Search in user names"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} appservice.SearchUsersResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/search [get]
func (h *Handler) SearchUsers(c *gin.Context) {
	req := appservice.SearchUsersRequest{
		Location:   c.Query("location"),
		SearchTerm: c.Query("search_term"),
		Page:       1,
		Limit:      10,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			if limit > 100 {
				limit = 100
			}
			req.Limit = limit
		}
	}

	result, err := h.userService.SearchUsers(&req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// Common response types
type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

// GetPublicProfile godoc
// @Summary Get a user's public profile
// @Description Get a user's profile by ID (only if their profile is public)
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} appservice.UserProfileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /public/users/{id} [get]
func (h *Handler) GetPublicProfile(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	profile, err := h.userService.GetPublicProfile(userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		}
		return
	}

	c.JSON(http.StatusOK, profile)
}

// ── E2EE Key Management ─────────────────────────────────────────────────────

type setE2EEKeysRequest struct {
	PublicKey          string `json:"public_key" binding:"required"`
	EncryptedKeyBackup string `json:"encrypted_key_backup" binding:"required"`
}

// SetE2EEKeys godoc
// @Summary Upload E2EE keys
// @Description Upload the user's E2EE public key and encrypted private key backup
// @Tags e2ee
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body setE2EEKeysRequest true "E2EE key data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me/e2ee-keys [put]
func (h *Handler) SetE2EEKeys(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req setE2EEKeysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "public_key and encrypted_key_backup are required"})
		return
	}

	if err := h.userService.SetE2EEKeys(userID, req.PublicKey, req.EncryptedKeyBackup); err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			response.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "E2EE keys updated successfully"})
}

// GetUserPublicKey godoc
// @Summary Get a user's E2EE public key
// @Description Fetch any user's E2EE public key for deriving a shared secret
// @Tags e2ee
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]string "public_key"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id}/public-key [get]
func (h *Handler) GetUserPublicKey(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	publicKey, err := h.userService.GetPublicKey(userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Public key not found"})
		} else {
			response.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"public_key": publicKey})
}

// GetKeyBackup godoc
// @Summary Get E2EE key backup
// @Description Fetch the authenticated user's encrypted E2EE private key backup
// @Tags e2ee
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "encrypted_key_backup"
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/me/key-backup [get]
func (h *Handler) GetKeyBackup(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	backup, err := h.userService.GetKeyBackup(userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Key backup not found"})
		} else {
			response.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"encrypted_key_backup": backup})
}
