package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"marketlens-go-backend/internal/service"
)

func TestRespondRepoError_Status(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"missing row", gorm.ErrRecordNotFound, http.StatusNotFound},
		{"validation", service.ValidationErrorf("bad"), http.StatusBadRequest},
		{"unique violation", &pgconn.PgError{Code: "23505"}, http.StatusConflict},
		{"foreign key violation", &pgconn.PgError{Code: "23503"}, http.StatusConflict},
		{"transient db error", &pgconn.PgError{Code: "40001"}, http.StatusServiceUnavailable},
		{"unexpected", errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			respondRepoError(c, "Failed", tt.err)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}
