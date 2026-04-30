package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetBySlug(slug string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	List(limit, offset int, filters UserFilters) ([]*models.User, int64, error)

	// E2EE key management
	UpdateE2EEKeys(userID uuid.UUID, publicKey, encryptedBackup string) error
	GetPublicKey(userID uuid.UUID) (string, error)
	GetKeyBackup(userID uuid.UUID) (string, error)
}

type UserFilters struct {
	IsPublic   *bool
	Location   string
	SearchTerm string
	// Geo: when Lat/Lng/WithinKm provided, restrict to users with non-null
	// coordinates within haversine distance.
	Lat       *float64
	Lng       *float64
	WithinKm  *float64
	RemoteOK  *bool // when true, ALSO include users with is_remote_ok=true (OR'd with radius)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	user.UserID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("SkillsOffered.Skill").
		Preload("SkillsWanted.Skill").
		Where("user_id = ?", id).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetBySlug(slug string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("SkillsOffered.Skill").
		Preload("SkillsWanted.Skill").
		Where("slug = ?", slug).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user %w", apperrors.ErrNotFound)
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	user.UpdatedAt = time.Now()
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Where("user_id = ?", id).Delete(&models.User{}).Error
}

func (r *userRepository) UpdateE2EEKeys(userID uuid.UUID, publicKey, encryptedBackup string) error {
	result := r.db.Model(&models.User{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"public_key":           publicKey,
			"encrypted_key_backup": encryptedBackup,
			"updated_at":           time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user %w", apperrors.ErrNotFound)
	}
	return nil
}

func (r *userRepository) GetPublicKey(userID uuid.UUID) (string, error) {
	var user models.User
	err := r.db.Select("public_key").
		Where("user_id = ?", userID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("user %w", apperrors.ErrNotFound)
		}
		return "", err
	}
	if user.PublicKey == nil || *user.PublicKey == "" {
		return "", fmt.Errorf("public key not set: %w", apperrors.ErrNotFound)
	}
	return *user.PublicKey, nil
}

func (r *userRepository) GetKeyBackup(userID uuid.UUID) (string, error) {
	var user models.User
	err := r.db.Select("encrypted_key_backup").
		Where("user_id = ?", userID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("user %w", apperrors.ErrNotFound)
		}
		return "", err
	}
	if user.EncryptedKeyBackup == nil || *user.EncryptedKeyBackup == "" {
		return "", fmt.Errorf("key backup not set: %w", apperrors.ErrNotFound)
	}
	return *user.EncryptedKeyBackup, nil
}

func (r *userRepository) List(limit, offset int, filters UserFilters) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if filters.IsPublic != nil {
		query = query.Where("is_public = ?", *filters.IsPublic)
	}

	if filters.Location != "" {
		query = query.Where("location ILIKE ?", "%"+filters.Location+"%")
	}

	if filters.SearchTerm != "" {
		query = query.Where("name ILIKE ?", "%"+filters.SearchTerm+"%")
	}

	// Geo radius (haversine) optionally OR'd with is_remote_ok=true.
	if filters.Lat != nil && filters.Lng != nil && filters.WithinKm != nil {
		haversine := "6371 * acos(LEAST(1.0, GREATEST(-1.0, " +
			"cos(radians(?)) * cos(radians(lat)) * cos(radians(lng) - radians(?)) " +
			"+ sin(radians(?)) * sin(radians(lat)))))"
		if filters.RemoteOK != nil && *filters.RemoteOK {
			query = query.Where(
				"(lat IS NOT NULL AND lng IS NOT NULL AND "+haversine+" <= ?) OR is_remote_ok = TRUE",
				*filters.Lat, *filters.Lng, *filters.Lat, *filters.WithinKm,
			)
		} else {
			query = query.Where(
				"lat IS NOT NULL AND lng IS NOT NULL AND "+haversine+" <= ?",
				*filters.Lat, *filters.Lng, *filters.Lat, *filters.WithinKm,
			)
		}
	} else if filters.RemoteOK != nil && *filters.RemoteOK {
		query = query.Where("is_remote_ok = TRUE")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.Preload("SkillsOffered.Skill").
		Preload("SkillsWanted.Skill").
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}
