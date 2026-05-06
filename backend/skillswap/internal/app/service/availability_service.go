package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AvailabilityService interface {
	// Availability CRUD operations
	CreateAvailabilitySlot(req *CreateAvailabilitySlotDTO) (*models.AvailabilitySlot, error)
	GetUserAvailabilitySlots(userID uuid.UUID) ([]models.AvailabilitySlot, error)
	GetAvailabilitySlot(slotID uuid.UUID, userID uuid.UUID) (*models.AvailabilitySlot, error)
	UpdateAvailabilitySlot(slotID uuid.UUID, userID uuid.UUID, req *UpdateAvailabilitySlotDTO) (*models.AvailabilitySlot, error)
	DeleteAvailabilitySlot(slotID uuid.UUID, userID uuid.UUID) error

	// Availability queries
	FindCommonAvailability(userID1, userID2 uuid.UUID) ([]CommonAvailabilitySlot, error)
	GetAvailabilityByDayAndTime(userID uuid.UUID, dayOfWeek int, startTime, endTime time.Time) ([]models.AvailabilitySlot, error)
}

// DTOs and Request structures
type CreateAvailabilitySlotDTO struct {
	UserID     uuid.UUID `json:"user_id"` // Will be set from JWT token
	Label      string    `json:"label" binding:"required,min=1,max=100"`
	DayBitmask int32     `json:"day_bitmask" binding:"required,min=1,max=127"` // 1-127 (binary representation of days)
	StartTime  string    `json:"start_time" binding:"required"`                // Format: "15:04"
	EndTime    string    `json:"end_time" binding:"required"`                  // Format: "15:04"
}

type UpdateAvailabilitySlotDTO struct {
	Label      string `json:"label" binding:"required,min=1,max=100"`
	DayBitmask int32  `json:"day_bitmask" binding:"required,min=1,max=127"`
	StartTime  string `json:"start_time" binding:"required"`
	EndTime    string `json:"end_time" binding:"required"`
}

