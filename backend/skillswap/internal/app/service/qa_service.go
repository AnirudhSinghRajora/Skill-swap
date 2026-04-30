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

// QAService handles the topic Q&A workflow: ask a question, post answers,
// upvote either, accept one answer per question.
//
// Vote toggling: we treat (voter_id, target_kind, target_id) as a unique
// key. The first call inserts a row and increments the count; the second
// call deletes the row and decrements. All counter mutations happen in a
// single transaction with the vote row write to keep the count and the
// vote rows consistent. Counts are stored denormalised on the parent row
// for cheap list rendering.
//
// Soft-delete (gorm.DeletedAt) is used for questions/answers so we can
// surface removed items to moderators without losing history. Listings
// transparently exclude soft-deleted rows via GORM's default scope.
type QAService interface {
	AskQuestion(asker uuid.UUID, in AskQuestionInput) (*models.Question, error)
	GetQuestion(id uuid.UUID) (*models.Question, []models.Answer, error)
	ListQuestions(filter ListQuestionsFilter) ([]models.Question, error)
	UpdateQuestion(id, asker uuid.UUID, title, body string) (*models.Question, error)
	DeleteQuestion(id, asker uuid.UUID) error

	PostAnswer(question, author uuid.UUID, body string) (*models.Answer, error)
	UpdateAnswer(id, author uuid.UUID, body string) (*models.Answer, error)
	DeleteAnswer(id, author uuid.UUID) error
	AcceptAnswer(answerID, asker uuid.UUID) error

	// ToggleVote inserts or removes a vote for (voter, kind, target) and
	// keeps the parent row's upvote_count in sync. The returned bool is
	// true when the vote is now present (added), false when it was
	// removed.
	ToggleVote(voter, target uuid.UUID, kind models.VoteTargetKind) (added bool, count int32, err error)
}

type AskQuestionInput struct {
	SkillID *uuid.UUID
	Title   string
	Body    string
}

type ListQuestionsFilter struct {
	SkillID *uuid.UUID
	Q       string // case-insensitive title prefix search
	Limit   int
	Offset  int
}

type qaService struct {
	db *gorm.DB
}

func NewQAService(db *gorm.DB) QAService { return &qaService{db: db} }

// titleMax is enforced both client- and server-side. Body has no hard
// cap here — the Postgres TEXT column is the only limit — because users
// pasting code or long explanations is expected.
const qaTitleMax = 200

func (s *qaService) AskQuestion(asker uuid.UUID, in AskQuestionInput) (*models.Question, error) {
	title := strings.TrimSpace(in.Title)
	body := strings.TrimSpace(in.Body)
	if title == "" || len(title) > qaTitleMax {
		return nil, fmt.Errorf("title length 1..%d: %w", qaTitleMax, apperrors.ErrValidation)
	}
	if body == "" {
		return nil, fmt.Errorf("body required: %w", apperrors.ErrValidation)
	}
	q := &models.Question{
		AskerID: asker,
		SkillID: in.SkillID,
		Title:   title,
		Body:    body,
	}
	if err := s.db.Create(q).Error; err != nil {
		return nil, err
	}
	return q, nil
}

func (s *qaService) GetQuestion(id uuid.UUID) (*models.Question, []models.Answer, error) {
	var q models.Question
	if err := s.db.Where("question_id = ?", id).First(&q).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, fmt.Errorf("question %w", apperrors.ErrNotFound)
		}
		return nil, nil, err
	}
	var answers []models.Answer
	// Order: accepted answer (if any) first, then by upvote desc, then by
	// creation. We do this in two queries to keep the SQL simple; the
	// payload is small enough that the second round-trip is fine.
	if err := s.db.Where("question_id = ?", id).
		Order("upvote_count DESC, created_at ASC").
		Find(&answers).Error; err != nil {
		return nil, nil, err
	}
	if q.AcceptedID != nil {
		// Move accepted answer to front in-memory.
		for i, a := range answers {
			if a.AnswerID == *q.AcceptedID && i > 0 {
				answers[0], answers[i] = answers[i], answers[0]
				break
			}
		}
	}
	return &q, answers, nil
}

func (s *qaService) ListQuestions(f ListQuestionsFilter) ([]models.Question, error) {
	if f.Limit <= 0 {
		f.Limit = 30
	}
	if f.Limit > 100 {
		f.Limit = 100
	}
	q := s.db.Model(&models.Question{})
	if f.SkillID != nil {
		q = q.Where("skill_id = ?", *f.SkillID)
	}
	if strings.TrimSpace(f.Q) != "" {
		q = q.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(f.Q))+"%")
	}
	var out []models.Question
	if err := q.Order("created_at DESC").Limit(f.Limit).Offset(f.Offset).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (s *qaService) UpdateQuestion(id, asker uuid.UUID, title, body string) (*models.Question, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" || len(title) > qaTitleMax {
		return nil, fmt.Errorf("title length 1..%d: %w", qaTitleMax, apperrors.ErrValidation)
	}
	if body == "" {
		return nil, fmt.Errorf("body required: %w", apperrors.ErrValidation)
	}
	var q models.Question
	if err := s.db.Where("question_id = ?", id).First(&q).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("question %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	if q.AskerID != asker {
		return nil, apperrors.ErrForbidden
	}
	q.Title = title
	q.Body = body
	if err := s.db.Save(&q).Error; err != nil {
		return nil, err
	}
	return &q, nil
}

