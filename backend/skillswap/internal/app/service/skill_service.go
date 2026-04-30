package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SkillService interface {
	// Skill CRUD operations
	GetAllSkills() ([]models.Skill, error)
	GetSkillByID(skillID uuid.UUID) (*models.Skill, error)
	CreateSkill(name string) (*models.Skill, error)
	UpdateSkill(skillID uuid.UUID, name string) (*models.Skill, error)
	DeleteSkill(skillID uuid.UUID) error

	// Taxonomy & alias resolution
	ResolveOrCreate(name string) (*models.Skill, error)
	ListCategories() ([]models.SkillCategory, error)

	// User skill management
	AddOfferedSkill(userID, skillID uuid.UUID) error
	RemoveOfferedSkill(userID, skillID uuid.UUID) error
	AddWantedSkill(userID, skillID uuid.UUID) error
	RemoveWantedSkill(userID, skillID uuid.UUID) error

	// User skill queries
	GetUserOfferedSkills(userID uuid.UUID) ([]models.Skill, error)
	GetUserWantedSkills(userID uuid.UUID) ([]models.Skill, error)
	GetUsersWithOfferedSkill(skillID uuid.UUID) ([]models.User, error)
	GetUsersWithWantedSkill(skillID uuid.UUID) ([]models.User, error)
}

type skillService struct {
	db *gorm.DB
}

func NewSkillService(db *gorm.DB) SkillService {
	return &skillService{db: db}
}

// GetAllSkills retrieves all available skills
func (s *skillService) GetAllSkills() ([]models.Skill, error) {
	var skills []models.Skill
	err := s.db.Find(&skills).Error
	return skills, err
}

