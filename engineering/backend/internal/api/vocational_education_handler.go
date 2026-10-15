package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createVocationalEducationInput struct {
	Level string `json:"level" binding:"required,max=50"`
}

type updateVocationalEducationInput struct {
	Level string `json:"level" binding:"required,max=50"`
}

type vocationalEducationListResponse struct {
	Count                 int                           `json:"count"`
	VocationalEducations []models.VocationalEducation `json:"vocational_educations"`
}

// Vocational Education CRUD
func (h *Handler) CreateVocationalEducationHandler(c *gin.Context) {
	var in createVocationalEducationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.VocationalEducation{Level: in.Level}
	if err := h.VocationalEducations.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create vocational education", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllVocationalEducationsHandler(c *gin.Context) {
	items, err := h.VocationalEducations.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch vocational education levels", err)
		return
	}
	c.JSON(http.StatusOK, vocationalEducationListResponse{Count: len(items), VocationalEducations: items})
}

func (h *Handler) GetVocationalEducationByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.VocationalEducations.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch vocational education", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateVocationalEducationHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateVocationalEducationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.VocationalEducations.Update(c.Request.Context(), id, map[string]interface{}{"level": in.Level})
	if err != nil {
		respondRepoError(c, "Failed to update vocational education", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteVocationalEducationHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.VocationalEducations.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete vocational education", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Vocational education deleted successfully"})
}
