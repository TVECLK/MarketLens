package api

import (
	"marketlens-go-backend/internal/ingestion"
	"net/http"

	"github.com/gin-gonic/gin"
)

type batchSaveJobsResponse struct {
	Message     string            `json:"message"`
	Submitted   int               `json:"submitted"`
	FailedCount int               `json:"failed_count"`
	FailedJobs  []failedJobOutput `json:"failed_jobs"`
}

type failedJobOutput struct {
	JobID      string `json:"job_id"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}

func toFailedJobOutputs(failed []ingestion.FailedJob) []failedJobOutput {
	out := make([]failedJobOutput, len(failed))
	for i, f := range failed {
		out[i] = failedJobOutput{JobID: f.JobID, StatusCode: f.StatusCode, Error: f.Error}
	}
	return out
}

type batchSaveJobInput struct {
	JobID        string `json:"job_id"`
	Employer     string `json:"employer"`
	JobRole      string `json:"job_role"`
	Location     string `json:"location"`
	Description  string `json:"description"`
	CrawlerRunID uint   `json:"crawler_run_id"`
	Source       string `json:"source"`
}

func (h *Handler) BatchSaveJobsHandler(c *gin.Context) {
	var in []batchSaveJobInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body structural mapping", err)
		return
	}

	if len(in) == 0 {
		respondError(c, http.StatusBadRequest, "The job batch cannot be empty", nil)
		return
	}

	rawJobs := make([]ingestion.RawJobInput, len(in))
	for i, j := range in {
		rawJobs[i] = ingestion.RawJobInput{
			JobID:        j.JobID,
			Employer:     j.Employer,
			JobRole:      j.JobRole,
			Location:     j.Location,
			Description:  j.Description,
			CrawlerRunID: j.CrawlerRunID,
			Source:       j.Source,
		}
	}

	failedJobs := h.ingestion.ProcessBatch(c.Request.Context(), rawJobs)

	c.JSON(http.StatusOK, batchSaveJobsResponse{Message: "Batch processed", Submitted: len(rawJobs), FailedCount: len(failedJobs), FailedJobs: toFailedJobOutputs(failedJobs)})
}

type reconcileStaleVacanciesInput struct {
	CrawlerRunID uint `json:"crawler_run_id" binding:"required"`
}

type reconcileStaleVacanciesResponse struct {
	Message             string `json:"message"`
	ReconciledStaleJobs int64  `json:"reconciled_stale_jobs"`
}

func (h *Handler) ReconcileStaleVacanciesHandler(c *gin.Context) {
	var in reconcileStaleVacanciesInput
	if err := c.ShouldBindJSON(&in); err != nil {
		respondError(c, http.StatusBadRequest, "Missing active execution tracking sequence identifier criteria", err)
		return
	}

	closedCount, err := h.JobPosts.ReconcileStaleVacancies(c.Request.Context(), in.CrawlerRunID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Global snapshot reconciliation sweeping routine failed", err)
		return
	}

	c.JSON(http.StatusOK, reconcileStaleVacanciesResponse{Message: "System-wide stale vacancy cleanup sweep finalized", ReconciledStaleJobs: closedCount})
}
