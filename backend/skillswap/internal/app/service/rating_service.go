package service

import (
	"errors"
	"fmt"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RatingService interface {
	// Rating CRUD operations
	CreateRating(req *CreateRatingDTO) (*models.SwapRating, error)
	GetRatingByID(ratingID uuid.UUID) (*models.SwapRating, error)
	UpdateRating(ratingID uuid.UUID, userID uuid.UUID, req *UpdateRatingDTO) (*models.SwapRating, error)
	DeleteRating(ratingID uuid.UUID, userID uuid.UUID) error

	// Rating queries
	GetSwapRatings(swapID uuid.UUID) ([]models.SwapRating, error)
	GetUserRatings(userID uuid.UUID, filter RatingFilter) ([]models.SwapRating, error)
	GetUserRatingStats(userID uuid.UUID) (*UserRatingStats, error)

	// Rating checks
	CanUserRateSwap(swapID uuid.UUID, raterID uuid.UUID) (bool, error)
	HasUserRatedSwap(swapID uuid.UUID, raterID uuid.UUID) (bool, error)
}

// DTOs and Request structures
type CreateRatingDTO struct {
	SwapID  uuid.UUID `json:"swap_id" binding:"required"`
	RaterID uuid.UUID `json:"rater_id"`
	RateeID uuid.UUID `json:"ratee_id" binding:"required"`
	Score   int16     `json:"score" binding:"required,min=1,max=5"`
	Comment *string   `json:"comment,omitempty" binding:"omitempty,max=1000"`
}

type UpdateRatingDTO struct {
	Score   int16   `json:"score" binding:"required,min=1,max=5"`
	Comment *string `json:"comment,omitempty" binding:"omitempty,max=1000"`
}

type RatingFilter struct {
	AsRater  bool `json:"as_rater,omitempty"` // Ratings given by user
	AsRatee  bool `json:"as_ratee,omitempty"` // Ratings received by user
	MinScore *int `json:"min_score,omitempty"`
	MaxScore *int `json:"max_score,omitempty"`
	Limit    int  `json:"limit,omitempty"`
	Offset   int  `json:"offset,omitempty"`
}

type UserRatingStats struct {
	UserID         uuid.UUID `json:"user_id"`
	TotalRatings   int       `json:"total_ratings"`
	AverageRating  float64   `json:"average_rating"`
	FiveStarCount  int       `json:"five_star_count"`
	FourStarCount  int       `json:"four_star_count"`
	ThreeStarCount int       `json:"three_star_count"`
	TwoStarCount   int       `json:"two_star_count"`
	OneStarCount   int       `json:"one_star_count"`
}

type ratingService struct {
	db *gorm.DB
}

func NewRatingService(db *gorm.DB) RatingService {
	return &ratingService{db: db}
}

// CreateRating creates a new rating for a completed swap
func (r *ratingService) CreateRating(req *CreateRatingDTO) (*models.SwapRating, error) {
	var rating *models.SwapRating
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Check if the swap exists and is completed (accepted)
		var swap models.SwapRequest
		if err := tx.First(&swap, "swap_id = ?", req.SwapID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("swap request not found: %w", apperrors.ErrNotFound)
			}
			return err
		}

		// Only allow rating for accepted swaps
		if swap.Status != models.StatusAccepted {
			return fmt.Errorf("can only rate completed (accepted) swaps: %w", apperrors.ErrWrongStatus)
		}

		// Verify the rater is part of the swap
		if swap.RequesterID != req.RaterID && swap.ResponderID != req.RaterID {
			return fmt.Errorf("only participants can rate a swap: %w", apperrors.ErrNotParticipant)
		}

		// Determine the ratee (the other participant)
		if swap.RequesterID == req.RaterID {
			req.RateeID = swap.ResponderID
		} else {
			req.RateeID = swap.RequesterID
		}

		// Check if user has already rated this swap
		var existingCount int64
		if err := tx.Model(&models.SwapRating{}).
			Where("swap_id = ? AND rater_id = ?", req.SwapID, req.RaterID).
			Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return fmt.Errorf("you have already rated this swap: %w", apperrors.ErrConflict)
		}

		rating = &models.SwapRating{
			SwapID:  req.SwapID,
			RaterID: req.RaterID,
			RateeID: req.RateeID,
			Score:   req.Score,
			Comment: req.Comment,
		}

		if err := tx.Create(rating).Error; err != nil {
			return err
		}

		// Load relationships
		return tx.Preload("Swap").Preload("Rater").Preload("Ratee").
			First(rating, rating.RatingID).Error
	})
	if err != nil {
		return nil, err
	}

	return rating, nil
}

// GetRatingByID retrieves a rating by its ID
func (r *ratingService) GetRatingByID(ratingID uuid.UUID) (*models.SwapRating, error) {
	var rating models.SwapRating
	err := r.db.Preload("Swap").Preload("Rater").Preload("Ratee").
		First(&rating, "rating_id = ?", ratingID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("rating not found: %w", apperrors.ErrNotFound)
		}
		return nil, err
	}

	return &rating, nil
}

