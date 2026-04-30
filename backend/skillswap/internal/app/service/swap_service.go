package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/config"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SwapService interface {
	// Swap request CRUD operations
	CreateSwapRequest(req *CreateSwapRequestDTO) (*models.SwapRequest, error)
	GetSwapRequestByID(swapID uuid.UUID) (*models.SwapRequest, error)
	GetUserSwapRequests(userID uuid.UUID, filter SwapRequestFilter) ([]models.SwapRequest, error)
	UpdateSwapStatus(swapID uuid.UUID, userID uuid.UUID, status models.SwapStatus) (*models.SwapRequest, error)
	DeleteSwapRequest(swapID uuid.UUID, userID uuid.UUID) error

	// Swap request queries
	GetSwapRequestsForUser(userID uuid.UUID) (*SwapRequestsResponse, error)
	GetPendingSwapRequests(userID uuid.UUID) ([]models.SwapRequest, error)
	GetSwapHistory(userID uuid.UUID) ([]models.SwapRequest, error)

	// Matching and recommendations
	FindPotentialMatches(userID uuid.UUID) ([]SwapMatch, error)

	// Trust & safety
	ReportNoShow(swapID, reporterID uuid.UUID, reason string) (*models.SwapRequest, error)
	RaiseDispute(swapID, userID uuid.UUID, reason string) (*models.SwapRequest, error)
	GetReliabilityStats(userID uuid.UUID) (ReliabilityStats, error)
}

// ReliabilityStats describes a user's swap reliability. Frontend hides the
// score until the user has at least 3 completed-or-no-show swaps.
type ReliabilityStats struct {
	SuccessfulSwaps  int     `json:"successful_swaps"`
	ReportedNoShows  int     `json:"reported_no_shows"`
	Score            float64 `json:"score"`     // 0..1
	Qualifies        bool    `json:"qualifies"` // denominator >= 3
}

// DTOs and Request structures
type CreateSwapRequestDTO struct {
	RequesterID    uuid.UUID `json:"requester_id"`
	ResponderID    uuid.UUID `json:"responder_id" binding:"required"`
	OfferedSkillID uuid.UUID `json:"offered_skill_id" binding:"required"`
	WantedSkillID  uuid.UUID `json:"wanted_skill_id" binding:"required"`
}

type SwapRequestFilter struct {
	Status   *models.SwapStatus `json:"status,omitempty"`
	Sent     bool               `json:"sent,omitempty"`     // Requests sent by user
	Received bool               `json:"received,omitempty"` // Requests received by user
	Limit    int                `json:"limit,omitempty"`
	Offset   int                `json:"offset,omitempty"`
}

type SwapRequestsResponse struct {
	Sent     []models.SwapRequest `json:"sent"`
	Received []models.SwapRequest `json:"received"`
}

type SwapMatch struct {
	User         models.User  `json:"user"`
	OfferedSkill models.Skill `json:"offered_skill"`
	WantedSkill  models.Skill `json:"wanted_skill"`
	MatchScore   int          `json:"match_score"` // 1-100 compatibility score
}

type swapService struct {
	db                  *gorm.DB
	notificationService *NotificationService
	cfg                 config.Config
}

func NewSwapService(db *gorm.DB, notificationService *NotificationService, cfg config.Config) SwapService {
	return &swapService{db: db, notificationService: notificationService, cfg: cfg}
}

