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

// ModerationService manages user blocks and abuse reports.
type ModerationService struct {
	db *gorm.DB
}

func NewModerationService(db *gorm.DB) *ModerationService {
	return &ModerationService{db: db}
}

// Block adds a (blocker, blocked) row. Idempotent: re-blocking is a no-op.
// Self-blocking is rejected.
func (s *ModerationService) Block(blocker, blocked uuid.UUID) error {
	if blocker == blocked {
		return fmt.Errorf("cannot block yourself: %w", apperrors.ErrSelfAction)
	}
	block := &models.UserBlock{BlockerID: blocker, BlockedID: blocked}
	// ON CONFLICT DO NOTHING
	return s.db.Where("blocker_id = ? AND blocked_id = ?", blocker, blocked).
		FirstOrCreate(block).Error
}

func (s *ModerationService) Unblock(blocker, blocked uuid.UUID) error {
	return s.db.Where("blocker_id = ? AND blocked_id = ?", blocker, blocked).
		Delete(&models.UserBlock{}).Error
}

// IsBlockedEitherWay returns true if a has blocked b OR b has blocked a.
// This is the function every list/chat/swap surface should consult.
func (s *ModerationService) IsBlockedEitherWay(a, b uuid.UUID) (bool, error) {
	var count int64
	err := s.db.Model(&models.UserBlock{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)",
			a, b, b, a).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// BlockedUserIDs returns the set of users that are blocked-by-me OR who-blocked-me.
// Used by list endpoints to filter results.
func (s *ModerationService) BlockedUserIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.db.Raw(`
		SELECT blocked_id FROM user_blocks WHERE blocker_id = ?
		UNION
		SELECT blocker_id FROM user_blocks WHERE blocked_id = ?
	`, userID, userID).Scan(&ids).Error
	return ids, err
}

// ListBlocks returns the users I have blocked (one direction only).
func (s *ModerationService) ListBlocks(blocker uuid.UUID) ([]models.UserBlock, error) {
	var rows []models.UserBlock
	err := s.db.Where("blocker_id = ?", blocker).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

// CreateReport files an abuse report. Reasons must be non-empty.
func (s *ModerationService) CreateReport(reporterID uuid.UUID, targetUserID uuid.UUID, kind models.ReportTargetKind, targetID *uuid.UUID, reason string) (*models.Report, error) {
	if reporterID == targetUserID {
		return nil, fmt.Errorf("cannot report yourself: %w", apperrors.ErrSelfAction)
	}
	if reason == "" {
		return nil, fmt.Errorf("reason is required: %w", apperrors.ErrValidation)
	}
	report := &models.Report{
		ReporterID:   reporterID,
		TargetUserID: targetUserID,
		TargetKind:   kind,
		TargetID:     targetID,
		Reason:       reason,
		Status:       models.ReportStatusOpen,
	}
	if err := s.db.Create(report).Error; err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}
	return report, nil
}

// ListReports returns reports filtered by status. Admin-only at the handler layer.
func (s *ModerationService) ListReports(status models.ReportStatus, limit, offset int) ([]models.Report, int64, error) {
	q := s.db.Model(&models.Report{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []models.Report
	if limit <= 0 {
		limit = 50
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

// ResolveReport marks a report resolved/dismissed with a resolver note.
func (s *ModerationService) ResolveReport(reportID, resolverID uuid.UUID, status models.ReportStatus, note string) error {
	if status != models.ReportStatusResolved && status != models.ReportStatusDismissed && status != models.ReportStatusReviewing {
		return fmt.Errorf("invalid status: %w", apperrors.ErrValidation)
	}
	now := time.Now()
	updates := map[string]any{
		"status":          status,
		"resolver_id":     resolverID,
		"resolved_at":     now,
		"resolution_note": note,
	}
	res := s.db.Model(&models.Report{}).Where("report_id = ?", reportID).Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("report %w", apperrors.ErrNotFound)
	}
	return nil
}

// GetReport fetches a single report. Admin-only.
func (s *ModerationService) GetReport(reportID uuid.UUID) (*models.Report, error) {
	var r models.Report
	err := s.db.Where("report_id = ?", reportID).First(&r).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("report %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &r, nil
}
