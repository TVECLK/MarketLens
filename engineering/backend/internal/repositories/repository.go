package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// exclusiveUpperBound converts an inclusive end-of-range date (toDate,
// always midnight UTC) into the exclusive upper bound used by every date
// range comparison across this package — the start of the next calendar
// day.
func exclusiveUpperBound(toDate time.Time) time.Time {
	return toDate.AddDate(0, 0, 1)
}

// ErrPermanentSaveFailure marks a job save failure as non-retryable: a
// missing/invalid foreign key (e.g. a province not registered in
// geo_data) or any database error that is not known to be transient
// (constraint violation, invalid value, bad syntax...). Retrying the
// identical save will fail the same way every time.
var ErrPermanentSaveFailure = errors.New("permanent save failure")

// IsTransientDBError reports whether err is a database failure that a
// later retry can be expected to cure, as surfaced by gorm's postgres
// driver (jackc/pgx) via *pgconn.PgError.Code (SQLSTATE): connection
// exceptions (class 08), serialization failure (40001), deadlock
// (40P01), lock timeout (55P03), insufficient resources (class 53),
// statement timeout or cancellation (57014), and admin shutdown /
// cannot-connect-now (57P01, 57P03). Errors with no SQLSTATE (a dropped
// connection, a timeout) are transient too. Every other SQLSTATE is
// treated as permanent, so an unknown error class fails fast instead of
// being retried, with the full LLM extraction rerun each time.
func IsTransientDBError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return true
	}
	switch pgErr.Code {
	case "40001", "40P01", "55P03", "57014", "57P01", "57P03":
		return true
	}
	return strings.HasPrefix(pgErr.Code, "08") || strings.HasPrefix(pgErr.Code, "53")
}

// wrapSaveErr classifies err as transient (see IsTransientDBError) or
// permanent (anything else).
func wrapSaveErr(err error, msg string) error {
	if IsTransientDBError(err) {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return fmt.Errorf("%w: %s: %v", ErrPermanentSaveFailure, msg, err)
}

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

// This function pings the database connection to confirm it's reachable - used for Kubernetes readiness probes
func (r *Store) Ping(ctx context.Context) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
