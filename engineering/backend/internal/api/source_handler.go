package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// IsActive is a pointer so binding:"required" accepts false (a plain bool
// would treat false as missing).
type updateSourceStatusInput struct {
	IsActive *bool `json:"is_active" binding:"required"`
}

// IsActive is optional: nil means active.
type createSourceInput struct {
	Name     string `json:"name"      binding:"required,max=255"`
	Type     string `json:"type"      binding:"required"`
	IsActive *bool  `json:"is_active"`
}

type renameSourceInput struct {
	Name string `json:"name" binding:"required,max=255"`
}

type sourceListResponse struct {
	Count   int             `json:"count"`
	Sources []models.Source `json:"sources"`
}

// ListSourcesHandler lists sources, optionally filtered by
// ?type=auto|manual and ?is-active=true|false.
func (h *Handler) ListSourcesHandler(c *gin.Context) {
	var isActive *bool
	if raw := c.Query("is-active"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is-active, must be true or false"})
			return
		}
		isActive = &v
	}

	items, err := h.Sources.ListSources(c.Request.Context(), c.Query("type"), isActive)
	if err != nil {
		respondRepoError(c, "Failed to fetch sources", err)
		return
	}
	c.JSON(http.StatusOK, sourceListResponse{Count: len(items), Sources: items})
}

func (h *Handler) GetSourceByIDHandler(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	item, err := h.Sources.GetSourceByID(c.Request.Context(), id)
	if err != nil {
		respondRepoError(c, "Failed to fetch source", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateSourceHandler adds a source, or restores a deleted one with the same
// name. A name used by a live source is a 400 "already exists".
func (h *Handler) CreateSourceHandler(c *gin.Context) {
	var in createSourceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.Sources.CreateSource(c.Request.Context(), in.Name, in.Type, in.IsActive)
	if err != nil {
		respondRepoError(c, "Failed to create source", err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

// RenameSourceHandler renames a manual source. Auto sources are a 400,
// because their name must match the crawler's SOURCE_NAME.
func (h *Handler) RenameSourceHandler(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var in renameSourceInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.Sources.RenameSource(c.Request.Context(), id, in.Name)
	if err != nil {
		respondRepoError(c, "Failed to rename source", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// DeleteSourceHandler soft-deletes a source. Its jobs keep their link to it.
func (h *Handler) DeleteSourceHandler(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.Sources.DeleteSource(c.Request.Context(), id); err != nil {
		respondRepoError(c, "Failed to delete source", err)
		return
	}
	c.JSON(http.StatusOK, messageResponse{Message: "Source deleted successfully"})
}

// UpdateSourceStatusHandler enables or disables a source.
func (h *Handler) UpdateSourceStatusHandler(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var in updateSourceStatusInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}
	item, err := h.Sources.UpdateSourceStatus(c.Request.Context(), id, *in.IsActive)
	if err != nil {
		respondRepoError(c, "Failed to update source status", err)
		return
	}
	c.JSON(http.StatusOK, item)
}

// CrawlerListSourcesHandler returns the sources the crawler should run:
// type auto and active.
func (h *Handler) CrawlerListSourcesHandler(c *gin.Context) {
	items, err := h.Sources.ListActiveAutoSources(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch crawler sources", err)
		return
	}
	c.JSON(http.StatusOK, sourceListResponse{Count: len(items), Sources: items})
}
