package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type statusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// HealthzHandler confirms the process itself is up and responding to HTTP - used for Kubernetes liveness probes
func (h *Handler) HealthzHandler(c *gin.Context) {
	c.JSON(http.StatusOK, statusResponse{Status: "ok"})
}

// ReadyzHandler confirms the database connection is reachable - used for Kubernetes readiness probes
func (h *Handler) ReadyzHandler(c *gin.Context) {
	if err := h.pinger.Ping(c.Request.Context()); err != nil {
		log.Printf("readiness check failed: %v", err)
		c.JSON(http.StatusServiceUnavailable, statusResponse{Status: "not ready", Message: "database unreachable"})
		return
	}
	c.JSON(http.StatusOK, statusResponse{Status: "ready"})
}
