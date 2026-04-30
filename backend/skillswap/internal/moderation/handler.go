package moderation

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	models "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/model"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *service.ModerationService
}

func NewHandler(svc *service.ModerationService) *Handler {
	return &Handler{svc: svc}
}

func userIDFromCtx(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, false
	}
	s, ok := v.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

type blockReq struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}

// Block: POST /api/v1/moderation/blocks
func (h *Handler) Block(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req blockReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	target, _ := uuid.Parse(req.UserID)
	if err := h.svc.Block(uid, target); err != nil {
		if errors.Is(err, apperrors.ErrSelfAction) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		resp.InternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User blocked"})
}

// Unblock: DELETE /api/v1/moderation/blocks/:user_id
func (h *Handler) Unblock(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	target, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if err := h.svc.Unblock(uid, target); err != nil {
		resp.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User unblocked"})
}

// ListBlocks: GET /api/v1/moderation/blocks
func (h *Handler) ListBlocks(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.svc.ListBlocks(uid)
	if err != nil {
		resp.InternalError(c, err)
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"blocked_id": r.BlockedID,
			"created_at": r.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"blocks": out})
}

type reportReq struct {
	TargetUserID string  `json:"target_user_id" binding:"required,uuid"`
	TargetKind   string  `json:"target_kind" binding:"required,oneof=user message swap"`
	TargetID     *string `json:"target_id"`
	Reason       string  `json:"reason" binding:"required,min=3,max=2000"`
}

// CreateReport: POST /api/v1/moderation/reports
func (h *Handler) CreateReport(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req reportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	targetUser, _ := uuid.Parse(req.TargetUserID)
	var tid *uuid.UUID
	if req.TargetID != nil && *req.TargetID != "" {
		parsed, err := uuid.Parse(*req.TargetID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target_id"})
			return
		}
		tid = &parsed
	}
	r, err := h.svc.CreateReport(uid, targetUser, models.ReportTargetKind(req.TargetKind), tid, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrSelfAction), errors.Is(err, apperrors.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"report_id": r.ReportID})
}

// ─── Admin endpoints ─────────────────────────────────────────────────────

// AdminListReports: GET /api/v1/admin/reports?status=open&limit=&offset=
func (h *Handler) AdminListReports(c *gin.Context) {
	status := models.ReportStatus(c.Query("status"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	rows, total, err := h.svc.ListReports(status, limit, offset)
	if err != nil {
		resp.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reports": rows,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

type resolveReq struct {
	Status string `json:"status" binding:"required,oneof=resolved dismissed reviewing"`
	Note   string `json:"note"`
}

// AdminResolveReport: POST /api/v1/admin/reports/:id/resolve
func (h *Handler) AdminResolveReport(c *gin.Context) {
	uid, ok := userIDFromCtx(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rid, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}
	var req resolveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ResolveReport(rid, uid, models.ReportStatus(req.Status), req.Note); err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		case errors.Is(err, apperrors.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			resp.InternalError(c, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Report updated"})
}