// CreateSwapRequest creates a new swap request
func (s *swapService) CreateSwapRequest(req *CreateSwapRequestDTO) (*models.SwapRequest, error) {
	// Validate that requester and responder are different
	if req.RequesterID == req.ResponderID {
		return nil, fmt.Errorf("cannot create swap request with yourself: %w", apperrors.ErrSelfAction)
	}

	// Refuse if either party has blocked the other (either direction).
	var blockCount int64
	if err := s.db.Model(&models.UserBlock{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
			req.RequesterID, req.ResponderID, req.ResponderID, req.RequesterID).
		Count(&blockCount).Error; err != nil {
		return nil, fmt.Errorf("failed to check block status: %w", err)
	}
	if blockCount > 0 {
		return nil, fmt.Errorf("this user is unavailable: %w", apperrors.ErrForbidden)
	}

	// Optional gate: when REQUIRE_EMAIL_VERIFICATION=true, the requester
	// must have a verified email before they can initiate a swap. We fetch
	// just the verified flag to avoid loading the full user row.
	if s.cfg.RequireEmailVerification {
		var verified bool
		if err := s.db.Model(&models.User{}).
			Select("email_verified").
			Where("user_id = ?", req.RequesterID).
			Scan(&verified).Error; err != nil {
			return nil, fmt.Errorf("failed to check verification status: %w", err)
		}
		if !verified {
			return nil, fmt.Errorf("please verify your email before sending swap requests: %w", apperrors.ErrEmailNotVerified)
		}
	}

	var swapRequest *models.SwapRequest
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Validate that requester offers the offered skill
		var offeredCount int64
		if err := tx.Model(&models.UserSkillOffered{}).
			Where("user_id = ? AND skill_id = ?", req.RequesterID, req.OfferedSkillID).
			Count(&offeredCount).Error; err != nil {
			return err
		}
		if offeredCount == 0 {
			return fmt.Errorf("you don't offer the specified skill: %w", apperrors.ErrValidation)
		}

		// Validate that responder wants the offered skill
		var wantedCount int64
		if err := tx.Model(&models.UserSkillWanted{}).
			Where("user_id = ? AND skill_id = ?", req.ResponderID, req.OfferedSkillID).
			Count(&wantedCount).Error; err != nil {
			return err
		}
		if wantedCount == 0 {
			return fmt.Errorf("responder doesn't want the offered skill: %w", apperrors.ErrValidation)
		}

		// Validate that responder offers the wanted skill
		var responderOffersCount int64
		if err := tx.Model(&models.UserSkillOffered{}).
			Where("user_id = ? AND skill_id = ?", req.ResponderID, req.WantedSkillID).
			Count(&responderOffersCount).Error; err != nil {
			return err
		}
		if responderOffersCount == 0 {
			return fmt.Errorf("responder doesn't offer the requested skill: %w", apperrors.ErrValidation)
		}

		// Check for existing pending request between same users and skills
		var existingCount int64
		if err := tx.Model(&models.SwapRequest{}).
			Where("requester_id = ? AND responder_id = ? AND offered_skill_id = ? AND wanted_skill_id = ? AND status = ?",
				req.RequesterID, req.ResponderID, req.OfferedSkillID, req.WantedSkillID, models.StatusPending).
			Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return fmt.Errorf("pending swap request already exists: %w", apperrors.ErrDuplicate)
		}

		swapRequest = &models.SwapRequest{
			RequesterID:    req.RequesterID,
			ResponderID:    req.ResponderID,
			OfferedSkillID: req.OfferedSkillID,
			WantedSkillID:  req.WantedSkillID,
			Status:         models.StatusPending,
		}

		if err := tx.Create(swapRequest).Error; err != nil {
			return err
		}

		// Load relationships
		return tx.Preload("Requester").Preload("Responder").
			Preload("OfferedSkill").Preload("WantedSkill").
			First(swapRequest, swapRequest.SwapID).Error
	})
	if err != nil {
		return nil, err
	}

	// Notify the responder about the new swap request
	if s.notificationService != nil {
		_ = s.notificationService.CreateSwapRequestNotification(
			swapRequest.ResponderID,
			swapRequest.RequesterID,
			swapRequest.SwapID,
			swapRequest.OfferedSkill.Name,
		)
	}

	return swapRequest, nil
}

