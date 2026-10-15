package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createFormalityInput struct {
	FormalityType string `json:"formality_type" binding:"required,max=50"`
}

type updateFormalityInput struct {
	FormalityType string `json:"formality_type" binding:"required,max=50"`
}

type formalityListResponse struct {
	Count       int                 `json:"count"`
	Formalities []models.Formality `json:"formalities"`
}

// Formality CRUD
func (h *Handler) CreateFormalityHandler(c *gin.Context) {
	var in createFormalityInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.Formality{FormalityType: in.FormalityType}
	if err := h.Formalities.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create formality", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllFormalitiesHandler(c *gin.Context) {
	items, err := h.Formalities.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch formalities", err)
		return
	}
	c.JSON(http.StatusOK, formalityListResponse{Count: len(items), Formalities: items})
}

func (h *Handler) GetFormalityByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.Formalities.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch formality", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateFormalityHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateFormalityInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.Formalities.Update(c.Request.Context(), id, map[string]interface{}{"formality_type": in.FormalityType})
	if err != nil {
		respondRepoError(c, "Failed to update formality", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteFormalityHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.Formalities.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete formality", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Formality deleted successfully"})
}
