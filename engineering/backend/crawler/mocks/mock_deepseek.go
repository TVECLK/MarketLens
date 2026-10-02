package mocks

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"marketlens-go-backend/crawler"
)

// Responder decides what content string to return for a given prompt.
// Tests provide this to control DeepSeek's simulated behavior.
type Responder func(prompt string) string

// NewMockDeepSeek starts an httptest.Server that mimics DeepSeek's
// response shape, and returns a *crawler.DeepSeekClient wired to call
// it instead of the real API. The server is registered for cleanup via
// t.Cleanup, so tests don't need to close it themselves.
func NewMockDeepSeek(t *testing.T, respond Responder) *crawler.DeepSeekClient {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Content string `json:"content"`
			} `json:"messages"`
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &req)

		prompt := ""
		if len(req.Messages) > 0 {
			prompt = req.Messages[0].Content
		}

		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": respond(prompt)}},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(server.Close)

	return crawler.NewDeepSeekClientWithURL(server.Client(), server.URL)
}

// FirstListedID scans a classifier prompt (built by AskToPick) for the
// first "N: label" line and returns N — a simple, generic way to have a
// mock respond with a plausible pick without hardcoding specific ids.
func FirstListedID(prompt string) int {
	b := []byte(prompt)
	for i := 0; i < len(b); i++ {
		if b[i] == '\n' && i+1 < len(b) {
			var n, consumed int
			for consumed = i + 1; consumed < len(b) && b[consumed] >= '0' && b[consumed] <= '9'; consumed++ {
				n = n*10 + int(b[consumed]-'0')
			}
			if consumed > i+1 && consumed < len(b) && b[consumed] == ':' {
				return n
			}
		}
	}
	return 0
}