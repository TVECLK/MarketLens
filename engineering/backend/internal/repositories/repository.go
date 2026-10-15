package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrPermanentSaveFailure marks a job save failure as non-retryable: a
// missing/invalid foreign key (e.g. a province not registered in
// geo_data) or a Postgres integrity-constraint violation (SQLSTATE class
// 23: foreign key, unique, not-null, check). Retrying the identical save
// will fail the same way every time.
var ErrPermanentSaveFailure = errors.New("permanent save failure")

// isPermanentDBError reports whether err is a Postgres integrity-
// constraint violation, as surfaced by gorm's postgres driver
// (jackc/pgx) via *pgconn.PgError.Code (SQLSTATE).
func isPermanentDBError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "23")
	}
	return false
}

// wrapSaveErr classifies err as permanent (constraint violation) or
// transient (anything else — connection issues, deadlocks, etc.).
func wrapSaveErr(err error, msg string) error {
	if isPermanentDBError(err) {
		return fmt.Errorf("%w: %s: %v", ErrPermanentSaveFailure, msg, err)
	}
	return fmt.Errorf("%s: %w", msg, err)
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
