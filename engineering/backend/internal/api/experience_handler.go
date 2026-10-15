package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createExperienceInput struct {
	Name string `json:"name" binding:"required,max=100"`
}

type updateExperienceInput struct {
	Name string `json:"name" binding:"required,max=100"`
}

type experienceListResponse struct {
	Count       int                 `json:"count"`
	Experiences []models.Experience `json:"experiences"`
}

// Experience CRUD
func (h *Handler) GetAllExperiencesHandler(c *gin.Context) {
	experiences, err := h.Experiences.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to retrieve experience levels", err)
		return
	}
	c.JSON(http.StatusOK, experienceListResponse{Count: len(experiences), Experiences: experiences})
}

func (h *Handler) CreateExperienceHandler(c *gin.Context) {
	var in createExperienceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.Experience{Name: in.Name}
	if err := h.Experiences.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create experience", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetExperienceByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.Experiences.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch experience", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateExperienceHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateExperienceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.Experiences.Update(c.Request.Context(), id, map[string]interface{}{"name": in.Name})
	if err != nil {
		respondRepoError(c, "Failed to update experience", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteExperienceHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.Experiences.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete experience", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Experience deleted successfully"})
}
