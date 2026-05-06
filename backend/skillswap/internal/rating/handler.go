package rating

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Response DTOs with proper JSON tags
type RatingResponse struct {
	RatingID  string        `json:"rating_id"`
	SwapID    string        `json:"swap_id"`
	RaterID   string        `json:"rater_id"`
	RateeID   string        `json:"ratee_id"`
	Score     int16         `json:"score"`
	Comment   *string       `json:"comment"`
	CreatedAt string        `json:"created_at"`
	Rater     *UserResponse `json:"rater,omitempty"`
	Ratee     *UserResponse `json:"ratee,omitempty"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	HasPhoto bool   `json:"has_photo"`
}

func convertRating(r *models.SwapRating) RatingResponse {
	resp := RatingResponse{
		RatingID:  r.RatingID.String(),
		SwapID:    r.SwapID.String(),
		RaterID:   r.RaterID.String(),
		RateeID:   r.RateeID.String(),
		Score:     r.Score,
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if r.Rater.UserID != uuid.Nil {
		resp.Rater = &UserResponse{
			UserID:   r.Rater.UserID.String(),
			Name:     r.Rater.Name,
			HasPhoto: len(r.Rater.PhotoData) > 0,
		}
	}
	if r.Ratee.UserID != uuid.Nil {
		resp.Ratee = &UserResponse{
			UserID:   r.Ratee.UserID.String(),
			Name:     r.Ratee.Name,
			HasPhoto: len(r.Ratee.PhotoData) > 0,
		}
	}
	return resp
}

func convertRatings(ratings []models.SwapRating) []RatingResponse {
	result := make([]RatingResponse, len(ratings))
	for i := range ratings {
		result[i] = convertRating(&ratings[i])
	}
	return result
}

type Handler struct {
	ratingService service.RatingService
}

func NewHandler(ratingService service.RatingService) *Handler {
	return &Handler{
		ratingService: ratingService,
	}
}

// CreateRating creates a new rating for a swap
// @Summary Create a new rating
// @Description Create a rating for a completed swap
// @Tags ratings
// @Accept json
// @Produce json
// @Param rating body service.CreateRatingDTO true "Rating data"
// @Success 201 {object} models.SwapRating
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 409 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/ratings [post]
func (h *Handler) CreateRating(c *gin.Context) {
	var req service.CreateRatingDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get rater ID from JWT token
	raterID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid, err := uuid.Parse(raterID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	req.RaterID = uid

	// Check if user can rate this swap
	canRate, err := h.ratingService.CanUserRateSwap(req.SwapID, req.RaterID)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if !canRate {
		c.JSON(http.StatusForbidden, gin.H{"error": "You cannot rate this swap"})
		return
	}

	// Check if user has already rated this swap
	hasRated, err := h.ratingService.HasUserRatedSwap(req.SwapID, req.RaterID)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if hasRated {
		c.JSON(http.StatusConflict, gin.H{"error": "You have already rated this swap"})
		return
	}

	rating, err := h.ratingService.CreateRating(&req)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusCreated, convertRating(rating))
}

// GetRating retrieves a rating by ID
// @Summary Get a rating by ID
// @Description Get details of a specific rating
// @Tags ratings
// @Accept json
// @Produce json
// @Param id path string true "Rating ID"
// @Success 200 {object} models.SwapRating
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/ratings/{id} [get]
func (h *Handler) GetRating(c *gin.Context) {
	ratingIDStr := c.Param("id")
	ratingID, err := uuid.Parse(ratingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
		return
	}

	rating, err := h.ratingService.GetRatingByID(ratingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
		return
	}

	c.JSON(http.StatusOK, convertRating(rating))
}

// UpdateRating updates an existing rating
// @Summary Update a rating
// @Description Update an existing rating (only by the rater)
// @Tags ratings
// @Accept json
// @Produce json
// @Param id path string true "Rating ID"
// @Param rating body service.UpdateRatingDTO true "Updated rating data"
// @Success 200 {object} models.SwapRating
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/ratings/{id} [put]
func (h *Handler) UpdateRating(c *gin.Context) {
	ratingIDStr := c.Param("id")
	ratingID, err := uuid.Parse(ratingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
		return
	}

	var req service.UpdateRatingDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	rating, err := h.ratingService.UpdateRating(ratingID, uid, &req)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
		case errors.Is(err, apperrors.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own ratings"})
		default:
			response.InternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, convertRating(rating))
}

// DeleteRating deletes a rating
// @Summary Delete a rating
// @Description Delete a rating (only by the rater)
// @Tags ratings
// @Accept json
// @Produce json
// @Param id path string true "Rating ID"
// @Success 204
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/ratings/{id} [delete]
func (h *Handler) DeleteRating(c *gin.Context) {
	ratingIDStr := c.Param("id")
	ratingID, err := uuid.Parse(ratingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating ID"})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	uid, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.ratingService.DeleteRating(ratingID, uid)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Rating not found"})
		case errors.Is(err, apperrors.ErrForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own ratings"})
		default:
			response.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetSwapRatings retrieves all ratings for a specific swap
// @Summary Get ratings for a swap
// @Description Get all ratings for a specific swap
// @Tags ratings
// @Accept json
// @Produce json
// @Param swap_id path string true "Swap ID"
// @Success 200 {array} models.SwapRating
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/ratings/swap/{swap_id} [get]
func (h *Handler) GetSwapRatings(c *gin.Context) {
	swapIDStr := c.Param("swap_id")
	swapID, err := uuid.Parse(swapIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid swap ID"})
		return
	}

	ratings, err := h.ratingService.GetSwapRatings(swapID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ratings": convertRatings(ratings)})
}

// GetUserRatings retrieves ratings for a user with filters
// @Summary Get user ratings
// @Description Get ratings for a user (given or received)
// @Tags ratings
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param as_rater query bool false "Get ratings given by user"
// @Param as_ratee query bool false "Get ratings received by user"
// @Param min_score query int false "Minimum score filter"
// @Param max_score query int false "Maximum score filter"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {array} models.SwapRating
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/users/{id}/ratings [get]
func (h *Handler) GetUserRatings(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse query parameters for filtering
	filter := service.RatingFilter{}

	if asRater := c.Query("as_rater"); asRater != "" {
		if val, err := strconv.ParseBool(asRater); err == nil {
			filter.AsRater = val
		}
	}

	if asRatee := c.Query("as_ratee"); asRatee != "" {
		if val, err := strconv.ParseBool(asRatee); err == nil {
			filter.AsRatee = val
		}
	}

	if minScore := c.Query("min_score"); minScore != "" {
		if val, err := strconv.Atoi(minScore); err == nil {
			if val < 1 {
				val = 1
			} else if val > 5 {
				val = 5
			}
			filter.MinScore = &val
		}
	}

	if maxScore := c.Query("max_score"); maxScore != "" {
		if val, err := strconv.Atoi(maxScore); err == nil {
			if val < 1 {
				val = 1
			} else if val > 5 {
				val = 5
			}
			filter.MaxScore = &val
		}
	}

	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			if val < 0 {
				val = 0
			} else if val > 100 {
				val = 100
			}
			filter.Limit = val
		}
	} else {
		filter.Limit = 20 // Default limit
	}

	if offset := c.Query("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil {
			if val < 0 {
				val = 0
			}
			filter.Offset = val
		}
	}

	// If neither as_rater nor as_ratee is specified, return both
	if !filter.AsRater && !filter.AsRatee {
		filter.AsRater = true
		filter.AsRatee = true
	}

	ratings, err := h.ratingService.GetUserRatings(userID, filter)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ratings": convertRatings(ratings)})
}

// GetUserRatingStats retrieves rating statistics for a user
// @Summary Get user rating statistics
// @Description Get rating statistics and summary for a user
// @Tags ratings
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} service.UserRatingStats
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /api/v1/users/{id}/ratings/stats [get]
func (h *Handler) GetUserRatingStats(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	stats, err := h.ratingService.GetUserRatingStats(userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}
