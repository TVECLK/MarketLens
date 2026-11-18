package repositories_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"marketlens-go-backend/internal/repositories"
)

func TestIsTransientDBError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"no SQLSTATE (dropped connection)", errors.New("connection reset by peer"), true},
		{"connection exception 08006", &pgconn.PgError{Code: "08006"}, true},
		{"serialization failure", &pgconn.PgError{Code: "40001"}, true},
		{"deadlock", &pgconn.PgError{Code: "40P01"}, true},
		{"out of memory 53200", &pgconn.PgError{Code: "53200"}, true},
		{"admin shutdown", &pgconn.PgError{Code: "57P01"}, true},
		{"cannot connect now", &pgconn.PgError{Code: "57P03"}, true},
		{"wrapped transient", fmt.Errorf("save: %w", &pgconn.PgError{Code: "40001"}), true},
		{"foreign key violation", &pgconn.PgError{Code: "23503"}, false},
		{"unique violation", &pgconn.PgError{Code: "23505"}, false},
		{"invalid enum value (class 22)", &pgconn.PgError{Code: "22P02"}, false},
		{"undefined table", &pgconn.PgError{Code: "42P01"}, false},
		{"lock timeout", &pgconn.PgError{Code: "55P03"}, true},
		{"statement timeout or cancellation", &pgconn.PgError{Code: "57014"}, true},
		{"object in use 55006 is not on the retry list", &pgconn.PgError{Code: "55006"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repositories.IsTransientDBError(tt.err); got != tt.want {
				t.Errorf("IsTransientDBError() = %v, want %v", got, tt.want)
			}
		})
	}
}
