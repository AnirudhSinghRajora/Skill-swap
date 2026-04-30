package session

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
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
	case errors.Is(err, apperrors.ErrConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "code": "conflict"})
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

// GET /api/v1/sessions/:id/calendar.ics
func (h *Handler) ICS(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	sess, err := h.svc.Get(id, user)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	body := buildICS(sess)
	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"skillswap-session-%s.ics\"", sess.SessionID.String()))
	c.String(http.StatusOK, body)
}

type notesPayload struct {
	SessionID  uuid.UUID `json:"session_id"`
	PrepNotes  string    `json:"prep_notes"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func notesPayloadOf(s *models.Session) notesPayload {
	body := ""
	if s.PrepNotes != nil {
		body = *s.PrepNotes
	}
	return notesPayload{SessionID: s.SessionID, PrepNotes: body, UpdatedAt: s.UpdatedAt}
}

// GET /api/v1/sessions/:id/notes
func (h *Handler) GetNotes(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	sess, err := h.svc.Get(id, user)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.Header("ETag", sess.UpdatedAt.UTC().Format(time.RFC3339Nano))
	c.JSON(http.StatusOK, notesPayloadOf(sess))
}

type updateNotesRequest struct {
	PrepNotes string `json:"prep_notes"`
}

// PUT /api/v1/sessions/:id/notes
//
// Optimistic concurrency: If-Match header carries the previous updated_at
// value (RFC 3339, as returned by GET). When supplied and stale, returns
// 409 with code=conflict so the client can refresh.
func (h *Handler) UpdateNotes(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	var req updateNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.PrepNotes) > 16384 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "prep_notes exceeds 16KB"})
		return
	}
	var ifMatch *time.Time
	if hdr := strings.TrimSpace(c.GetHeader("If-Match")); hdr != "" {
		t, perr := time.Parse(time.RFC3339Nano, hdr)
		if perr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "If-Match must be RFC3339Nano"})
			return
		}
		ifMatch = &t
	}
	sess, err := h.svc.UpdateNotes(id, user, req.PrepNotes, ifMatch)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.Header("ETag", sess.UpdatedAt.UTC().Format(time.RFC3339Nano))
	c.JSON(http.StatusOK, notesPayloadOf(sess))
}

// icsEscape escapes commas, semicolons, backslashes and newlines per RFC 5545.
func icsEscape(s string) string {
	r := strings.NewReplacer("\\", "\\\\", ";", "\\;", ",", "\\,", "\n", "\\n", "\r", "")
	return r.Replace(s)
}

func buildICS(sess *models.Session) string {
	utc := func(t time.Time) string { return t.UTC().Format("20060102T150405Z") }
	now := utc(time.Now())
	uid := fmt.Sprintf("session-%s@skillswap", sess.SessionID.String())
	summary := icsEscape(fmt.Sprintf("SkillSwap session (%s)", sess.Status))
	desc := icsEscape(fmt.Sprintf("Swap %s - open in SkillSwap to join.", sess.SwapID.String()))
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//SkillSwap//Sessions//EN",
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"UID:" + uid,
		"DTSTAMP:" + now,
		"DTSTART:" + utc(sess.ScheduledStart),
		"DTEND:" + utc(sess.ScheduledEnd),
		"SUMMARY:" + summary,
		"DESCRIPTION:" + desc,
		"STATUS:" + strings.ToUpper(string(sess.Status)),
		"END:VEVENT",
		"END:VCALENDAR",
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}
