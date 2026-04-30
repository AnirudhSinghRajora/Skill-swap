package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
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
	IntroMessage   string    `json:"intro_message,omitempty" binding:"max=280"`
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

	// Per-user soft rate limit: at most 5 swap requests created per rolling 24h.
	var dailyCount int64
	if err := s.db.Model(&models.SwapRequest{}).
		Where("requester_id = ? AND created_at > ?", req.RequesterID, time.Now().Add(-24*time.Hour)).
		Count(&dailyCount).Error; err != nil {
		return nil, fmt.Errorf("failed to check daily quota: %w", err)
	}
	if dailyCount >= 5 {
		return nil, fmt.Errorf("daily swap-request limit reached (5/day): %w", apperrors.ErrRateLimited)
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
		if intro := strings.TrimSpace(req.IntroMessage); intro != "" {
			swapRequest.IntroMessage = &intro
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

// FindPotentialMatches finds potential swap matches for a user using a
// bidirectional score computed entirely in SQL.
//
// Score formula (per candidate user u, vs viewer v):
//   forward = COUNT(skills v offers AND u wants)
//   reverse = COUNT(skills u offers AND v wants)
//   raw     = forward + reverse + 2 * LEAST(forward, reverse)
//   penalty = SUM(|level_v - level_u|) over the matched offered/wanted pairs (capped)
//   recency = exp(-days_since_active / 30)            -- users.updated_at proxy
//   score   = GREATEST(0, (raw - penalty)) * recency
//
// One row is returned per (candidate, representative offered skill, representative
// wanted skill); the representative pair is the one with the smallest level gap.
func (s *swapService) FindPotentialMatches(userID uuid.UUID) ([]SwapMatch, error) {
	type row struct {
		UserID         uuid.UUID `gorm:"column:user_id"`
		Name           string    `gorm:"column:name"`
		Email          string    `gorm:"column:email"`
		Location       *string   `gorm:"column:location"`
		IsPublic       bool      `gorm:"column:is_public"`
		UpdatedAt      time.Time `gorm:"column:updated_at"`
		OfferedSkillID uuid.UUID `gorm:"column:offered_skill_id"`
		OfferedName    string    `gorm:"column:offered_name"`
		WantedSkillID  uuid.UUID `gorm:"column:wanted_skill_id"`
		WantedName     string    `gorm:"column:wanted_name"`
		Score          float64   `gorm:"column:score"`
	}

	const q = `
WITH
viewer_offered AS (
    SELECT skill_id, level FROM user_skills_offered WHERE user_id = ?
),
viewer_wanted AS (
    SELECT skill_id, level FROM user_skills_wanted WHERE user_id = ?
),
-- forward: viewer offers skill that candidate wants
fwd AS (
    SELECT uw.user_id AS cand_id,
           uw.skill_id AS skill_id,
           ABS(COALESCE(vo.level,2) - COALESCE(uw.level,2)) AS gap
    FROM user_skills_wanted uw
    JOIN viewer_offered vo ON vo.skill_id = uw.skill_id
    WHERE uw.user_id <> ?
),
-- reverse: candidate offers skill that viewer wants
rev AS (
    SELECT uo.user_id AS cand_id,
           uo.skill_id AS skill_id,
           ABS(COALESCE(vw.level,2) - COALESCE(uo.level,2)) AS gap
    FROM user_skills_offered uo
    JOIN viewer_wanted vw ON vw.skill_id = uo.skill_id
    WHERE uo.user_id <> ?
),
agg AS (
    SELECT
        c.cand_id,
        (SELECT COUNT(*) FROM fwd f WHERE f.cand_id = c.cand_id) AS fwd_cnt,
        (SELECT COUNT(*) FROM rev r WHERE r.cand_id = c.cand_id) AS rev_cnt,
        COALESCE((SELECT SUM(gap) FROM fwd f WHERE f.cand_id = c.cand_id),0)
        + COALESCE((SELECT SUM(gap) FROM rev r WHERE r.cand_id = c.cand_id),0) AS total_gap
    FROM (SELECT DISTINCT cand_id FROM (SELECT cand_id FROM fwd UNION ALL SELECT cand_id FROM rev) u) c
),
rep_fwd AS (
    SELECT DISTINCT ON (cand_id) cand_id, skill_id FROM fwd ORDER BY cand_id, gap ASC
),
rep_rev AS (
    SELECT DISTINCT ON (cand_id) cand_id, skill_id FROM rev ORDER BY cand_id, gap ASC
)
SELECT
    u.user_id, u.name, u.email, u.location, u.is_public, u.updated_at,
    rf.skill_id AS offered_skill_id, so.name AS offered_name,
    rr.skill_id AS wanted_skill_id,  sw.name AS wanted_name,
    GREATEST(
      0::float8,
      (a.fwd_cnt + a.rev_cnt + 2 * LEAST(a.fwd_cnt, a.rev_cnt) - LEAST(a.total_gap, a.fwd_cnt + a.rev_cnt))::float8
    ) * EXP(-EXTRACT(EPOCH FROM (NOW() - u.updated_at)) / (60.0*60.0*24.0*30.0)) AS score
FROM agg a
JOIN users u ON u.user_id = a.cand_id
LEFT JOIN rep_fwd rf ON rf.cand_id = a.cand_id
LEFT JOIN rep_rev rr ON rr.cand_id = a.cand_id
LEFT JOIN skills so ON so.skill_id = rf.skill_id
LEFT JOIN skills sw ON sw.skill_id = rr.skill_id
WHERE u.is_public = TRUE AND u.deleted_at IS NULL
  AND a.fwd_cnt > 0 AND a.rev_cnt > 0
ORDER BY score DESC
LIMIT 20
`

	var rows []row
	if err := s.db.Raw(q, userID, userID, userID, userID).Scan(&rows).Error; err != nil {
		return nil, err
	}

	matches := make([]SwapMatch, 0, len(rows))
	for _, r := range rows {
		matches = append(matches, SwapMatch{
			User: models.User{
				UserID:    r.UserID,
				Name:      r.Name,
				Email:     r.Email,
				Location:  r.Location,
				IsPublic:  r.IsPublic,
				UpdatedAt: r.UpdatedAt,
			},
			OfferedSkill: models.Skill{SkillID: r.OfferedSkillID, Name: r.OfferedName},
			WantedSkill:  models.Skill{SkillID: r.WantedSkillID, Name: r.WantedName},
			MatchScore:   int(r.Score*10 + 0.5), // scale to a stable integer
		})
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

