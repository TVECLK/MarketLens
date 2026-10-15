package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) StartCrawlerRunHandler(c *gin.Context) {
	run, err := h.CrawlerRuns.CreateCrawlerRun(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to initialize new tracker session initialization block", err)
		return
	}
	c.JSON(http.StatusCreated, run)
}

func (h *Handler) CompleteCrawlerRunHandler(c *gin.Context) {
	id, ok := parseIDParamMsg(c, "Invalid crawler run ID parameter")
	if !ok {
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"` // 'COMPLETED' or 'FAILED'
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Missing state status flag criteria", err)
		return
	}

	if err := h.CrawlerRuns.CompleteCrawlerRun(c.Request.Context(), id, req.Status); err != nil {
		respondRepoError(c, "Failed to execute closure metrics logic sequence", err)
		return
	}

	c.JSON(http.StatusOK, messageResponse{Message: fmt.Sprintf("Crawler run %d finalized with state: %s", id, req.Status)})
}
