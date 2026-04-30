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

// SessionService manages scheduled meet-ups for an accepted swap.
//
// Authorization rules (enforced here, not in handlers):
//   - Either swap participant may propose, list, accept, cancel, or complete.
//   - The proposer cannot accept their own proposal.
//   - Times must be in the future at proposal time and end > start.
type SessionService interface {
	Propose(swapID, proposerID uuid.UUID, start, end time.Time) (*models.Session, error)
	ListForSwap(swapID, viewerID uuid.UUID) ([]models.Session, error)
	Get(sessionID, viewerID uuid.UUID) (*models.Session, error)
	Accept(sessionID, userID uuid.UUID) (*models.Session, error)
	Cancel(sessionID, userID uuid.UUID) (*models.Session, error)
	Complete(sessionID, userID uuid.UUID) (*models.Session, error)
}

type sessionService struct {
	db *gorm.DB
}

func NewSessionService(db *gorm.DB) SessionService {
	return &sessionService{db: db}
}

// participants loads the accepted-swap pair for the given swap id and
// verifies the viewer is one of them.
func (s *sessionService) participants(swapID, viewerID uuid.UUID) (*models.SwapRequest, error) {
	var sw models.SwapRequest
	if err := s.db.Where("swap_id = ?", swapID).First(&sw).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("swap %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	if sw.RequesterID != viewerID && sw.ResponderID != viewerID {
		return nil, fmt.Errorf("not a participant: %w", apperrors.ErrNotParticipant)
	}
	return &sw, nil
}

func (s *sessionService) Propose(swapID, proposerID uuid.UUID, start, end time.Time) (*models.Session, error) {
	if !end.After(start) {
		return nil, fmt.Errorf("end must be after start: %w", apperrors.ErrValidation)
	}
	if start.Before(time.Now()) {
		return nil, fmt.Errorf("start must be in the future: %w", apperrors.ErrValidation)
	}
	sw, err := s.participants(swapID, proposerID)
	if err != nil {
		return nil, err
	}
	if sw.Status != models.StatusAccepted {
		return nil, fmt.Errorf("swap must be accepted to schedule sessions: %w", apperrors.ErrWrongStatus)
	}
	sess := &models.Session{
		SwapID:         swapID,
		ProposerID:     proposerID,
		ScheduledStart: start,
		ScheduledEnd:   end,
		Status:         models.SessionProposed,
	}
	if err := s.db.Create(sess).Error; err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *sessionService) ListForSwap(swapID, viewerID uuid.UUID) ([]models.Session, error) {
	if _, err := s.participants(swapID, viewerID); err != nil {
		return nil, err
	}
	var out []models.Session
	if err := s.db.Where("swap_id = ?", swapID).Order("scheduled_start ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// loadAndAuthorize fetches a session and verifies the user is a participant
// of its parent swap.
func (s *sessionService) loadAndAuthorize(sessionID, userID uuid.UUID) (*models.Session, *models.SwapRequest, error) {
	var sess models.Session
	if err := s.db.Where("session_id = ?", sessionID).First(&sess).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("session %w", apperrors.ErrNotFound)
		}
		return nil, nil, err
	}
	sw, err := s.participants(sess.SwapID, userID)
	if err != nil {
		return nil, nil, err
	}
	return &sess, sw, nil
}

func (s *sessionService) Get(sessionID, viewerID uuid.UUID) (*models.Session, error) {
	sess, _, err := s.loadAndAuthorize(sessionID, viewerID)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *sessionService) Accept(sessionID, userID uuid.UUID) (*models.Session, error) {
	sess, _, err := s.loadAndAuthorize(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if sess.ProposerID == userID {
		return nil, fmt.Errorf("proposer cannot accept own session: %w", apperrors.ErrSelfAction)
	}
	if sess.Status != models.SessionProposed {
		return nil, fmt.Errorf("session not in proposed state: %w", apperrors.ErrWrongStatus)
	}
	if err := s.db.Model(sess).Update("status", models.SessionAccepted).Error; err != nil {
		return nil, err
	}
	sess.Status = models.SessionAccepted
	return sess, nil
}

func (s *sessionService) Cancel(sessionID, userID uuid.UUID) (*models.Session, error) {
	sess, _, err := s.loadAndAuthorize(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if sess.Status == models.SessionCancelled || sess.Status == models.SessionCompleted {
		return nil, fmt.Errorf("session already terminal: %w", apperrors.ErrWrongStatus)
	}
	if err := s.db.Model(sess).Update("status", models.SessionCancelled).Error; err != nil {
		return nil, err
	}
	sess.Status = models.SessionCancelled
	return sess, nil
}

func (s *sessionService) Complete(sessionID, userID uuid.UUID) (*models.Session, error) {
	sess, _, err := s.loadAndAuthorize(sessionID, userID)
	if err != nil {
		return nil, err
	}
	if sess.Status != models.SessionAccepted {
		return nil, fmt.Errorf("only accepted sessions can be completed: %w", apperrors.ErrWrongStatus)
	}
	if err := s.db.Model(sess).Update("status", models.SessionCompleted).Error; err != nil {
		return nil, err
	}
	sess.Status = models.SessionCompleted
	return sess, nil
}
