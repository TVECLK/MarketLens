package mcpserver

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"marketlens-go-backend/internal/service"
)

func TestToolError(t *testing.T) {
	pgErr := &pgconn.PgError{
		Code:    "42P01",
		Message: `relation "job_post" does not exist`,
		Detail:  "column secret_column",
	}

	tests := []struct {
		name string
		err  error
		want string // exact message
	}{
		{"record not found", gorm.ErrRecordNotFound, "get_thing: not found"},
		{"wrapped record not found", fmt.Errorf("lookup: %w", gorm.ErrRecordNotFound), "get_thing: not found"},
		{"validation keeps its message", service.ValidationErrorf("to-date must not be before from-date"), "validation failed: to-date must not be before from-date"},
		{"database error is hidden", pgErr, "get_thing: internal error"},
		{"wrapped database error is hidden", fmt.Errorf("query failed: %w", pgErr), "get_thing: internal error"},
		{"unknown error is hidden", errors.New("dial tcp 10.0.0.5:5432: connection refused"), "get_thing: internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toolError("get_thing", tt.err)
			if got == nil || got.Error() != tt.want {
				t.Fatalf("toolError() = %v, want %q", got, tt.want)
			}
			for _, leak := range []string{"relation", "secret_column", "10.0.0.5", "job_post"} {
				if strings.Contains(got.Error(), leak) {
					t.Errorf("message %q leaks %q", got.Error(), leak)
				}
			}
		})
	}
}