// GetSkillByID retrieves a skill by its ID
func (s *skillService) GetSkillByID(skillID uuid.UUID) (*models.Skill, error) {
	var skill models.Skill
	err := s.db.First(&skill, "skill_id = ?", skillID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("skill not found: %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &skill, nil
}

// CreateSkill creates a new skill
func (s *skillService) CreateSkill(name string) (*models.Skill, error) {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 2 || len(trimmed) > 100 {
		return nil, fmt.Errorf("skill name must be 2-100 characters: %w", apperrors.ErrValidation)
	}

	skill := &models.Skill{
		Name: trimmed,
	}

	err := s.db.Create(skill).Error
	if err != nil {
		return nil, err
	}

	return skill, nil
}

// UpdateSkill updates an existing skill
func (s *skillService) UpdateSkill(skillID uuid.UUID, name string) (*models.Skill, error) {
	skill, err := s.GetSkillByID(skillID)
	if err != nil {
		return nil, err
	}

	skill.Name = name
	err = s.db.Save(skill).Error
	if err != nil {
		return nil, err
	}

	return skill, nil
}

// DeleteSkill deletes a skill (admin only)
func (s *skillService) DeleteSkill(skillID uuid.UUID) error {
	// Check if skill exists
	_, err := s.GetSkillByID(skillID)
	if err != nil {
		return err
	}

	// Use transaction to ensure count checks and delete are atomic
	return s.db.Transaction(func(tx *gorm.DB) error {
		var offeredCount, wantedCount int64
		if err := tx.Model(&models.UserSkillOffered{}).Where("skill_id = ?", skillID).Count(&offeredCount).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.UserSkillWanted{}).Where("skill_id = ?", skillID).Count(&wantedCount).Error; err != nil {
			return err
		}

		if offeredCount > 0 || wantedCount > 0 {
			return fmt.Errorf("skill is in use and cannot be deleted: %w", apperrors.ErrInUse)
		}

		return tx.Delete(&models.Skill{}, "skill_id = ?", skillID).Error
	})
}

// AddOfferedSkill adds a skill to user's offered skills
func (s *skillService) AddOfferedSkill(userID, skillID uuid.UUID) error {
	// Check if skill exists
	_, err := s.GetSkillByID(skillID)
	if err != nil {
		return err
	}

	// Check if already exists
	var count int64
	if err := s.db.Model(&models.UserSkillOffered{}).Where("user_id = ? AND skill_id = ?", userID, skillID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("skill already in offered skills: %w", apperrors.ErrConflict)
	}

	userSkill := &models.UserSkillOffered{
		UserID:  userID,
		SkillID: skillID,
	}

	return s.db.Create(userSkill).Error
}

// RemoveOfferedSkill removes a skill from user's offered skills
func (s *skillService) RemoveOfferedSkill(userID, skillID uuid.UUID) error {
	result := s.db.Delete(&models.UserSkillOffered{}, "user_id = ? AND skill_id = ?", userID, skillID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("offered skill not found: %w", apperrors.ErrNotFound)
	}
	return nil
}

// AddWantedSkill adds a skill to user's wanted skills
func (s *skillService) AddWantedSkill(userID, skillID uuid.UUID) error {
	// Check if skill exists
	_, err := s.GetSkillByID(skillID)
	if err != nil {
		return err
	}

	// Check if already exists
	var count int64
	if err := s.db.Model(&models.UserSkillWanted{}).Where("user_id = ? AND skill_id = ?", userID, skillID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("skill already in wanted skills: %w", apperrors.ErrConflict)
	}

	userSkill := &models.UserSkillWanted{
		UserID:  userID,
		SkillID: skillID,
	}

	return s.db.Create(userSkill).Error
}

// RemoveWantedSkill removes a skill from user's wanted skills
func (s *skillService) RemoveWantedSkill(userID, skillID uuid.UUID) error {
	result := s.db.Delete(&models.UserSkillWanted{}, "user_id = ? AND skill_id = ?", userID, skillID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("wanted skill not found: %w", apperrors.ErrNotFound)
	}
	return nil
}

// GetUserOfferedSkills retrieves all skills offered by a user
func (s *skillService) GetUserOfferedSkills(userID uuid.UUID) ([]models.Skill, error) {
	var skills []models.Skill
	err := s.db.Table("skills").
		Joins("JOIN user_skills_offered ON skills.skill_id = user_skills_offered.skill_id").
		Where("user_skills_offered.user_id = ?", userID).
		Find(&skills).Error

	return skills, err
}

// GetUserWantedSkills retrieves all skills wanted by a user
func (s *skillService) GetUserWantedSkills(userID uuid.UUID) ([]models.Skill, error) {
	var skills []models.Skill
	err := s.db.Table("skills").
		Joins("JOIN user_skills_wanted ON skills.skill_id = user_skills_wanted.skill_id").
		Where("user_skills_wanted.user_id = ?", userID).
		Find(&skills).Error

	return skills, err
}

// GetUsersWithOfferedSkill retrieves all users who offer a specific skill
func (s *skillService) GetUsersWithOfferedSkill(skillID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := s.db.Table("users").
		Joins("JOIN user_skills_offered ON users.user_id = user_skills_offered.user_id").
		Where("user_skills_offered.skill_id = ? AND users.is_public = true AND users.deleted_at IS NULL", skillID).
		Find(&users).Error

	return users, err
}

// GetUsersWithWantedSkill retrieves all users who want a specific skill
func (s *skillService) GetUsersWithWantedSkill(skillID uuid.UUID) ([]models.User, error) {
	var users []models.User
	err := s.db.Table("users").
		Joins("JOIN user_skills_wanted ON users.user_id = user_skills_wanted.user_id").
		Where("user_skills_wanted.skill_id = ? AND users.is_public = true AND users.deleted_at IS NULL", skillID).
		Find(&users).Error

	return users, err
}

// ResolveOrCreate resolves a free-text skill name to a canonical skill,
// creating it if it does not exist. Lookup order: exact (case-insensitive)
// → alias → create new.
func (s *skillService) ResolveOrCreate(name string) (*models.Skill, error) {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 2 || len(trimmed) > 100 {
		return nil, fmt.Errorf("skill name must be 2-100 characters: %w", apperrors.ErrValidation)
	}

	// 1. Exact match (case-insensitive)
	var skill models.Skill
	if err := s.db.Where("LOWER(name) = LOWER(?)", trimmed).First(&skill).Error; err == nil {
		return &skill, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 2. Alias lookup
	var alias models.SkillAlias
	if err := s.db.Where("LOWER(alias_name) = LOWER(?)", trimmed).First(&alias).Error; err == nil {
		if err := s.db.First(&skill, "skill_id = ?", alias.CanonicalSkillID).Error; err == nil {
			return &skill, nil
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 3. Create new skill (uncategorized)
	created := &models.Skill{Name: trimmed}
	if err := s.db.Create(created).Error; err != nil {
		return nil, err
	}
	return created, nil
}

// ListCategories returns all skill categories ordered by sort_order then name.
func (s *skillService) ListCategories() ([]models.SkillCategory, error) {
	var cats []models.SkillCategory
	err := s.db.Order("sort_order ASC, name ASC").Find(&cats).Error
	return cats, err
}
