package llm_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"marketlens-go-backend/internal/llm"
)

func TestChat_StatusClassification(t *testing.T) {
	tests := []struct {
		status        int
		wantTransient bool
	}{
		{http.StatusRequestTimeout, true},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
		// A rejected key or endpoint is the deployment's problem, not the
		// job's, so it must be retryable.
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, true},
		{http.StatusNotFound, true},
		// The LLM refusing this particular prompt is not transient.
		{http.StatusBadRequest, false},
		{http.StatusUnprocessableEntity, false},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			t.Cleanup(server.Close)

			client := llm.NewClient(llm.Config{APIKey: "k", Model: "m", APIURL: server.URL}, server.Client())
			_, err := client.Chat(context.Background(), "prompt")
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if got := errors.Is(err, llm.ErrTransient); got != tt.wantTransient {
				t.Errorf("status %d: errors.Is(err, ErrTransient) = %t, want %t (err: %v)", tt.status, got, tt.wantTransient, err)
			}
		})
	}
}
