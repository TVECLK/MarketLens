package llm_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
		{http.StatusPaymentRequired, true},
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

func TestConfig_Validate(t *testing.T) {
	full := llm.Config{APIKey: "k", APIURL: "http://llm", Model: "m"}

	tests := []struct {
		name        string
		cfg         llm.Config
		wantMissing []string
	}{
		{"complete config is valid", full, nil},
		{"missing key", llm.Config{APIURL: full.APIURL, Model: full.Model}, []string{"LLM_API_KEY"}},
		{"missing url", llm.Config{APIKey: full.APIKey, Model: full.Model}, []string{"LLM_API_URL"}},
		{"missing model", llm.Config{APIKey: full.APIKey, APIURL: full.APIURL}, []string{"LLM_MODEL"}},
		{"everything missing", llm.Config{}, []string{"LLM_API_KEY", "LLM_API_URL", "LLM_MODEL"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if len(tt.wantMissing) == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			for _, name := range tt.wantMissing {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("error %q does not mention %s", err, name)
				}
			}
		})
	}
}

func TestChat_SendsRequestAndReturnsContent(t *testing.T) {
	var gotMethod, gotAuth, gotContentType string
	var gotBody struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Temperature    float64 `json:"temperature"`
		ResponseFormat struct {
			Type string `json:"type"`
		} `json:"response_format"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"id\": 7}"}}]}`))
	}))
	t.Cleanup(server.Close)

	client := llm.NewClient(llm.Config{APIKey: "secret", Model: "test-model", APIURL: server.URL}, server.Client())
	got, err := client.Chat(context.Background(), "pick one, as json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != `{"id": 7}` {
		t.Errorf("content = %q, want %q", got, `{"id": 7}`)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotAuth != "Bearer secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer secret")
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotBody.Model != "test-model" {
		t.Errorf("model = %q, want test-model", gotBody.Model)
	}
	if len(gotBody.Messages) != 1 || gotBody.Messages[0].Role != "user" || gotBody.Messages[0].Content != "pick one, as json" {
		t.Errorf("unexpected messages: %+v", gotBody.Messages)
	}
	if gotBody.Temperature != 0 {
		t.Errorf("temperature = %v, want 0", gotBody.Temperature)
	}
	if gotBody.ResponseFormat.Type != "json_object" {
		t.Errorf("response_format.type = %q, want json_object", gotBody.ResponseFormat.Type)
	}
}

func TestChat_BadResponseBodyIsTransient(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"unparseable JSON", "not json"},
		{"empty body", ""},
		{"no choices", `{"choices":[]}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(tt.body))
			}))
			t.Cleanup(server.Close)

			client := llm.NewClient(llm.Config{APIKey: "k", Model: "m", APIURL: server.URL}, server.Client())
			_, err := client.Chat(context.Background(), "prompt")
			if !errors.Is(err, llm.ErrTransient) {
				t.Errorf("expected a transient error, got %v", err)
			}
		})
	}
}

// TestChat_RequestFailures checks failures that happen before any usable
// response: network problems and cancellation are transient, while a URL
// that cannot be turned into a request fails identically every time.
func TestChat_RequestFailures(t *testing.T) {
	tests := []struct {
		name          string
		wantTransient bool
		// setup returns the API URL, http client and context for the call.
		setup func(t *testing.T) (string, *http.Client, context.Context)
	}{
		{
			name:          "nothing listening is transient",
			wantTransient: true,
			setup: func(t *testing.T) (string, *http.Client, context.Context) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				url, httpClient := server.URL, server.Client()
				server.Close()
				return url, httpClient, context.Background()
			},
		},
		{
			name:          "canceled context is transient",
			wantTransient: true,
			setup: func(t *testing.T) (string, *http.Client, context.Context) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				t.Cleanup(server.Close)
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return server.URL, server.Client(), ctx
			},
		},
		{
			name:          "malformed LLM_API_URL is not transient",
			wantTransient: false,
			setup: func(t *testing.T) (string, *http.Client, context.Context) {
				return "http://bad host/\x7f", http.DefaultClient, context.Background()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, httpClient, ctx := tt.setup(t)

			client := llm.NewClient(llm.Config{APIKey: "k", Model: "m", APIURL: url}, httpClient)
			_, err := client.Chat(ctx, "prompt")
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if got := errors.Is(err, llm.ErrTransient); got != tt.wantTransient {
				t.Errorf("errors.Is(err, ErrTransient) = %t, want %t (err: %v)", got, tt.wantTransient, err)
			}
		})
	}
}
