package session

import (
	"errors"
	"net/http"
	"time"

	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc appservice.SessionService
}

func NewHandler(svc appservice.SessionService) *Handler {
	return &Handler{svc: svc}
}

type proposeRequest struct {
	ScheduledStart time.Time `json:"scheduled_start" binding:"required"`
	ScheduledEnd   time.Time `json:"scheduled_end" binding:"required"`
}

func parseUser(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return uuid.Nil, false
	}
	u, err := uuid.Parse(v.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return uuid.Nil, false
	}
	return u, true
}

func (h *Handler) writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrNotParticipant), errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrValidation),
		errors.Is(err, apperrors.ErrWrongStatus),
		errors.Is(err, apperrors.ErrSelfAction):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		resp.InternalError(c, err)
	}
}

// POST /api/v1/swaps/:id/sessions
func (h *Handler) Propose(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	swapID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid swap id"})
		return
	}
	var req proposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sess, err := h.svc.Propose(swapID, user, req.ScheduledStart, req.ScheduledEnd)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, sess)
}

// GET /api/v1/swaps/:id/sessions
func (h *Handler) List(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	swapID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid swap id"})
		return
	}
	out, err := h.svc.ListForSwap(swapID, user)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": out})
}

func (h *Handler) action(c *gin.Context, fn func(uuid.UUID, uuid.UUID) (interface{}, error)) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	sess, err := fn(id, user)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, sess)
}

// PUT /api/v1/sessions/:id/accept
func (h *Handler) Accept(c *gin.Context) {
	h.action(c, func(id, u uuid.UUID) (interface{}, error) { return h.svc.Accept(id, u) })
}

// PUT /api/v1/sessions/:id/cancel
func (h *Handler) Cancel(c *gin.Context) {
	h.action(c, func(id, u uuid.UUID) (interface{}, error) { return h.svc.Cancel(id, u) })
}

// PUT /api/v1/sessions/:id/complete
func (h *Handler) Complete(c *gin.Context) {
	h.action(c, func(id, u uuid.UUID) (interface{}, error) { return h.svc.Complete(id, u) })
}
