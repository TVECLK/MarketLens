package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"marketlens-go-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// parseIDParam reads and validates the ":id" path param used by every
// CRUD handler's GetByID/Update/Delete, writing the 400 response itself
// on failure.
func parseIDParam(c *gin.Context) (uint, bool) {
	return parseIDParamMsg(c, "Invalid id parameter")
}

// parseIDParamMsg is parseIDParam with a caller-chosen error message, for
// the handlers that name the parent entity in their 400 (e.g. "Invalid
// major group id parameter").
func parseIDParamMsg(c *gin.Context, message string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": message})
		return 0, false
	}
	return uint(id), true
}

// dateRange is the parsed from-date/to-date query pair every analytics
// endpoint requires. Order validation (to >= from) is the service layer's
// job, since it's a business rule shared with the MCP tools, not a parsing
// concern.
type dateRange struct {
	FromStr string
	ToStr   string
	From    time.Time
	To      time.Time
}

// parseDateRangeQuery reads and parses the required "from-date"/"to-date"
// query params (format YYYY-MM-DD), writing the 400 response itself on any
// failure. ok is false iff it already wrote a response. The actual parsing
// lives in service.ParseDateRange, shared with the MCP tools.
func parseDateRangeQuery(c *gin.Context) (dateRange, bool) {
	fromDateStr := c.Query("from-date")
	toDateStr := c.Query("to-date")

	from, to, err := service.ParseDateRange(fromDateStr, toDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return dateRange{}, false
	}

	return dateRange{FromStr: fromDateStr, ToStr: toDateStr, From: from, To: to}, true
}

// parseOptionalDateRangeQuery is parseDateRangeQuery, but treats both
// params being absent as "no date filter" (nil, nil) instead of a 400 —
// for parent-lookup/list endpoints where a date range narrows an
// otherwise-unfiltered result, rather than being mandatory like the
// analytics endpoints.
func parseOptionalDateRangeQuery(c *gin.Context) (fromDate, toDate *time.Time, ok bool) {
	if c.Query("from-date") == "" && c.Query("to-date") == "" {
		return nil, nil, true
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return nil, nil, false
	}
	return &dr.From, &dr.To, true
}

// messageResponse is the shared shape for endpoints that only report
// success as a human-readable message (Delete, crawler-run finalize) — a
// typed stand-in for the gin.H{"message": "..."} every one of those used.
type messageResponse struct {
	Message string `json:"message"`
}

// respondError logs the real error server-side and sends only a generic
// message to the client — returning err.Error() directly risks leaking
// database driver/internal details.
func respondError(c *gin.Context, status int, publicMessage string, err error) {
	if err != nil {
		log.Printf("%s: %v", publicMessage, err)
	}
	c.JSON(status, gin.H{"error": publicMessage})
}

// postgresUniqueViolation is Postgres's error code for a unique constraint
// violation (class 23 — Integrity Constraint Violation).
const postgresUniqueViolation = "23505"

// respondRepoError maps a repository/service error to the right HTTP
// status: a missing row (gorm.ErrRecordNotFound) is a 404, a rejected
// business rule (service.ErrValidation) or a unique constraint violation is
// a 400, anything else is an unexpected server error (logged, 500).
func respondRepoError(c *gin.Context, publicMessage string, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": publicMessage + ": not found"})
		return
	}
	if errors.Is(err, service.ErrValidation) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
		c.JSON(http.StatusBadRequest, gin.H{"error": publicMessage + ": already exists"})
		return
	}
	respondError(c, http.StatusInternalServerError, publicMessage, err)
}
