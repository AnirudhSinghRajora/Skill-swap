package qa

import (
	"errors"
	"net/http"
	"strconv"

	appservice "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc appservice.QAService
}

func NewHandler(svc appservice.QAService) *Handler { return &Handler{svc: svc} }

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
	case errors.Is(err, apperrors.ErrValidation),
		errors.Is(err, apperrors.ErrWrongStatus),
		errors.Is(err, apperrors.ErrSelfAction):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		resp.InternalError(c, err)
	}
}

type askReq struct {
	SkillID *uuid.UUID `json:"skill_id"`
	Title   string     `json:"title" binding:"required"`
	Body    string     `json:"body"  binding:"required"`
}

// POST /api/v1/questions
func (h *Handler) Ask(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	var r askReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q, err := h.svc.AskQuestion(user, appservice.AskQuestionInput{
		SkillID: r.SkillID,
		Title:   r.Title,
		Body:    r.Body,
	})
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, q)
}

// GET /api/v1/questions/:id
func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	q, answers, err := h.svc.GetQuestion(id)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"question": q, "answers": answers})
}

// GET /api/v1/questions
func (h *Handler) List(c *gin.Context) {
	f := appservice.ListQuestionsFilter{Q: c.Query("q")}
	if v := c.Query("skill_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.SkillID = &id
		}
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
	out, err := h.svc.ListQuestions(f)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

type editQReq struct {
	Title string `json:"title" binding:"required"`
	Body  string `json:"body"  binding:"required"`
}

// PUT /api/v1/questions/:id
func (h *Handler) Update(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var r editQReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	q, err := h.svc.UpdateQuestion(id, user, r.Title, r.Body)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, q)
}

// DELETE /api/v1/questions/:id
func (h *Handler) Delete(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteQuestion(id, user); err != nil {
		h.writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type answerReq struct {
	Body string `json:"body" binding:"required"`
}

// POST /api/v1/questions/:id/answers
func (h *Handler) Answer(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	qid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var r answerReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := h.svc.PostAnswer(qid, user, r.Body)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

// PUT /api/v1/answers/:id
func (h *Handler) UpdateAnswer(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var r answerReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := h.svc.UpdateAnswer(id, user, r.Body)
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

// DELETE /api/v1/answers/:id
func (h *Handler) DeleteAnswer(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteAnswer(id, user); err != nil {
		h.writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// POST /api/v1/answers/:id/accept
func (h *Handler) Accept(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.AcceptAnswer(id, user); err != nil {
		h.writeErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type voteReq struct {
	TargetKind string    `json:"target_kind" binding:"required"`
	TargetID   uuid.UUID `json:"target_id"   binding:"required"`
}

// POST /api/v1/votes — toggle a vote on a question or answer.
func (h *Handler) Vote(c *gin.Context) {
	user, ok := parseUser(c)
	if !ok {
		return
	}
	var r voteReq
	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	added, count, err := h.svc.ToggleVote(user, r.TargetID, models.VoteTargetKind(r.TargetKind))
	if err != nil {
		h.writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"voted": added, "upvote_count": count})
}
