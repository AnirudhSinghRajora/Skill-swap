package skill

import (
	"errors"
	"net/http"

	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/apperrors"
	"github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/app/service"
	resp "github.com/Sky-walkerX/Skill-swap/backend/skillswap/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	skillService service.SkillService
}

func NewHandler(skillService service.SkillService) *Handler {
	return &Handler{
		skillService: skillService,
	}
}

// Response structs
type SkillResponse struct {
	SkillID   string `json:"skill_id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateSkillRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type UpdateSkillRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type UserSkillRequest struct {
	SkillID string `json:"skill_id" binding:"required,uuid"`
}

// GetAllSkills godoc
// @Summary Get all skills
// @Description Get a list of all available skills
// @Tags skills
// @Accept json
// @Produce json
// @Success 200 {array} SkillResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/skills [get]
func (h *Handler) GetAllSkills(c *gin.Context) {
	skills, err := h.skillService.GetAllSkills()
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	var response []SkillResponse
	for _, skill := range skills {
		response = append(response, SkillResponse{
			SkillID:   skill.SkillID.String(),
			Name:      skill.Name,
			CreatedAt: skill.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetSkill godoc
// @Summary Get skill by ID
// @Description Get a specific skill by its ID
// @Tags skills
// @Accept json
// @Produce json
// @Param id path string true "Skill ID"
// @Success 200 {object} SkillResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/skills/{id} [get]
func (h *Handler) GetSkill(c *gin.Context) {
	skillIDStr := c.Param("id")
	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid skill ID"})
		return
	}

	skill, err := h.skillService.GetSkillByID(skillID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Skill not found"})
			return
		}
		resp.InternalError(c, err)
		return
	}

	response := SkillResponse{
		SkillID:   skill.SkillID.String(),
		Name:      skill.Name,
		CreatedAt: skill.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// CreateSkill godoc
// @Summary Create new skill
// @Description Create a new skill (admin only)
// @Tags skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param skill body CreateSkillRequest true "Skill data"
// @Success 201 {object} SkillResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /api/v1/admin/skills [post]
func (h *Handler) CreateSkill(c *gin.Context) {
	var req CreateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	skill, err := h.skillService.CreateSkill(req.Name)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		resp.InternalError(c, err)
		return
	}

	response := SkillResponse{
		SkillID:   skill.SkillID.String(),
		Name:      skill.Name,
		CreatedAt: skill.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateSkill godoc
// @Summary Update skill
// @Description Update an existing skill (admin only)
// @Tags skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Skill ID"
// @Param skill body UpdateSkillRequest true "Updated skill data"
// @Success 200 {object} SkillResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/skills/{id} [put]
func (h *Handler) UpdateSkill(c *gin.Context) {
	skillIDStr := c.Param("id")
	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid skill ID"})
		return
	}

	var req UpdateSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	skill, err := h.skillService.UpdateSkill(skillID, req.Name)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Skill not found"})
			return
		}
		resp.InternalError(c, err)
		return
	}

	response := SkillResponse{
		SkillID:   skill.SkillID.String(),
		Name:      skill.Name,
		CreatedAt: skill.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSkill godoc
// @Summary Delete skill
// @Description Delete a skill (admin only)
// @Tags skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Skill ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/admin/skills/{id} [delete]
func (h *Handler) DeleteSkill(c *gin.Context) {
	skillIDStr := c.Param("id")
	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid skill ID"})
		return
	}

	err = h.skillService.DeleteSkill(skillID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Skill not found"})
		case errors.Is(err, apperrors.ErrInUse):
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Skill is in use and cannot be deleted"})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// AddOfferedSkill godoc
// @Summary Add offered skill
// @Description Add a skill to user's offered skills
// @Tags user-skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param skill body UserSkillRequest true "Skill to add"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/skills/offered [post]
func (h *Handler) AddOfferedSkill(c *gin.Context) {
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

	var req UserSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	skillID, err := uuid.Parse(req.SkillID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid skill ID"})
		return
	}

	err = h.skillService.AddOfferedSkill(userID, skillID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Skill not found"})
		case errors.Is(err, apperrors.ErrConflict):
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Skill already in offered skills"})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveOfferedSkill godoc
// @Summary Remove offered skill
// @Description Remove a skill from user's offered skills
// @Tags user-skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Skill ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/skills/offered/{id} [delete]
func (h *Handler) RemoveOfferedSkill(c *gin.Context) {
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

	skillIDStr := c.Param("id")
	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid skill ID"})
		return
	}

	err = h.skillService.RemoveOfferedSkill(userID, skillID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Offered skill not found"})
		} else {
			resp.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// AddWantedSkill godoc
// @Summary Add wanted skill
// @Description Add a skill to user's wanted skills
// @Tags user-skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param skill body UserSkillRequest true "Skill to add"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/skills/wanted [post]
func (h *Handler) AddWantedSkill(c *gin.Context) {
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

	var req UserSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	skillID, err := uuid.Parse(req.SkillID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid skill ID"})
		return
	}

	err = h.skillService.AddWantedSkill(userID, skillID)
	if err != nil {
		switch {
		case errors.Is(err, apperrors.ErrNotFound):
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Skill not found"})
		case errors.Is(err, apperrors.ErrConflict):
			c.JSON(http.StatusConflict, ErrorResponse{Error: "Skill already in wanted skills"})
		default:
			resp.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// RemoveWantedSkill godoc
// @Summary Remove wanted skill
// @Description Remove a skill from user's wanted skills
// @Tags user-skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Skill ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/users/skills/wanted/{id} [delete]
func (h *Handler) RemoveWantedSkill(c *gin.Context) {
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

	skillIDStr := c.Param("id")
	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid skill ID"})
		return
	}

	err = h.skillService.RemoveWantedSkill(userID, skillID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "Wanted skill not found"})
		} else {
			resp.InternalError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserOfferedSkills godoc
// @Summary Get user's offered skills
// @Description Get all skills offered by the authenticated user
// @Tags user-skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SkillResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/skills/offered [get]
func (h *Handler) GetUserOfferedSkills(c *gin.Context) {
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

	skills, err := h.skillService.GetUserOfferedSkills(userID)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	var response []SkillResponse
	for _, skill := range skills {
		response = append(response, SkillResponse{
			SkillID:   skill.SkillID.String(),
			Name:      skill.Name,
			CreatedAt: skill.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetUserWantedSkills godoc
// @Summary Get user's wanted skills
// @Description Get all skills wanted by the authenticated user
// @Tags user-skills
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SkillResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/skills/wanted [get]
func (h *Handler) GetUserWantedSkills(c *gin.Context) {
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

	skills, err := h.skillService.GetUserWantedSkills(userID)
	if err != nil {
		resp.InternalError(c, err)
		return
	}

	var response []SkillResponse
	for _, skill := range skills {
		response = append(response, SkillResponse{
			SkillID:   skill.SkillID.String(),
			Name:      skill.Name,
			CreatedAt: skill.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	c.JSON(http.StatusOK, response)
}

// ResolveSkill — POST /api/v1/skills/resolve  body {name}
// Returns the canonical skill (creating one if no exact/alias match).
func (h *Handler) ResolveSkill(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required,min=2,max=100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	skill, err := h.skillService.ResolveOrCreate(req.Name)
	if err != nil {
		if errors.Is(err, apperrors.ErrValidation) {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		resp.InternalError(c, err)
		return
	}
	_ = uuid.Nil
	c.JSON(http.StatusOK, SkillResponse{
SkillID:   skill.SkillID.String(),
		Name:      skill.Name,
		CreatedAt: skill.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListCategories — GET /api/v1/skills/categories (public)
func (h *Handler) ListCategories(c *gin.Context) {
	cats, err := h.skillService.ListCategories()
	if err != nil {
		resp.InternalError(c, err)
		return
	}
	type categoryDTO struct {
		CategoryID string  `json:"category_id"`
		Name       string  `json:"name"`
		Slug       string  `json:"slug"`
		ParentID   *string `json:"parent_id"`
		SortOrder  int     `json:"sort_order"`
	}
	out := make([]categoryDTO, 0, len(cats))
	for _, c := range cats {
		var parent *string
		if c.ParentID != nil {
			s := c.ParentID.String()
			parent = &s
		}
		out = append(out, categoryDTO{
CategoryID: c.CategoryID.String(),
			Name:       c.Name,
			Slug:       c.Slug,
			ParentID:   parent,
			SortOrder:  c.SortOrder,
		})
	}
	c.JSON(http.StatusOK, out)
}
