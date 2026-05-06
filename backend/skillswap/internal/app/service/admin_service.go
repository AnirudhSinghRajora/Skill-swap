package service

import (
	"errors"
	"fmt"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdminService interface {
	// User management
	GetAllUsers(filter AdminUserFilter) ([]models.User, int64, error)
	BanUser(adminID, userID uuid.UUID) error
	UnbanUser(adminID, userID uuid.UUID) error
	DeleteUser(adminID, userID uuid.UUID) error
	MakeUserAdmin(adminID, userID uuid.UUID) error
	RemoveUserAdmin(adminID, userID uuid.UUID) error

	// Swap management
	GetAllSwaps(filter AdminSwapFilter) ([]models.SwapRequest, int64, error)
	CancelSwap(adminID, swapID uuid.UUID, reason string) error

	// Platform statistics
	GetPlatformStats() (*PlatformStats, error)

	// Content moderation
	GetReportedContent() ([]ReportedContent, error)
}

// DTOs and filters
type AdminUserFilter struct {
	Search    string `json:"search,omitempty"`
	IsBanned  *bool  `json:"is_banned,omitempty"`
	IsAdmin   *bool  `json:"is_admin,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`    // "created_at", "name", "email"
	SortOrder string `json:"sort_order,omitempty"` // "asc", "desc"
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

type AdminSwapFilter struct {
	Status      *string    `json:"status,omitempty"`
	RequesterID *uuid.UUID `json:"requester_id,omitempty"`
	ResponderID *uuid.UUID `json:"responder_id,omitempty"`
	SortBy      string     `json:"sort_by,omitempty"`    // "created_at", "updated_at"
	SortOrder   string     `json:"sort_order,omitempty"` // "asc", "desc"
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

type PlatformStats struct {
	TotalUsers        int64   `json:"total_users"`
	ActiveUsers       int64   `json:"active_users"` // Users with activity in last 30 days
	TotalSwaps        int64   `json:"total_swaps"`
	CompletedSwaps    int64   `json:"completed_swaps"`
	PendingSwaps      int64   `json:"pending_swaps"`
	TotalSkills       int64   `json:"total_skills"`
	TotalRatings      int64   `json:"total_ratings"`
	AverageRating     float64 `json:"average_rating"`
	NewUsersThisMonth int64   `json:"new_users_this_month"`
	SwapsThisMonth    int64   `json:"swaps_this_month"`
}

type ReportedContent struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"` // "user", "swap", "rating"
	ContentID   uuid.UUID `json:"content_id"`
	ReporterID  uuid.UUID `json:"reporter_id"`
	Reason      string    `json:"reason"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // "pending", "resolved", "dismissed"
	CreatedAt   string    `json:"created_at"`
}

type adminService struct {
	db *gorm.DB
}

func NewAdminService(db *gorm.DB) AdminService {
	return &adminService{db: db}
}

// GetAllUsers retrieves all users with filtering and pagination
func (a *adminService) GetAllUsers(filter AdminUserFilter) ([]models.User, int64, error) {
	query := a.db.Model(&models.User{})

	// Apply filters
	if filter.Search != "" {
		searchTerm := fmt.Sprintf("%%%s%%", filter.Search)
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}

	if filter.IsBanned != nil {
		query = query.Where("is_banned = ?", *filter.IsBanned)
	}

	if filter.IsAdmin != nil {
		query = query.Where("is_admin = ?", *filter.IsAdmin)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "desc"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(20) // Default limit
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var users []models.User
	err := query.Find(&users).Error
	return users, total, err
}

// BanUser bans a user (only admins can do this)
func (a *adminService) BanUser(adminID, userID uuid.UUID) error {
	// Verify admin permissions
	if err := a.verifyAdminPermissions(adminID); err != nil {
		return err
	}

	// Cannot ban another admin
	var targetUser models.User
	if err := a.db.First(&targetUser, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
		}
		return err
	}

	if targetUser.IsAdmin {
		return fmt.Errorf("cannot ban an admin user: %w", apperrors.ErrForbidden)
	}

	return a.db.Model(&models.User{}).Where("user_id = ?", userID).Update("is_banned", true).Error
}

// UnbanUser unbans a user
func (a *adminService) UnbanUser(adminID, userID uuid.UUID) error {
	if err := a.verifyAdminPermissions(adminID); err != nil {
		return err
	}

	return a.db.Model(&models.User{}).Where("user_id = ?", userID).Update("is_banned", false).Error
}

// DeleteUser soft deletes a user
func (a *adminService) DeleteUser(adminID, userID uuid.UUID) error {
	if err := a.verifyAdminPermissions(adminID); err != nil {
		return err
	}

	// Cannot delete another admin
	var targetUser models.User
	if err := a.db.First(&targetUser, "user_id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found: %w", apperrors.ErrNotFound)
		}
		return err
	}

	if targetUser.IsAdmin {
		return fmt.Errorf("cannot delete an admin user: %w", apperrors.ErrForbidden)
	}

	return a.db.Delete(&models.User{}, "user_id = ?", userID).Error
}

// MakeUserAdmin grants admin privileges to a user
func (a *adminService) MakeUserAdmin(adminID, userID uuid.UUID) error {
	if err := a.verifyAdminPermissions(adminID); err != nil {
		return err
	}

	return a.db.Model(&models.User{}).Where("user_id = ?", userID).Update("is_admin", true).Error
}

// RemoveUserAdmin removes admin privileges from a user
func (a *adminService) RemoveUserAdmin(adminID, userID uuid.UUID) error {
	if err := a.verifyAdminPermissions(adminID); err != nil {
		return err
	}

	// Cannot remove admin from self
	if adminID == userID {
		return fmt.Errorf("cannot remove admin privileges from yourself: %w", apperrors.ErrForbidden)
	}

	return a.db.Model(&models.User{}).Where("user_id = ?", userID).Update("is_admin", false).Error
}

// GetAllSwaps retrieves all swaps with filtering and pagination
func (a *adminService) GetAllSwaps(filter AdminSwapFilter) ([]models.SwapRequest, int64, error) {
	query := a.db.Model(&models.SwapRequest{}).
		Preload("Requester").
		Preload("Responder").
		Preload("OfferedSkill").
		Preload("WantedSkill")

	// Apply filters
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.RequesterID != nil {
		query = query.Where("requester_id = ?", *filter.RequesterID)
	}

	if filter.ResponderID != nil {
		query = query.Where("responder_id = ?", *filter.ResponderID)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	sortOrder := "desc"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(20) // Default limit
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var swaps []models.SwapRequest
	err := query.Find(&swaps).Error
	return swaps, total, err
}

// CancelSwap cancels a swap (admin intervention)
func (a *adminService) CancelSwap(adminID, swapID uuid.UUID, reason string) error {
	if err := a.verifyAdminPermissions(adminID); err != nil {
		return err
	}

	// TODO: Add audit log for admin actions
	return a.db.Model(&models.SwapRequest{}).
		Where("swap_id = ?", swapID).
		Updates(map[string]interface{}{
			"status": models.StatusCancelled,
			// TODO: Add admin_notes field to store the reason
		}).Error
}

// GetPlatformStats retrieves platform-wide statistics
func (a *adminService) GetPlatformStats() (*PlatformStats, error) {
	stats := &PlatformStats{}

	// Total users
	if err := a.db.Model(&models.User{}).Count(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	// Active users (users with swaps in last 30 days)
	if err := a.db.Model(&models.User{}).
		Joins("JOIN swap_requests ON users.user_id = swap_requests.requester_id OR users.user_id = swap_requests.responder_id").
		Where("swap_requests.created_at > NOW() - INTERVAL '30 days'").
		Distinct("users.user_id").
		Count(&stats.ActiveUsers).Error; err != nil {
		return nil, err
	}

	// Total swaps
	if err := a.db.Model(&models.SwapRequest{}).Count(&stats.TotalSwaps).Error; err != nil {
		return nil, err
	}

	// Completed swaps
	if err := a.db.Model(&models.SwapRequest{}).
		Where("status = ?", models.StatusAccepted).
		Count(&stats.CompletedSwaps).Error; err != nil {
		return nil, err
	}

	// Pending swaps
	if err := a.db.Model(&models.SwapRequest{}).
		Where("status = ?", models.StatusPending).
		Count(&stats.PendingSwaps).Error; err != nil {
		return nil, err
	}

	// Total skills
	if err := a.db.Model(&models.Skill{}).Count(&stats.TotalSkills).Error; err != nil {
		return nil, err
	}

	// Total ratings
	if err := a.db.Model(&models.SwapRating{}).Count(&stats.TotalRatings).Error; err != nil {
		return nil, err
	}

	// Average rating
	if stats.TotalRatings > 0 {
		if err := a.db.Model(&models.SwapRating{}).
			Select("AVG(score)").
			Scan(&stats.AverageRating).Error; err != nil {
			return nil, err
		}
	}

	// New users this month
	if err := a.db.Model(&models.User{}).
		Where("created_at > DATE_TRUNC('month', NOW())").
		Count(&stats.NewUsersThisMonth).Error; err != nil {
		return nil, err
	}

	// Swaps this month
	if err := a.db.Model(&models.SwapRequest{}).
		Where("created_at > DATE_TRUNC('month', NOW())").
		Count(&stats.SwapsThisMonth).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

// GetReportedContent retrieves reported content (placeholder implementation)
func (a *adminService) GetReportedContent() ([]ReportedContent, error) {
	// TODO: Implement reporting system
	// For now, return empty slice
	return []ReportedContent{}, nil
}

// verifyAdminPermissions checks if the user has admin privileges
func (a *adminService) verifyAdminPermissions(userID uuid.UUID) error {
	var user models.User
	if err := a.db.First(&user, "user_id = ? AND is_admin = true", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin privileges required: %w", apperrors.ErrForbidden)
		}
		return err
	}

	if user.IsBanned {
		return fmt.Errorf("admin account is banned: %w", apperrors.ErrForbidden)
	}

	return nil
}
