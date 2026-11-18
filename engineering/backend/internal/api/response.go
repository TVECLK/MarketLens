package api

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// parseIDParam reads and validates the given path param key, writing the
// 400 response itself on failure.
func parseIDParam(c *gin.Context, key string) (uint, bool) {
	return parseIDParamMsg(c, key, "Invalid id parameter")
}

// parseIDParamMsg is parseIDParam with a caller-chosen error message, for
// the handlers that name the parent entity in their 400 (e.g. "Invalid
// major group id parameter").
func parseIDParamMsg(c *gin.Context, key, message string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(key), 10, 64)
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

// Postgres SQLSTATE codes (class 23 — Integrity Constraint Violation)
// that mean the request conflicts with existing data.
const (
	postgresUniqueViolation     = "23505"
	postgresForeignKeyViolation = "23503"
)

// isTransientDBFailure reports whether err is a database outage a retry can
// cure: a retryable Postgres error (see repositories.IsTransientDBError) or
// a failure with no SQLSTATE that is recognizably about the connection — a
// timeout, a refused or reset connection, a dropped driver connection or a
// connection closed mid-read. repositories.IsTransientDBError treats every
// error without a SQLSTATE as transient, which suits the ingestion retry
// loop but would turn any unexpected error here into a 503, so those are
// matched narrowly instead.
func isTransientDBFailure(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return repositories.IsTransientDBError(err)
	}
	var connErr *pgconn.ConnectError
	var netErr net.Error
	return errors.As(err, &connErr) ||
		errors.As(err, &netErr) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, driver.ErrBadConn) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNREFUSED)
}

// respondRepoError maps a repository/service error to the right HTTP
// status: a missing row (gorm.ErrRecordNotFound) is a 404, a rejected
// business rule (service.ErrValidation) is a 400, a unique or foreign key
// violation (duplicate name, deleting a row that is still referenced) is a
// 409, a transient database failure (a retryable Postgres error or a lost connection) is a 503 so the client may retry, and
// anything else is an unexpected server error (logged, 500).
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
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case postgresUniqueViolation:
			c.JSON(http.StatusConflict, gin.H{"error": publicMessage + ": already exists"})
			return
		case postgresForeignKeyViolation:
			c.JSON(http.StatusConflict, gin.H{"error": publicMessage + ": conflicts with related records"})
			return
		}
	}
	if isTransientDBFailure(err) {
		respondError(c, http.StatusServiceUnavailable, publicMessage+": database temporarily unavailable", err)
		return
	}
	respondError(c, http.StatusInternalServerError, publicMessage, err)
}
