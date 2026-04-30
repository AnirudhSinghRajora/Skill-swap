package swap

import (
	"errors"
	"net/http"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type noShowReq struct {
	Reason string `json:"reason" binding:"max=2000"`
}

// ReportNoShow: POST /api/v1/swaps/:id/no-show
func (h *Handler) ReportNoShow(c *gin.Context) {
	uidStr, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}
	swapID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid swap id"})
		return
	}
	var req noShowReq
	_ = c.ShouldBindJSON(&req)
	swap, err := h.swapService.ReportNoShow(swapID, uid, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrNotParticipant):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrWrongStatus), errors.Is(err, apperrors.ErrValidation):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrConflict):
			c.JSON(http.StatusConflict, ErrorResponse{Error: err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}
	c.JSON(http.StatusOK, h.convertToSwapResponse(swap, false))
}

type disputeReq struct {
	Reason string `json:"reason" binding:"required,min=3,max=2000"`
}

// RaiseDispute: POST /api/v1/swaps/:id/dispute
func (h *Handler) RaiseDispute(c *gin.Context) {
	uidStr, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
		return
	}
	uid, err := uuid.Parse(uidStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}
	swapID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid swap id"})
		return
	}
	var req disputeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	swap, err := h.swapService.RaiseDispute(swapID, uid, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrNotParticipant):
			c.JSON(http.StatusForbidden, ErrorResponse{Error: err.Error()})
		case errors.Is(err, apperrors.ErrValidation):
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}
	c.JSON(http.StatusOK, h.convertToSwapResponse(swap, false))
}

// GetReliability: GET /api/v1/users/:id/reliability
// Returns reliability stats; only exposes the score when qualifies==true.
func (h *Handler) GetReliability(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user id"})
		return
	}
	stats, err := h.swapService.GetReliabilityStats(targetID)
	if err != nil {
		resp.InternalError(c, err)
		return
	}
	if !stats.Qualifies {
		c.JSON(http.StatusOK, gin.H{
			"qualifies":         false,
			"successful_swaps":  stats.SuccessfulSwaps,
			"reported_no_shows": stats.ReportedNoShows,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"qualifies":         true,
		"successful_swaps":  stats.SuccessfulSwaps,
		"reported_no_shows": stats.ReportedNoShows,
		"score":             stats.Score,
	})
}