func (s *qaService) DeleteQuestion(id, asker uuid.UUID) error {
	var q models.Question
	if err := s.db.Where("question_id = ?", id).First(&q).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("question %w", apperrors.ErrNotFound)
		}
		return err
	}
	if q.AskerID != asker {
		return apperrors.ErrForbidden
	}
	return s.db.Delete(&q).Error
}

func (s *qaService) PostAnswer(question, author uuid.UUID, body string) (*models.Answer, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("body required: %w", apperrors.ErrValidation)
	}
	var q models.Question
	if err := s.db.Where("question_id = ?", question).First(&q).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("question %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	a := &models.Answer{QuestionID: question, AuthorID: author, Body: body}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(a).Error; err != nil {
			return err
		}
		return tx.Model(&models.Question{}).Where("question_id = ?", question).
			UpdateColumn("answer_count", gorm.Expr("answer_count + 1")).Error
	})
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *qaService) UpdateAnswer(id, author uuid.UUID, body string) (*models.Answer, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, fmt.Errorf("body required: %w", apperrors.ErrValidation)
	}
	var a models.Answer
	if err := s.db.Where("answer_id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("answer %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	if a.AuthorID != author {
		return nil, apperrors.ErrForbidden
	}
	a.Body = body
	if err := s.db.Save(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *qaService) DeleteAnswer(id, author uuid.UUID) error {
	var a models.Answer
	if err := s.db.Where("answer_id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("answer %w", apperrors.ErrNotFound)
		}
		return err
	}
	if a.AuthorID != author {
		return apperrors.ErrForbidden
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&a).Error; err != nil {
			return err
		}
		// If this was the accepted answer, clear it.
		if err := tx.Model(&models.Question{}).
			Where("question_id = ? AND accepted_answer_id = ?", a.QuestionID, a.AnswerID).
			UpdateColumn("accepted_answer_id", nil).Error; err != nil {
			return err
		}
		return tx.Model(&models.Question{}).
			Where("question_id = ? AND answer_count > 0", a.QuestionID).
			UpdateColumn("answer_count", gorm.Expr("answer_count - 1")).Error
	})
}

func (s *qaService) AcceptAnswer(answerID, asker uuid.UUID) error {
	var a models.Answer
	if err := s.db.Where("answer_id = ?", answerID).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("answer %w", apperrors.ErrNotFound)
		}
		return err
	}
	var q models.Question
	if err := s.db.Where("question_id = ?", a.QuestionID).First(&q).Error; err != nil {
		return err
	}
	if q.AskerID != asker {
		return apperrors.ErrForbidden
	}
	q.AcceptedID = &a.AnswerID
	return s.db.Save(&q).Error
}

func (s *qaService) ToggleVote(voter, target uuid.UUID, kind models.VoteTargetKind) (bool, int32, error) {
	if kind != models.VoteTargetQuestion && kind != models.VoteTargetAnswer {
		return false, 0, fmt.Errorf("bad target_kind: %w", apperrors.ErrValidation)
	}
	var added bool
	var count int32
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing models.Vote
		err := tx.Where("voter_id = ? AND target_kind = ? AND target_id = ?", voter, kind, target).
			First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Insert + increment.
			if err := tx.Create(&models.Vote{VoterID: voter, TargetKind: kind, TargetID: target}).Error; err != nil {
				return err
			}
			added = true
			return s.bumpCount(tx, kind, target, +1, &count)
		}
		// Delete + decrement.
		if err := tx.Delete(&existing).Error; err != nil {
			return err
		}
		added = false
		return s.bumpCount(tx, kind, target, -1, &count)
	})
	return added, count, err
}

// bumpCount adjusts the parent row's denormalised upvote_count and writes
// back the resulting value via RETURNING-style read so callers can echo
// it without a second round-trip.
func (s *qaService) bumpCount(tx *gorm.DB, kind models.VoteTargetKind, target uuid.UUID, delta int, out *int32) error {
	var table, idCol string
	switch kind {
	case models.VoteTargetQuestion:
		table, idCol = "questions", "question_id"
	case models.VoteTargetAnswer:
		table, idCol = "answers", "answer_id"
	default:
		return fmt.Errorf("unsupported kind %q: %w", kind, apperrors.ErrValidation)
	}
	expr := "upvote_count + ?"
	if delta < 0 {
		// GREATEST guards against any drift that would push the count
		// below zero (e.g. a manual DB tweak).
		expr = "GREATEST(upvote_count + ?, 0)"
	}
	if err := tx.Table(table).Where(idCol+" = ?", target).
		Update("upvote_count", gorm.Expr(expr, delta)).Error; err != nil {
		return err
	}
	return tx.Table(table).Select("upvote_count").Where(idCol+" = ?", target).
		Row().Scan(out)
}

// guard against unused imports when the tests below get pruned
var _ = time.Now
