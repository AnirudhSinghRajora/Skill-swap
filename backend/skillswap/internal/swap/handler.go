package swap

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	swapService appservice.SwapService
}

func NewHandler(swapService appservice.SwapService) *Handler {
	return &Handler{
		swapService: swapService,
	}
}

// Request and response structures
type CreateSwapRequestRequest struct {
	ResponderID    string `json:"responder_id" binding:"required,uuid"`
	OfferedSkillID string `json:"offered_skill_id" binding:"required,uuid"`
	WantedSkillID  string `json:"wanted_skill_id" binding:"required,uuid"`
}

type UpdateSwapStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=accepted rejected cancelled"`
}

type SwapRequestResponse struct {
	SwapID             string         `json:"swap_id"`
	RequesterID        string         `json:"requester_id"`
	ResponderID        string         `json:"responder_id"`
	OfferedSkillID     string         `json:"offered_skill_id"`
	WantedSkillID      string         `json:"wanted_skill_id"`
	Status             string         `json:"status"`
	RequesterCompleted bool           `json:"requester_completed"`
	ResponderCompleted bool           `json:"responder_completed"`
	CreatedAt          string         `json:"created_at"`
	UpdatedAt          string         `json:"updated_at"`
	Requester          *UserResponse  `json:"requester,omitempty"`
	Responder          *UserResponse  `json:"responder,omitempty"`
	OfferedSkill       *SkillResponse `json:"offered_skill,omitempty"`
	WantedSkill        *SkillResponse `json:"wanted_skill,omitempty"`
}

type UserResponse struct {
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email,omitempty"`
	Location string `json:"location,omitempty"`
	HasPhoto bool   `json:"has_photo"`
}

type SkillResponse struct {
	SkillID string `json:"skill_id"`
	Name    string `json:"name"`
}

type SwapRequestsResponse struct {
	Sent     []SwapRequestResponse `json:"sent"`
	Received []SwapRequestResponse `json:"received"`
}

