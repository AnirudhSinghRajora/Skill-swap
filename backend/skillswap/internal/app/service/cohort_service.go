package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CohortService handles small private cohorts (kind=cohort) and large
// public drop-in rooms (kind=room). Both share the same schema; the
// service applies different defaults and rules per kind.
//
// Capacity races on Join are serialized with `SELECT ... FOR UPDATE` on
// the cohorts row inside a transaction.
type CohortService interface {
	Create(hostID uuid.UUID, in CreateCohortInput) (*models.Cohort, error)
	Get(cohortID uuid.UUID) (*models.Cohort, []models.CohortMember, error)
	List(filter ListCohortFilter) ([]models.Cohort, error)
	Join(cohortID, userID uuid.UUID) (*models.CohortMember, error)
	Leave(cohortID, userID uuid.UUID) error
	AddSession(cohortID, hostID uuid.UUID, start, end time.Time) (*models.CohortSession, error)
	ListSessions(cohortID uuid.UUID) ([]models.CohortSession, error)
	RemoveMember(cohortID, hostID, memberID uuid.UUID) error
	// AuthorizeJoin returns the cohort + cohort_session if the viewer is
	// allowed to join the LiveKit room for that session right now.
	// For kind=cohort, viewer must be a member; for kind=room, anyone may
	// join during the time window. The window is [start-15min, end+15min].
	AuthorizeJoin(cohortSessionID, viewerID uuid.UUID) (*models.Cohort, *models.CohortSession, error)
}

type CreateCohortInput struct {
	SkillID         uuid.UUID
	Title           string
	Description     string
	Capacity        int16
	Kind            models.CohortKind
	IsPublic        bool
	SchedulePattern *string
}

type ListCohortFilter struct {
	SkillID  *uuid.UUID
	Kind     *models.CohortKind
	Status   *models.CohortStatus
	HasSeats bool
	Limit    int
	Offset   int
}

type cohortService struct {
	db *gorm.DB
}

func NewCohortService(db *gorm.DB) CohortService {
	return &cohortService{db: db}
}

const maxHostingPerUser = 3

func (s *cohortService) Create(hostID uuid.UUID, in CreateCohortInput) (*models.Cohort, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 120 {
		return nil, fmt.Errorf("title must be 1-120 chars: %w", apperrors.ErrValidation)
	}
	if in.SkillID == uuid.Nil {
		return nil, fmt.Errorf("skill_id required: %w", apperrors.ErrValidation)
	}
	kind := in.Kind
	if kind == "" {
		kind = models.CohortKindCohort
	}
	cap := in.Capacity
	if cap <= 0 {
		if kind == models.CohortKindRoom {
			cap = 16
		} else {
			cap = 4
		}
	}
	if cap < 2 || cap > 50 {
		return nil, fmt.Errorf("capacity must be 2-50: %w", apperrors.ErrValidation)
	}
	if kind == models.CohortKindRoom && !in.IsPublic {
		// Rooms are always public; force the flag.
		in.IsPublic = true
	}

	// Spam-prevention: a single user cannot host more than maxHostingPerUser
	// open/in-progress cohorts at once.
	var hosting int64
	if err := s.db.Model(&models.Cohort{}).
		Where("host_id = ? AND status IN ?", hostID, []models.CohortStatus{models.CohortOpen, models.CohortInProgress}).
		Count(&hosting).Error; err != nil {
		return nil, err
	}
	if hosting >= maxHostingPerUser {
		return nil, fmt.Errorf("hosting limit reached: %w", apperrors.ErrRateLimited)
	}

	c := &models.Cohort{
		HostID:          hostID,
		SkillID:         in.SkillID,
		Title:           title,
		Description:     strings.TrimSpace(in.Description),
		Capacity:        cap,
		Kind:            kind,
		IsPublic:        in.IsPublic,
		Status:          models.CohortOpen,
		SchedulePattern: in.SchedulePattern,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(c).Error; err != nil {
			return err
		}
		// Auto-add host as member with role=host.
		if err := tx.Create(&models.CohortMember{
			CohortID: c.CohortID,
			UserID:   hostID,
			Role:     models.CohortRoleHost,
		}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *cohortService) Get(cohortID uuid.UUID) (*models.Cohort, []models.CohortMember, error) {
	var c models.Cohort
	if err := s.db.Where("cohort_id = ?", cohortID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("cohort %w", apperrors.ErrNotFound)
		}
		return nil, nil, err
	}
	var members []models.CohortMember
	if err := s.db.Where("cohort_id = ?", cohortID).Order("joined_at ASC").Find(&members).Error; err != nil {
		return nil, nil, err
	}
	return &c, members, nil
}

func (s *cohortService) List(f ListCohortFilter) ([]models.Cohort, error) {
	q := s.db.Model(&models.Cohort{}).Where("is_public = TRUE")
	if f.SkillID != nil {
		q = q.Where("skill_id = ?", *f.SkillID)
	}
	if f.Kind != nil {
		q = q.Where("kind = ?", *f.Kind)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	} else {
		q = q.Where("status IN ?", []models.CohortStatus{models.CohortOpen, models.CohortInProgress})
	}
	if f.HasSeats {
		// Subquery: members count < capacity.
		q = q.Where("(SELECT COUNT(*) FROM cohort_members m WHERE m.cohort_id = cohorts.cohort_id) < cohorts.capacity")
	}
	limit := f.Limit
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	q = q.Order("created_at DESC").Limit(limit).Offset(f.Offset)
	var out []models.Cohort
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *cohortService) Join(cohortID, userID uuid.UUID) (*models.CohortMember, error) {
	var m *models.CohortMember
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Lock the cohort row so concurrent joins serialize.
		var c models.Cohort
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("cohort_id = ?", cohortID).
			First(&c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("cohort %w", apperrors.ErrNotFound)
			}
			return err
		}
		if c.Status != models.CohortOpen && c.Status != models.CohortInProgress {
			return fmt.Errorf("cohort not joinable: %w", apperrors.ErrWrongStatus)
		}
		// Already a member?
		var existing models.CohortMember
		err := tx.Where("cohort_id = ? AND user_id = ?", cohortID, userID).First(&existing).Error
		if err == nil {
			return fmt.Errorf("already a member: %w", apperrors.ErrDuplicate)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// Capacity check inside the lock.
		var count int64
		if err := tx.Model(&models.CohortMember{}).Where("cohort_id = ?", cohortID).Count(&count).Error; err != nil {
			return err
		}
		if count >= int64(c.Capacity) {
			return fmt.Errorf("cohort full: %w", apperrors.ErrConflict)
		}
		nm := &models.CohortMember{
			CohortID: cohortID,
			UserID:   userID,
			Role:     models.CohortRoleMember,
		}
		if err := tx.Create(nm).Error; err != nil {
			return err
		}
		m = nm
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *cohortService) Leave(cohortID, userID uuid.UUID) error {
	var c models.Cohort
	if err := s.db.Where("cohort_id = ?", cohortID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("cohort %w", apperrors.ErrNotFound)
		}
		return err
	}
	if c.HostID == userID {
		return fmt.Errorf("host cannot leave; cancel instead: %w", apperrors.ErrSelfAction)
	}
	res := s.db.Where("cohort_id = ? AND user_id = ?", cohortID, userID).Delete(&models.CohortMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("not a member: %w", apperrors.ErrNotFound)
	}
	return nil
}

