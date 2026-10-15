package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createGenderInput struct {
	GenderType string `json:"gender_type" binding:"required,max=50"`
}

type updateGenderInput struct {
	GenderType string `json:"gender_type" binding:"required,max=50"`
}

type genderListResponse struct {
	Count   int             `json:"count"`
	Genders []models.Gender `json:"genders"`
}

// Gender CRUD
func (h *Handler) CreateGenderHandler(c *gin.Context) {
	var in createGenderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.Gender{GenderType: in.GenderType}
	if err := h.Genders.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create gender", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllGendersHandler(c *gin.Context) {
	items, err := h.Genders.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch genders", err)
		return
	}
	c.JSON(http.StatusOK, genderListResponse{Count: len(items), Genders: items})
}

func (h *Handler) GetGenderByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.Genders.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch gender", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateGenderHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateGenderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.Genders.Update(c.Request.Context(), id, map[string]interface{}{"gender_type": in.GenderType})
	if err != nil {
		respondRepoError(c, "Failed to update gender", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteGenderHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.Genders.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete gender", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Gender deleted successfully"})
}