type MatchResponse struct {
	User         UserResponse  `json:"user"`
	OfferedSkill SkillResponse `json:"offered_skill"`
	WantedSkill  SkillResponse `json:"wanted_skill"`
	MatchScore   int           `json:"match_score"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Helper function to convert models to responses
func (h *Handler) convertToSwapResponse(swap *models.SwapRequest, includeDetails bool) SwapRequestResponse {
	response := SwapRequestResponse{
		SwapID:             swap.SwapID.String(),
		RequesterID:        swap.RequesterID.String(),
		ResponderID:        swap.ResponderID.String(),
		OfferedSkillID:     swap.OfferedSkillID.String(),
		WantedSkillID:      swap.WantedSkillID.String(),
		Status:             string(swap.Status),
		RequesterCompleted: swap.RequesterCompleted,
		ResponderCompleted: swap.ResponderCompleted,
		CreatedAt:          swap.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:          swap.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if includeDetails {
		if swap.Requester.UserID != uuid.Nil {
			response.Requester = &UserResponse{
				UserID:   swap.Requester.UserID.String(),
				Name:     swap.Requester.Name,
				Location: h.getStringValue(swap.Requester.Location),
				HasPhoto: len(swap.Requester.PhotoData) > 0,
			}
		}

		if swap.Responder.UserID != uuid.Nil {
			response.Responder = &UserResponse{
				UserID:   swap.Responder.UserID.String(),
				Name:     swap.Responder.Name,
				Location: h.getStringValue(swap.Responder.Location),
				HasPhoto: len(swap.Responder.PhotoData) > 0,
			}
		}

		if swap.OfferedSkill.SkillID != uuid.Nil {
			response.OfferedSkill = &SkillResponse{
				SkillID: swap.OfferedSkill.SkillID.String(),
				Name:    swap.OfferedSkill.Name,
			}
		}

		if swap.WantedSkill.SkillID != uuid.Nil {
			response.WantedSkill = &SkillResponse{
				SkillID: swap.WantedSkill.SkillID.String(),
				Name:    swap.WantedSkill.Name,
			}
		}
	}

	return response
}

func (h *Handler) getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// CreateSwapRequest godoc
// @Summary Create swap request
// @Description Create a new skill swap request
// @Tags swaps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param swap body CreateSwapRequestRequest true "Swap request data"
// @Success 201 {object} SwapRequestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/swaps [post]
func (h *Handler) CreateSwapRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	var req CreateSwapRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Parse UUIDs
	responderID, err := uuid.Parse(req.ResponderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid responder ID"})
		return
	}

	offeredSkillID, err := uuid.Parse(req.OfferedSkillID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid offered skill ID"})
		return
	}

	wantedSkillID, err := uuid.Parse(req.WantedSkillID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid wanted skill ID"})
		return
	}

	// Create swap request
	swapDTO := &appservice.CreateSwapRequestDTO{
		RequesterID:    userID,
		ResponderID:    responderID,
		OfferedSkillID: offeredSkillID,
		WantedSkillID:  wantedSkillID,
	}

	swap, err := h.swapService.CreateSwapRequest(swapDTO)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrSelfAction), errors.Is(err, apperrors.ErrValidation):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrEmailNotVerified):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error(), "code": "email_not_verified"})
		case errors.Is(err, apperrors.ErrDuplicate):
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	response := h.convertToSwapResponse(swap, true)
	c.JSON(http.StatusCreated, response)
}

// GetSwapRequest godoc
// @Summary Get swap request
// @Description Get a specific swap request by ID
// @Tags swaps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Swap ID"
// @Success 200 {object} SwapRequestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/swaps/{id} [get]
func (h *Handler) GetSwapRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	swapIDStr := c.Param("id")
	swapID, err := uuid.Parse(swapIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid swap ID"})
		return
	}

	swap, err := h.swapService.GetSwapRequestByID(swapID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Swap request not found"})
			return
		}
		resp.InternalError(c, err)
		return
	}

	// Check if user is involved in the swap
	if swap.RequesterID != userID && swap.ResponderID != userID {
		c.JSON(http.StatusForbidden, ErrorResponse{Error: "Access denied"})
		return
	}

	response := h.convertToSwapResponse(swap, true)
	c.JSON(http.StatusOK, response)
}

// GetUserSwapRequests godoc
// @Summary Get user's swap requests
// @Description Get all swap requests for the authenticated user
// @Tags swaps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(pending, accepted, rejected, cancelled)
// @Param sent query bool false "Include sent requests"
// @Param received query bool false "Include received requests"
// @Param limit query int false "Limit number of results"
// @Param offset query int false "Offset for pagination"
// @Success 200 {object} SwapRequestsResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/swaps [get]
func (h *Handler) GetUserSwapRequests(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	// Parse query parameters
	filter := appservice.SwapRequestFilter{}

	if statusStr := c.Query("status"); statusStr != "" {
		status := models.SwapStatus(statusStr)
		switch status {
		case models.StatusPending, models.StatusAccepted, models.StatusRejected, models.StatusCancelled:
			filter.Status = &status
		default:
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid status. Must be one of: pending, accepted, rejected, cancelled"})
			return
		}
	}

	if sentStr := c.Query("sent"); sentStr != "" {
		filter.Sent = sentStr == "true"
	}

	if receivedStr := c.Query("received"); receivedStr != "" {
		filter.Received = receivedStr == "true"
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			if limit < 0 {
				limit = 0
			} else if limit > 100 {
				limit = 100
			}
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			if offset < 0 {
				offset = 0
			}
			filter.Offset = offset
		}
	}

	// If neither sent nor received specified, show organized view
	if !filter.Sent && !filter.Received {
		swapRequests, err := h.swapService.GetSwapRequestsForUser(userID)
		if err != nil {
			resp.InternalError(c, err)
			return
		}

		response := SwapRequestsResponse{
			Sent:     make([]SwapRequestResponse, len(swapRequests.Sent)),
			Received: make([]SwapRequestResponse, len(swapRequests.Received)),
		}

		for i, swap := range swapRequests.Sent {
			response.Sent[i] = h.convertToSwapResponse(&swap, true)
		}

		for i, swap := range swapRequests.Received {
			response.Received[i] = h.convertToSwapResponse(&swap, true)
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// Get filtered requests
	swaps, err := h.swapService.GetUserSwapRequests(userID, filter)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	var response []SwapRequestResponse
	for _, swap := range swaps {
		response = append(response, h.convertToSwapResponse(&swap, true))
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSwapStatus godoc
// @Summary Update swap status
// @Description Accept, reject, or cancel a swap request
// @Tags swaps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Swap ID"
// @Param status body UpdateSwapStatusRequest true "New status"
// @Success 200 {object} SwapRequestResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/swaps/{id}/status [put]
func (h *Handler) UpdateSwapStatus(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	swapIDStr := c.Param("id")
	swapID, err := uuid.Parse(swapIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid swap ID"})
		return
	}

	var req UpdateSwapStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	status := models.SwapStatus(req.Status)
	swap, err := h.swapService.UpdateSwapStatus(swapID, userID, status)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Swap request not found"})
		case errors.Is(err, apperrors.ErrForbidden):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrWrongStatus):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	response := h.convertToSwapResponse(swap, true)
	c.JSON(http.StatusOK, response)
}

// DeleteSwapRequest godoc
// @Summary Delete swap request
// @Description Delete a swap request (requester only)
// @Tags swaps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Swap ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/swaps/{id} [delete]
func (h *Handler) DeleteSwapRequest(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	swapIDStr := c.Param("id")
	swapID, err := uuid.Parse(swapIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid swap ID"})
		return
	}

	err = h.swapService.DeleteSwapRequest(swapID, userID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Swap request not found"})
		case errors.Is(err, apperrors.ErrForbidden):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrWrongStatus):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetPotentialMatches godoc
// @Summary Get potential matches
// @Description Find potential swap matches for the authenticated user
// @Tags swaps
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} MatchResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/swaps/matches [get]
func (h *Handler) GetPotentialMatches(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid user ID"})
		return
	}

	matches, err := h.swapService.FindPotentialMatches(userID)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	var response []MatchResponse
	for _, match := range matches {
		response = append(response, MatchResponse{
			User: UserResponse{
				UserID:   match.User.UserID.String(),
				Name:     match.User.Name,
				Location: h.getStringValue(match.User.Location),
				HasPhoto: len(match.User.PhotoData) > 0,
			},
			OfferedSkill: SkillResponse{
				SkillID: match.OfferedSkill.SkillID.String(),
				Name:    match.OfferedSkill.Name,
			},
			WantedSkill: SkillResponse{
				SkillID: match.WantedSkill.SkillID.String(),
				Name:    match.WantedSkill.Name,
			},
			MatchScore: match.MatchScore,
		})
	}

	c.JSON(http.StatusOK, response)
}