type CommonAvailabilitySlot struct {
	Day       string `json:"day"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Duration  int    `json:"duration_minutes"`
}

type availabilityService struct {
	db *gorm.DB
}

func NewAvailabilityService(db *gorm.DB) AvailabilityService {
	return &availabilityService{db: db}
}

// CreateAvailabilitySlot creates a new availability slot for a user
func (a *availabilityService) CreateAvailabilitySlot(req *CreateAvailabilitySlotDTO) (*models.AvailabilitySlot, error) {
	// Parse time strings
	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format, use HH:MM: %w", apperrors.ErrValidation)
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format, use HH:MM: %w", apperrors.ErrValidation)
	}

	// Validate time range
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return nil, fmt.Errorf("end_time must be after start_time: %w", apperrors.ErrValidation)
	}

	// Validate day bitmask (1-127, representing Monday=1, Tuesday=2, ..., Sunday=64)
	if req.DayBitmask < 1 || req.DayBitmask > 127 {
		return nil, fmt.Errorf("day_bitmask must be between 1 and 127: %w", apperrors.ErrValidation)
	}

	slot := &models.AvailabilitySlot{
		UserID:     req.UserID,
		Label:      req.Label,
		DayBitmask: req.DayBitmask,
		StartTime:  startTime,
		EndTime:    endTime,
	}

	err = a.db.Create(slot).Error
	if err != nil {
		return nil, err
	}

	// Load user relation
	err = a.db.Preload("User").First(slot, slot.SlotID).Error
	return slot, err
}

// GetUserAvailabilitySlots retrieves all availability slots for a user
func (a *availabilityService) GetUserAvailabilitySlots(userID uuid.UUID) ([]models.AvailabilitySlot, error) {
	var slots []models.AvailabilitySlot
	err := a.db.Where("user_id = ?", userID).
		Order("day_bitmask ASC, start_time ASC").
		Find(&slots).Error

	return slots, err
}

// GetAvailabilitySlot retrieves a specific availability slot
func (a *availabilityService) GetAvailabilitySlot(slotID uuid.UUID, userID uuid.UUID) (*models.AvailabilitySlot, error) {
	var slot models.AvailabilitySlot
	err := a.db.Where("slot_id = ? AND user_id = ?", slotID, userID).
		First(&slot).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("availability slot not found: %w", apperrors.ErrNotFound)
		}
		return nil, err
	}

	return &slot, nil
}

// UpdateAvailabilitySlot updates an existing availability slot
func (a *availabilityService) UpdateAvailabilitySlot(slotID uuid.UUID, userID uuid.UUID, req *UpdateAvailabilitySlotDTO) (*models.AvailabilitySlot, error) {
	slot, err := a.GetAvailabilitySlot(slotID, userID)
	if err != nil {
		return nil, err
	}

	// Parse time strings
	startTime, err := time.Parse("15:04", req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format, use HH:MM: %w", apperrors.ErrValidation)
	}

	endTime, err := time.Parse("15:04", req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format, use HH:MM: %w", apperrors.ErrValidation)
	}

	// Validate time range
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		return nil, fmt.Errorf("end_time must be after start_time: %w", apperrors.ErrValidation)
	}

	// Validate day bitmask
	if req.DayBitmask < 1 || req.DayBitmask > 127 {
		return nil, fmt.Errorf("day_bitmask must be between 1 and 127: %w", apperrors.ErrValidation)
	}

	// Update fields
	slot.Label = req.Label
	slot.DayBitmask = req.DayBitmask
	slot.StartTime = startTime
	slot.EndTime = endTime

	err = a.db.Save(slot).Error
	if err != nil {
		return nil, err
	}

	return slot, nil
}

// DeleteAvailabilitySlot deletes an availability slot
func (a *availabilityService) DeleteAvailabilitySlot(slotID uuid.UUID, userID uuid.UUID) error {
	result := a.db.Where("slot_id = ? AND user_id = ?", slotID, userID).
		Delete(&models.AvailabilitySlot{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("availability slot not found: %w", apperrors.ErrNotFound)
	}

	return nil
}

// FindCommonAvailability finds overlapping availability between two users
func (a *availabilityService) FindCommonAvailability(userID1, userID2 uuid.UUID) ([]CommonAvailabilitySlot, error) {
	var user1Slots []models.AvailabilitySlot
	var user2Slots []models.AvailabilitySlot

	// Get availability slots for both users
	err := a.db.Where("user_id = ?", userID1).Find(&user1Slots).Error
	if err != nil {
		return nil, err
	}

	err = a.db.Where("user_id = ?", userID2).Find(&user2Slots).Error
	if err != nil {
		return nil, err
	}

	var commonSlots []CommonAvailabilitySlot

	// Find overlapping slots
	for _, slot1 := range user1Slots {
		for _, slot2 := range user2Slots {
			// Check if days overlap using bitwise AND
			if slot1.DayBitmask&slot2.DayBitmask != 0 {
				// Find time overlap
				start := slot1.StartTime
				if slot2.StartTime.After(start) {
					start = slot2.StartTime
				}

				end := slot1.EndTime
				if slot2.EndTime.Before(end) {
					end = slot2.EndTime
				}

				// If there's a valid overlap
				if start.Before(end) {
					// Convert overlapping days to string representation
					days := getDaysFromBitmask(slot1.DayBitmask & slot2.DayBitmask)

					for _, day := range days {
						duration := int(end.Sub(start).Minutes())
						commonSlots = append(commonSlots, CommonAvailabilitySlot{
							Day:       day,
							StartTime: start.Format("15:04"),
							EndTime:   end.Format("15:04"),
							Duration:  duration,
						})
					}
				}
			}
		}
	}

	return commonSlots, nil
}

// GetAvailabilityByDayAndTime finds availability slots for specific day and time range
func (a *availabilityService) GetAvailabilityByDayAndTime(userID uuid.UUID, dayOfWeek int, startTime, endTime time.Time) ([]models.AvailabilitySlot, error) {
	dayBitmask := 1 << (dayOfWeek - 1) // Convert day (1-7) to bitmask

	var slots []models.AvailabilitySlot
	err := a.db.Where("user_id = ? AND (day_bitmask & ?) > 0 AND start_time <= ? AND end_time >= ?",
		userID, dayBitmask, endTime.Format("15:04"), startTime.Format("15:04")).
		Find(&slots).Error

	return slots, err
}

// Helper function to convert bitmask to day names
func getDaysFromBitmask(bitmask int32) []string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var result []string

	for i, day := range days {
		if bitmask&(1<<i) != 0 {
			result = append(result, day)
		}
	}

	return result
}
