package cohort

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc appservice.CohortService
}

func NewHandler(svc appservice.CohortService) *Handler { return &Handler{svc: svc} }

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
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrConflict), errors.Is(err, apperrors.ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrRateLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error(), "code": "rate_limited"})
	case errors.Is(err, apperrors.ErrValidation),
		errors.Is(err, apperrors.ErrWrongStatus),
		errors.Is(err, apperrors.ErrSelfAction):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		resp.InternalError(c, err)
	}
}

type createReq struct {
	SkillID         uuid.UUID `json:"skill_id" binding:"required"`
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	Capacity        int16     `json:"capacity"`
	Kind            string    `json:"kind"`
	IsPublic        *bool     `json:"is_public"`
	SchedulePattern *string   `json:"schedule_pattern"`
}

// POST /api/v1/cohorts
func (h *Handler) Create(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pub := true
	if req.IsPublic != nil {
		pub = *req.IsPublic
	}
	in := appservice.CreateCohortInput{
		SkillID:         req.SkillID,
		Title:           req.Title,
		Description:     req.Description,
		Capacity:        req.Capacity,
		Kind:            models.CohortKind(req.Kind),
		IsPublic:        pub,
		SchedulePattern: req.SchedulePattern,
	}
	out, err := h.svc.Create(user, in)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}

// GET /api/v1/cohorts
func (h *Handler) List(c *gin.Context) {
	f := appservice.ListCohortFilter{}
	if v := c.Query("skill_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.SkillID = &id
		}
	}
	if v := c.Query("kind"); v != "" {
		k := models.CohortKind(v)
		f.Kind = &k
	}
	if v := c.Query("status"); v != "" {
		st := models.CohortStatus(v)
		f.Status = &st
	}
	if c.Query("has_seats") == "true" {
		f.HasSeats = true
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Offset = n
		}
	}
	out, err := h.svc.List(f)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohorts": out})
}

// GET /api/v1/cohorts/:id
func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cohort id"})
		return
	}
	co, members, err := h.svc.Get(id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohort": co, "members": members})
}

// POST /api/v1/cohorts/:id/join
func (h *Handler) Join(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cohort id"})
		return
	}
	m, err := h.svc.Join(id, user)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

// POST /api/v1/cohorts/:id/leave
func (h *Handler) Leave(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cohort id"})
		return
	}
	if err := h.svc.Leave(id, user); err != nil {
		h.writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type addSessionReq struct {
	ScheduledStart time.Time `json:"scheduled_start" binding:"required"`
	ScheduledEnd   time.Time `json:"scheduled_end" binding:"required"`
}

// POST /api/v1/cohorts/:id/sessions
func (h *Handler) AddSession(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cohort id"})
		return
	}
	var req addSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cs, err := h.svc.AddSession(id, user, req.ScheduledStart, req.ScheduledEnd)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, cs)
}

// GET /api/v1/cohorts/:id/sessions
func (h *Handler) ListSessions(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cohort id"})
		return
	}
	out, err := h.svc.ListSessions(id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": out})
}

// DELETE /api/v1/cohorts/:id/members/:user_id  (host only)
func (h *Handler) RemoveMember(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	cohortID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cohort id"})
		return
	}
	memberID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.svc.RemoveMember(cohortID, user, memberID); err != nil {
		h.writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