// GetSwapRequestByID retrieves a swap request by ID
func (s *swapService) GetSwapRequestByID(swapID uuid.UUID) (*models.SwapRequest, error) {
	var swapRequest models.SwapRequest
	err := s.db.Preload("Requester").Preload("Responder").
		Preload("OfferedSkill").Preload("WantedSkill").
		First(&swapRequest, "swap_id = ?", swapID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("swap request not found: %w", apperrors.ErrNotFound)
		}
		return nil, err
	}

	return &swapRequest, nil
}

// GetUserSwapRequests retrieves swap requests for a user with filtering
func (s *swapService) GetUserSwapRequests(userID uuid.UUID, filter SwapRequestFilter) ([]models.SwapRequest, error) {
	query := s.db.Model(&models.SwapRequest{}).
		Preload("Requester").Preload("Responder").
		Preload("OfferedSkill").Preload("WantedSkill")

	// Apply filters
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.Sent && filter.Received {
		query = query.Where("requester_id = ? OR responder_id = ?", userID, userID)
	} else if filter.Sent {
		query = query.Where("requester_id = ?", userID)
	} else if filter.Received {
		query = query.Where("responder_id = ?", userID)
	} else {
		// Default: show both sent and received
		query = query.Where("requester_id = ? OR responder_id = ?", userID, userID)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// Order by creation date (newest first)
	query = query.Order("created_at DESC")

	var swapRequests []models.SwapRequest
	err := query.Find(&swapRequests).Error

	return swapRequests, err
}

// UpdateSwapStatus updates the status of a swap request
func (s *swapService) UpdateSwapStatus(swapID uuid.UUID, userID uuid.UUID, status models.SwapStatus) (*models.SwapRequest, error) {
	// Get the swap request
	swapRequest, err := s.GetSwapRequestByID(swapID)
	if err != nil {
		return nil, err
	}

	// Only responder can accept/reject requests
	if status == models.StatusAccepted || status == models.StatusRejected {
		if swapRequest.ResponderID != userID {
			return nil, fmt.Errorf("only responder can accept or reject requests: %w", apperrors.ErrForbidden)
		}
	}

	// Only requester or responder can cancel
	if status == models.StatusCancelled {
		if swapRequest.RequesterID != userID && swapRequest.ResponderID != userID {
			return nil, fmt.Errorf("only requester or responder can cancel requests: %w", apperrors.ErrForbidden)
		}
	}

	// Validate status transitions
	if swapRequest.Status != models.StatusPending && status != models.StatusCancelled {
		return nil, fmt.Errorf("can only modify pending requests: %w", apperrors.ErrWrongStatus)
	}

	// Update status
	swapRequest.Status = status
	err = s.db.Save(swapRequest).Error
	if err != nil {
		return nil, err
	}

	// Auto-create conversation when a swap is accepted
	if status == models.StatusAccepted {
		conv := &models.Conversation{SwapID: swapRequest.SwapID}
		if createErr := s.db.Create(conv).Error; createErr != nil {
			// Log but don't fail — conversation can be created on first chat open
			log.Printf("Info: could not auto-create conversation for swap %s: %v", swapID, createErr)
		}
	}

	// Notify the requester about the status change
	if s.notificationService != nil {
		_ = s.notificationService.CreateSwapStatusNotification(
			swapRequest.RequesterID,
			swapRequest.SwapID,
			string(status),
			swapRequest.OfferedSkill.Name,
		)
	}

	return swapRequest, nil
}

// DeleteSwapRequest deletes a swap request (only requester can delete)
func (s *swapService) DeleteSwapRequest(swapID uuid.UUID, userID uuid.UUID) error {
	swapRequest, err := s.GetSwapRequestByID(swapID)
	if err != nil {
		return err
	}

	// Only requester can delete
	if swapRequest.RequesterID != userID {
		return fmt.Errorf("only requester can delete swap requests: %w", apperrors.ErrForbidden)
	}

	// Can only delete pending requests
	if swapRequest.Status != models.StatusPending {
		return fmt.Errorf("can only delete pending requests: %w", apperrors.ErrWrongStatus)
	}

	return s.db.Delete(&models.SwapRequest{}, "swap_id = ?", swapID).Error
}

// GetSwapRequestsForUser retrieves organized swap requests for a user
func (s *swapService) GetSwapRequestsForUser(userID uuid.UUID) (*SwapRequestsResponse, error) {
	// Get sent requests
	sentFilter := SwapRequestFilter{Sent: true, Limit: 50}
	sent, err := s.GetUserSwapRequests(userID, sentFilter)
	if err != nil {
		return nil, err
	}

	// Get received requests
	receivedFilter := SwapRequestFilter{Received: true, Limit: 50}
	received, err := s.GetUserSwapRequests(userID, receivedFilter)
	if err != nil {
		return nil, err
	}

	return &SwapRequestsResponse{
		Sent:     sent,
		Received: received,
	}, nil
}

// GetPendingSwapRequests retrieves pending swap requests for a user
func (s *swapService) GetPendingSwapRequests(userID uuid.UUID) ([]models.SwapRequest, error) {
	status := models.StatusPending
	filter := SwapRequestFilter{Status: &status, Limit: 100}
	return s.GetUserSwapRequests(userID, filter)
}

// GetSwapHistory retrieves completed swap requests for a user
func (s *swapService) GetSwapHistory(userID uuid.UUID) ([]models.SwapRequest, error) {
	var swapRequests []models.SwapRequest
	err := s.db.Model(&models.SwapRequest{}).
		Preload("Requester").Preload("Responder").
		Preload("OfferedSkill").Preload("WantedSkill").
		Where("(requester_id = ? OR responder_id = ?) AND status IN ?",
			userID, userID, []models.SwapStatus{models.StatusAccepted, models.StatusRejected, models.StatusCancelled}).
		Order("updated_at DESC").
		Limit(50).
		Find(&swapRequests).Error

	return swapRequests, err
}

// FindPotentialMatches finds potential swap matches for a user
func (s *swapService) FindPotentialMatches(userID uuid.UUID) ([]SwapMatch, error) {
	var matches []SwapMatch

	// Get user's offered skills
	var userOfferedSkills []models.Skill
	err := s.db.Table("skills").
		Joins("JOIN user_skills_offered ON skills.skill_id = user_skills_offered.skill_id").
		Where("user_skills_offered.user_id = ?", userID).
		Find(&userOfferedSkills).Error
	if err != nil {
		return nil, err
	}

	// Get user's wanted skills
	var userWantedSkills []models.Skill
	err = s.db.Table("skills").
		Joins("JOIN user_skills_wanted ON skills.skill_id = user_skills_wanted.skill_id").
		Where("user_skills_wanted.user_id = ?", userID).
		Find(&userWantedSkills).Error
	if err != nil {
		return nil, err
	}

	// Find users who want what we offer and offer what we want
	for _, offeredSkill := range userOfferedSkills {
		for _, wantedSkill := range userWantedSkills {
			// Find users who want our offered skill AND offer our wanted skill
			var potentialUsers []models.User
			err := s.db.Table("users").
				Joins("JOIN user_skills_wanted ON users.user_id = user_skills_wanted.user_id").
				Joins("JOIN user_skills_offered ON users.user_id = user_skills_offered.user_id").
				Where("user_skills_wanted.skill_id = ? AND user_skills_offered.skill_id = ? AND users.user_id != ? AND users.is_public = true AND users.deleted_at IS NULL",
					offeredSkill.SkillID, wantedSkill.SkillID, userID).
				Find(&potentialUsers).Error

			if err != nil {
				continue
			}

			for _, user := range potentialUsers {
				// Calculate match score (simple algorithm for now)
				matchScore := 80 // Base score for mutual skill match

				matches = append(matches, SwapMatch{
					User:         user,
					OfferedSkill: offeredSkill,
					WantedSkill:  wantedSkill,
					MatchScore:   matchScore,
				})
			}
		}
	}

	// Limit results
	if len(matches) > 20 {
		matches = matches[:20]
	}

	return matches, nil
}

// ReportNoShow flags that the OTHER party did not show up for an accepted swap.
// Allowed only for participants, only on accepted swaps, and only after 24h
// have elapsed since the swap was accepted (UpdatedAt as proxy).
func (s *swapService) ReportNoShow(swapID, reporterID uuid.UUID, reason string) (*models.SwapRequest, error) {
swap, err := s.GetSwapRequestByID(swapID)
if err != nil {
return nil, err
}
if swap.RequesterID != reporterID && swap.ResponderID != reporterID {
return nil, fmt.Errorf("not a participant: %w", apperrors.ErrNotParticipant)
}
if swap.Status != models.StatusAccepted {
return nil, fmt.Errorf("swap must be accepted to report no-show: %w", apperrors.ErrWrongStatus)
}
if time.Since(swap.UpdatedAt) < 24*time.Hour {
return nil, fmt.Errorf("can only report no-show after 24h: %w", apperrors.ErrValidation)
}
if swap.NoShowFlag {
return nil, fmt.Errorf("no-show already reported: %w", apperrors.ErrConflict)
}
updates := map[string]any{
"no_show_flag":         true,
"no_show_reporter_id":  reporterID,
"no_show_reason":       reason,
}
if err := s.db.Model(&models.SwapRequest{}).Where("swap_id = ?", swapID).Updates(updates).Error; err != nil {
return nil, err
}
return s.GetSwapRequestByID(swapID)
}

// RaiseDispute attaches a free-form dispute reason to a swap. Either party
// can dispute at any non-pending status; admins triage via the admin queue.
func (s *swapService) RaiseDispute(swapID, userID uuid.UUID, reason string) (*models.SwapRequest, error) {
if reason == "" {
return nil, fmt.Errorf("reason required: %w", apperrors.ErrValidation)
}
swap, err := s.GetSwapRequestByID(swapID)
if err != nil {
return nil, err
}
if swap.RequesterID != userID && swap.ResponderID != userID {
return nil, fmt.Errorf("not a participant: %w", apperrors.ErrNotParticipant)
}
if err := s.db.Model(&models.SwapRequest{}).Where("swap_id = ?", swapID).
Update("dispute_reason", reason).Error; err != nil {
return nil, err
}
return s.GetSwapRequestByID(swapID)
}

// GetReliabilityStats returns per-user swap-reliability counters.
// Successful = completed swaps where this user participated AND no_show_flag is false.
// ReportedNoShows = swaps where this user is on the OTHER side of the no-show reporter.
func (s *swapService) GetReliabilityStats(userID uuid.UUID) (ReliabilityStats, error) {
var successful int64
if err := s.db.Model(&models.SwapRequest{}).
Where("(requester_id = ? OR responder_id = ?) AND status = ? AND no_show_flag = false",
userID, userID, models.StatusCompleted).
Count(&successful).Error; err != nil {
return ReliabilityStats{}, err
}
// no-shows reported AGAINST this user: they are the non-reporter participant on a flagged swap.
var noShows int64
if err := s.db.Model(&models.SwapRequest{}).
Where(`no_show_flag = true AND (
(requester_id = ? AND no_show_reporter_id = responder_id)
OR (responder_id = ? AND no_show_reporter_id = requester_id)
)`, userID, userID).
Count(&noShows).Error; err != nil {
return ReliabilityStats{}, err
}
denom := successful + noShows
stats := ReliabilityStats{
SuccessfulSwaps: int(successful),
ReportedNoShows: int(noShows),
Qualifies:       denom >= 3,
}
if denom > 0 {
stats.Score = float64(successful) / float64(denom)
}
return stats, nil
}

