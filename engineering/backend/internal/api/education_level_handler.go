package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createEducationLevelInput struct {
	Level string `json:"level" binding:"required,max=100"`
}

type updateEducationLevelInput struct {
	Level string `json:"level" binding:"required,max=100"`
}

type educationLevelListResponse struct {
	Count           int                     `json:"count"`
	EducationLevels []models.EducationLevel `json:"education_levels"`
}

// Education level CRUD
func (h *Handler) CreateEducationLevelHandler(c *gin.Context) {
	var in createEducationLevelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item := models.EducationLevel{Level: in.Level}
	if err := h.EducationLevels.Create(c.Request.Context(), &item); err != nil {
		respondRepoError(c, "Failed to create education level", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetAllEducationLevelsHandler(c *gin.Context) {
	items, err := h.EducationLevels.GetAll(c.Request.Context(), 0, 0, nil, nil)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch education levels", err)
		return
	}
	c.JSON(http.StatusOK, educationLevelListResponse{Count: len(items), EducationLevels: items})
}

func (h *Handler) GetEducationLevelByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	item, err := h.EducationLevels.GetByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch education level", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) UpdateEducationLevelHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	var in updateEducationLevelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.EducationLevels.Update(c.Request.Context(), id, map[string]interface{}{"level": in.Level})
	if err != nil {
		respondRepoError(c, "Failed to update education level", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) DeleteEducationLevelHandler(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := h.EducationLevels.Delete(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete education level", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Education level deleted successfully"})
}