func (s *cohortService) AddSession(cohortID, hostID uuid.UUID, start, end time.Time) (*models.CohortSession, error) {
	if !end.After(start) {
		return nil, fmt.Errorf("end must be after start: %w", apperrors.ErrValidation)
	}
	if start.Before(time.Now()) {
		return nil, fmt.Errorf("start must be in the future: %w", apperrors.ErrValidation)
	}
	var c models.Cohort
	if err := s.db.Where("cohort_id = ?", cohortID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cohort %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	if c.HostID != hostID {
		return nil, fmt.Errorf("only host may schedule: %w", apperrors.ErrForbidden)
	}
	cs := &models.CohortSession{
		CohortID:       cohortID,
		ScheduledStart: start,
		ScheduledEnd:   end,
		Status:         "scheduled",
	}
	if err := s.db.Create(cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

func (s *cohortService) ListSessions(cohortID uuid.UUID) ([]models.CohortSession, error) {
	var out []models.CohortSession
	if err := s.db.Where("cohort_id = ?", cohortID).Order("scheduled_start ASC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *cohortService) RemoveMember(cohortID, hostID, memberID uuid.UUID) error {
	var c models.Cohort
	if err := s.db.Where("cohort_id = ?", cohortID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("cohort %w", apperrors.ErrNotFound)
		}
		return err
	}
	if c.HostID != hostID {
		return fmt.Errorf("only host may remove members: %w", apperrors.ErrForbidden)
	}
	if memberID == hostID {
		return fmt.Errorf("host cannot remove self: %w", apperrors.ErrSelfAction)
	}
	res := s.db.Where("cohort_id = ? AND user_id = ?", cohortID, memberID).Delete(&models.CohortMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("not a member: %w", apperrors.ErrNotFound)
	}
	return nil
}

// AuthorizeJoin enforces the time window + kind-specific membership rules.
func (s *cohortService) AuthorizeJoin(cohortSessionID, viewerID uuid.UUID) (*models.Cohort, *models.CohortSession, error) {
	var cs models.CohortSession
	if err := s.db.Where("cohort_session_id = ?", cohortSessionID).First(&cs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("cohort session %w", apperrors.ErrNotFound)
		}
		return nil, nil, err
	}
	now := time.Now()
	open := cs.ScheduledStart.Add(-15 * time.Minute)
	close := cs.ScheduledEnd.Add(15 * time.Minute)
	if now.Before(open) || now.After(close) {
		return nil, nil, fmt.Errorf("session not within join window: %w", apperrors.ErrWrongStatus)
	}
	var c models.Cohort
	if err := s.db.Where("cohort_id = ?", cs.CohortID).First(&c).Error; err != nil {
		return nil, nil, err
	}
	if c.Kind == models.CohortKindCohort {
		var m models.CohortMember
		if err := s.db.Where("cohort_id = ? AND user_id = ?", c.CohortID, viewerID).First(&m).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, nil, fmt.Errorf("not a member: %w", apperrors.ErrForbidden)
			}
			return nil, nil, err
		}
	}
	// Rooms (kind=room) are open to any authenticated user during the window.
	return &c, &cs, nil
}
