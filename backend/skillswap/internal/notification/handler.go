package notification

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	notificationService *service.NotificationService
}

func NewHandler(notificationService *service.NotificationService) *Handler {
	return &Handler{
		notificationService: notificationService,
	}
}

// GetUserNotifications retrieves notifications for the authenticated user
// @Summary Get user notifications
// @Description Get notifications for the authenticated user with pagination
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param unread_only query bool false "Get only unread notifications"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications [get]
func (h *Handler) GetUserNotifications(c *gin.Context) {
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

	// Parse query parameters
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	unreadOnly := c.Query("unread_only") == "true"

	notifications, total, err := h.notificationService.GetUserNotifications(uid, page, limit, unreadOnly)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Convert to response format
	response := make([]models.NotificationResponse, len(notifications))
	for i, notification := range notifications {
		response[i] = models.NotificationResponse{
			NotificationID: notification.NotificationID,
			Type:           notification.Type,
			Title:          notification.Title,
			Message:        notification.Message,
			IsRead:         notification.IsRead,
			RelatedID:      notification.RelatedID,
			CreatedAt:      notification.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": response,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// MarkNotificationsAsRead marks specified notifications as read
// @Summary Mark notifications as read
// @Description Mark specified notifications as read for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body models.MarkAsReadRequest true "Notification IDs to mark as read"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/mark-read [put]
func (h *Handler) MarkNotificationsAsRead(c *gin.Context) {
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

	var req models.MarkAsReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if len(req.NotificationIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No notification IDs provided"})
		return
	}

	if err := h.notificationService.MarkNotificationsAsRead(uid, req.NotificationIDs); err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications marked as read"})
}

// MarkAllAsRead marks all notifications as read for the authenticated user
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/mark-all-read [put]
func (h *Handler) MarkAllAsRead(c *gin.Context) {
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

	if err := h.notificationService.MarkAllAsRead(uid); err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification deletes a specific notification
// @Summary Delete notification
// @Description Delete a specific notification for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/{id} [delete]
func (h *Handler) DeleteNotification(c *gin.Context) {
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

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := h.notificationService.DeleteNotification(uid, notificationID); err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// GetNotificationStats returns notification statistics for the authenticated user
// @Summary Get notification statistics
// @Description Get notification statistics (total, read, unread) for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.NotificationStatsResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/stats [get]
func (h *Handler) GetNotificationStats(c *gin.Context) {
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

	stats, err := h.notificationService.GetNotificationStats(uid)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateNotification creates a new notification (admin only)
// @Summary Create notification
// @Description Create a new notification (admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body models.NotificationRequest true "Notification details"
// @Security BearerAuth
// @Success 201 {object} models.NotificationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications [post]
func (h *Handler) CreateNotification(c *gin.Context) {
	// Check if user is admin
	isAdmin, exists := c.Get("is_admin")
	isAdminBool, ok := isAdmin.(bool)
	if !exists || !ok || !isAdminBool {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Validate notification type
	switch req.Type {
	case models.NotificationTypeSwapRequest, models.NotificationTypeSwapAccepted,
		models.NotificationTypeSwapRejected, models.NotificationTypeSwapCompleted,
		models.NotificationTypeNewRating, models.NotificationTypeSkillMatched,
		models.NotificationTypeSystemAlert, models.NotificationTypeAdminNotice:
		// valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification type"})
		return
	}

	notification, err := h.notificationService.CreateNotification(&req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response := models.NotificationResponse{
		NotificationID: notification.NotificationID,
		Type:           notification.Type,
		Title:          notification.Title,
		Message:        notification.Message,
		IsRead:         notification.IsRead,
		RelatedID:      notification.RelatedID,
		CreatedAt:      notification.CreatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetNotificationByID retrieves a specific notification
// @Summary Get notification by ID
// @Description Get a specific notification by its ID for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Security BearerAuth
// @Success 200 {object} models.NotificationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notifications/{id} [get]
func (h *Handler) GetNotificationByID(c *gin.Context) {
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

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	notification, err := h.notificationService.GetNotificationByID(uid, notificationID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		response.InternalError(c, err)
		return
	}

	response := models.NotificationResponse{
		NotificationID: notification.NotificationID,
		Type:           notification.Type,
		Title:          notification.Title,
		Message:        notification.Message,
		IsRead:         notification.IsRead,
		RelatedID:      notification.RelatedID,
		CreatedAt:      notification.CreatedAt,
	}

	c.JSON(http.StatusOK, response)
}