// UpdateRating updates an existing rating
func (r *ratingService) UpdateRating(ratingID uuid.UUID, userID uuid.UUID, req *UpdateRatingDTO) (*models.SwapRating, error) {
	rating, err := r.GetRatingByID(ratingID)
	if err != nil {
		return nil, err
	}

	// Only the rater can update their rating
	if rating.RaterID != userID {
		return nil, fmt.Errorf("only the rater can update this rating: %w", apperrors.ErrForbidden)
	}

	// Update fields
	rating.Score = req.Score
	rating.Comment = req.Comment

	err = r.db.Save(rating).Error
	if err != nil {
		return nil, err
	}

	return rating, nil
}

// DeleteRating deletes a rating
func (r *ratingService) DeleteRating(ratingID uuid.UUID, userID uuid.UUID) error {
	rating, err := r.GetRatingByID(ratingID)
	if err != nil {
		return err
	}

	// Only the rater can delete their rating
	if rating.RaterID != userID {
		return fmt.Errorf("only the rater can delete this rating: %w", apperrors.ErrForbidden)
	}

	return r.db.Delete(&models.SwapRating{}, "rating_id = ?", ratingID).Error
}

// GetSwapRatings retrieves all ratings for a specific swap
func (r *ratingService) GetSwapRatings(swapID uuid.UUID) ([]models.SwapRating, error) {
	var ratings []models.SwapRating
	err := r.db.Preload("Rater").Preload("Ratee").
		Where("swap_id = ?", swapID).
		Order("created_at DESC").
		Find(&ratings).Error

	return ratings, err
}

// GetUserRatings retrieves ratings for a user with filtering
func (r *ratingService) GetUserRatings(userID uuid.UUID, filter RatingFilter) ([]models.SwapRating, error) {
	query := r.db.Model(&models.SwapRating{}).
		Preload("Swap").Preload("Rater").Preload("Ratee")

	// Apply filters
	if filter.AsRater && filter.AsRatee {
		query = query.Where("rater_id = ? OR ratee_id = ?", userID, userID)
	} else if filter.AsRater {
		query = query.Where("rater_id = ?", userID)
	} else if filter.AsRatee {
		query = query.Where("ratee_id = ?", userID)
	} else {
		// Default: show ratings received by user
		query = query.Where("ratee_id = ?", userID)
	}

	// Score filters
	if filter.MinScore != nil {
		query = query.Where("score >= ?", *filter.MinScore)
	}
	if filter.MaxScore != nil {
		query = query.Where("score <= ?", *filter.MaxScore)
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

	var ratings []models.SwapRating
	err := query.Find(&ratings).Error

	return ratings, err
}

// GetUserRatingStats calculates rating statistics for a user
func (r *ratingService) GetUserRatingStats(userID uuid.UUID) (*UserRatingStats, error) {
	stats := &UserRatingStats{
		UserID: userID,
	}

	// Get total count and average
	var totalRatings int64
	var totalScore int64

	err := r.db.Model(&models.SwapRating{}).
		Where("ratee_id = ?", userID).
		Count(&totalRatings).Error
	if err != nil {
		return nil, err
	}

	if totalRatings == 0 {
		return stats, nil
	}

	err = r.db.Model(&models.SwapRating{}).
		Select("SUM(score)").
		Where("ratee_id = ?", userID).
		Scan(&totalScore).Error
	if err != nil {
		return nil, err
	}

	stats.TotalRatings = int(totalRatings)
	stats.AverageRating = float64(totalScore) / float64(totalRatings)

	// Get count by score
	scores := []struct {
		Score int16
		Count int64
	}{}

	err = r.db.Model(&models.SwapRating{}).
		Select("score, COUNT(*) as count").
		Where("ratee_id = ?", userID).
		Group("score").
		Find(&scores).Error
	if err != nil {
		return nil, err
	}

	// Map to stats
	for _, score := range scores {
		switch score.Score {
		case 5:
			stats.FiveStarCount = int(score.Count)
		case 4:
			stats.FourStarCount = int(score.Count)
		case 3:
			stats.ThreeStarCount = int(score.Count)
		case 2:
			stats.TwoStarCount = int(score.Count)
		case 1:
			stats.OneStarCount = int(score.Count)
		}
	}

	return stats, nil
}

// CanUserRateSwap checks if a user can rate a specific swap
func (r *ratingService) CanUserRateSwap(swapID uuid.UUID, raterID uuid.UUID) (bool, error) {
	// Check if swap exists and is accepted
	var swap models.SwapRequest
	err := r.db.First(&swap, "swap_id = ? AND status = ?", swapID, models.StatusAccepted).Error
	if err != nil {
		return false, nil // Swap not found or not accepted
	}

	// Check if user is part of the swap
	if swap.RequesterID != raterID && swap.ResponderID != raterID {
		return false, nil
	}

	// Check if user has already rated
	hasRated, err := r.HasUserRatedSwap(swapID, raterID)
	if err != nil {
		return false, err
	}

	return !hasRated, nil
}

// HasUserRatedSwap checks if a user has already rated a specific swap
func (r *ratingService) HasUserRatedSwap(swapID uuid.UUID, raterID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.SwapRating{}).
		Where("swap_id = ? AND rater_id = ?", swapID, raterID).
		Count(&count).Error

	return count > 0, err
}
