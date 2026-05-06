package service

import (
	"fmt"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationService struct {
	db *gorm.DB
}

func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// CreateNotification creates a new notification
func (s *NotificationService) CreateNotification(req *models.NotificationRequest) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Title:     req.Title,
		Message:   req.Message,
		RelatedID: req.RelatedID,
		IsRead:    false,
	}

	if err := s.db.Create(notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return notification, nil
}

// CreateSwapRequestNotification creates notification for swap request
func (s *NotificationService) CreateSwapRequestNotification(receiverID, requesterID uuid.UUID, swapRequestID uuid.UUID, skillName string) error {
	var requester models.User
	if err := s.db.Select("name").First(&requester, "user_id = ?", requesterID).Error; err != nil {
		return fmt.Errorf("failed to get requester info: %w", err)
	}

	req := &models.NotificationRequest{
		UserID:    receiverID,
		Type:      models.NotificationTypeSwapRequest,
		Title:     "New Swap Request",
		Message:   fmt.Sprintf("%s wants to swap skills for %s", requester.Name, skillName),
		RelatedID: &swapRequestID,
	}

	_, err := s.CreateNotification(req)
	return err
}

// CreateSwapStatusNotification creates notification for swap status changes
func (s *NotificationService) CreateSwapStatusNotification(userID uuid.UUID, swapRequestID uuid.UUID, status string, skillName string) error {
	var title, message string
	var notificationType models.NotificationType

	switch status {
	case "accepted":
		title = "Swap Request Accepted"
		message = fmt.Sprintf("Your swap request for %s has been accepted!", skillName)
		notificationType = models.NotificationTypeSwapAccepted
	case "rejected":
		title = "Swap Request Rejected"
		message = fmt.Sprintf("Your swap request for %s has been rejected.", skillName)
		notificationType = models.NotificationTypeSwapRejected
	case "completed":
		title = "Swap Completed"
		message = fmt.Sprintf("Your skill swap for %s has been marked as completed. Don't forget to rate your partner!", skillName)
		notificationType = models.NotificationTypeSwapCompleted
	default:
		return fmt.Errorf("unknown status: %s", status)
	}

	req := &models.NotificationRequest{
		UserID:    userID,
		Type:      notificationType,
		Title:     title,
		Message:   message,
		RelatedID: &swapRequestID,
	}

	_, err := s.CreateNotification(req)
	return err
}

// CreateRatingNotification creates notification for new rating
func (s *NotificationService) CreateRatingNotification(userID, raterID uuid.UUID, rating int, comment string) error {
	var rater models.User
	if err := s.db.Select("name").First(&rater, "user_id = ?", raterID).Error; err != nil {
		return fmt.Errorf("failed to get rater info: %w", err)
	}

	message := fmt.Sprintf("%s rated you %d stars", rater.Name, rating)
	if comment != "" {
		message += fmt.Sprintf(": \"%s\"", comment)
	}

	req := &models.NotificationRequest{
		UserID:  userID,
		Type:    models.NotificationTypeNewRating,
		Title:   "New Rating Received",
		Message: message,
	}

	_, err := s.CreateNotification(req)
	return err
}

// CreateSkillMatchNotification creates notification for skill matches
func (s *NotificationService) CreateSkillMatchNotification(userID uuid.UUID, matchedUsers []string, skillName string) error {
	message := fmt.Sprintf("Found %d potential matches for your wanted skill: %s", len(matchedUsers), skillName)
	if len(matchedUsers) > 0 {
		message += fmt.Sprintf(". Users: %v", matchedUsers)
	}

	req := &models.NotificationRequest{
		UserID:  userID,
		Type:    models.NotificationTypeSkillMatched,
		Title:   "Skill Match Found",
		Message: message,
	}

	_, err := s.CreateNotification(req)
	return err
}

// CreateSystemNotification creates system-wide notification
func (s *NotificationService) CreateSystemNotification(userIDs []uuid.UUID, title, message string) error {
	notifications := make([]models.Notification, len(userIDs))

	for i, userID := range userIDs {
		notifications[i] = models.Notification{
			UserID:  userID,
			Type:    models.NotificationTypeSystemAlert,
			Title:   title,
			Message: message,
			IsRead:  false,
		}
	}

	if err := s.db.Create(&notifications).Error; err != nil {
		return fmt.Errorf("failed to create system notifications: %w", err)
	}

	return nil
}

// GetUserNotifications retrieves notifications for a user with pagination
func (s *NotificationService) GetUserNotifications(userID uuid.UUID, page, limit int, unreadOnly bool) ([]models.Notification, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var notifications []models.Notification
	var total int64

	query := s.db.Model(&models.Notification{}).Where("user_id = ?", userID)

	if unreadOnly {
		query = query.Where("is_read = ?", false)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count notifications: %w", err)
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, total, nil
}

// MarkNotificationsAsRead marks notifications as read
func (s *NotificationService) MarkNotificationsAsRead(userID uuid.UUID, notificationIDs []uuid.UUID) error {
	result := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND notification_id IN ?", userID, notificationIDs).
		Update("is_read", true)

	if result.Error != nil {
		return fmt.Errorf("failed to mark notifications as read: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no notifications found or already read")
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllAsRead(userID uuid.UUID) error {
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

// DeleteNotification soft deletes a notification
func (s *NotificationService) DeleteNotification(userID, notificationID uuid.UUID) error {
	result := s.db.Where("user_id = ? AND notification_id = ?", userID, notificationID).
		Delete(&models.Notification{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found: %w", apperrors.ErrNotFound)
	}

	return nil
}

// GetNotificationStats returns notification statistics for a user
func (s *NotificationService) GetNotificationStats(userID uuid.UUID) (*models.NotificationStatsResponse, error) {
	var stats models.NotificationStatsResponse

	// Total notifications
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalNotifications).Error; err != nil {
		return nil, fmt.Errorf("failed to count total notifications: %w", err)
	}

	// Unread notifications
	if err := s.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&stats.UnreadCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	stats.ReadCount = stats.TotalNotifications - stats.UnreadCount

	return &stats, nil
}

// CleanupOldNotifications removes notifications older than specified days
func (s *NotificationService) CleanupOldNotifications(daysOld int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)

	result := s.db.Where("created_at < ?", cutoffDate).Delete(&models.Notification{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old notifications: %w", result.Error)
	}

	return nil
}

// GetNotificationByID retrieves a specific notification
func (s *NotificationService) GetNotificationByID(userID, notificationID uuid.UUID) (*models.Notification, error) {
	var notification models.Notification

	if err := s.db.Where("user_id = ? AND notification_id = ?", userID, notificationID).
		First(&notification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found: %w", apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return &notification, nil
}
