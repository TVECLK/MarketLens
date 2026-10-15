package api

import (
	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetJobsByBucketKeysHandler(c *gin.Context) {
	var req struct {
		BucketKeys []string `json:"bucket_keys" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request structure payload mapping", err)
		return
	}

	matchingJobs, err := h.JobPosts.GetJobsByBucketKeys(c.Request.Context(), req.BucketKeys)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "LSH bucket radar execution sequence encountered a processing exception", err)
		return
	}

	c.JSON(http.StatusOK, matchingJobs)
}

type batchSaveJobsResponse struct {
	Message     string                `json:"message"`
	Submitted   int                   `json:"submitted"`
	FailedCount int                   `json:"failed_count"`
	FailedJobs  []ingestion.FailedJob `json:"failed_jobs"`
}

func (h *Handler) BatchSaveJobsHandler(c *gin.Context) {
	var rawJobs []ingestion.RawJobInput
	if err := c.ShouldBindJSON(&rawJobs); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body structural mapping", err)
		return
	}

	if len(rawJobs) == 0 {
		respondError(c, http.StatusBadRequest, "The job batch cannot be empty", nil)
		return
	}

	failedJobs := h.ingestion.ProcessBatch(c.Request.Context(), rawJobs)

	c.JSON(http.StatusOK, batchSaveJobsResponse{Message: "Batch processed", Submitted: len(rawJobs), FailedCount: len(failedJobs), FailedJobs: failedJobs})
}

type reconcileStaleVacanciesResponse struct {
	Message             string `json:"message"`
	ReconciledStaleJobs int64  `json:"reconciled_stale_jobs"`
}

func (h *Handler) ReconcileStaleVacanciesHandler(c *gin.Context) {
	var payload models.ReconciliationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		respondError(c, http.StatusBadRequest, "Missing active execution tracking sequence identifier criteria", err)
		return
	}

	closedCount, err := h.JobPosts.ReconcileStaleVacancies(c.Request.Context(), payload.CrawlerRunID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Global snapshot reconciliation sweeping routine failed", err)
		return
	}

	c.JSON(http.StatusOK, reconcileStaleVacanciesResponse{Message: "System-wide stale vacancy cleanup sweep finalized", ReconciledStaleJobs: closedCount})
}
