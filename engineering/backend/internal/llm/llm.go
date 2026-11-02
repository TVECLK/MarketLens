package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrTransient marks a failure as retryable: a network problem, the LLM
// timing out (408) or rate limiting (429) or erroring (5xx), or the LLM
// responding with unparseable/empty content. None of these are about the
// prompt's own content — a retry later should succeed.
var ErrTransient = errors.New("transient llm failure")

// Config holds the chat-completion provider's settings, loaded once from
// the environment in main.go and passed in explicitly — see NewClient.
// This package has no knowledge of which vendor is behind APIURL.
type Config struct {
	APIKey string
	APIURL string
	Model  string
}

// Validate reports which required fields are missing. Call this once at
// startup so a misconfigured deployment fails fast instead of every chat
// call failing at request time with an opaque "unauthorized" or
// "connection refused" error.
func (c Config) Validate() error {
	var missing []string
	if c.APIKey == "" {
		missing = append(missing, "LLM_API_KEY")
	}
	if c.APIURL == "" {
		missing = append(missing, "LLM_API_URL")
	}
	if c.Model == "" {
		missing = append(missing, "LLM_MODEL")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}
	return nil
}

// Client asks a chat-completion model to answer a free-form prompt.
type Client struct {
	httpClient *http.Client
	apiKey     string
	apiURL     string
	model      string
}

// NewClient builds a client from cfg.
func NewClient(cfg Config, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		apiKey:     cfg.APIKey,
		apiURL:     cfg.APIURL,
		model:      cfg.Model,
	}
}

type llmRequest struct {
	Model       string       `json:"model"`
	Messages    []llmMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	// ResponseFormat asks the model to return strict JSON with no
	// markdown code-fence wrapping. Requires the word "json" to appear
	// somewhere in the prompt.
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type llmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type llmResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Chat sends a single-turn prompt to the LLM and returns the raw response
// content string.
func (c *Client) Chat(ctx context.Context, prompt string) (string, error) {
	reqBody, err := json.Marshal(llmRequest{
		Model:          c.model,
		Messages:       []llmMessage{{Role: "user", Content: prompt}},
		Temperature:    0.0,
		ResponseFormat: &responseFormat{Type: "json_object"},
	})
	if err != nil {
		return "", fmt.Errorf("marshaling llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("building llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: calling llm: %v", ErrTransient, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusRequestTimeout || resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return "", fmt.Errorf("%w: llm returned status %d", ErrTransient, resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm returned status %d", resp.StatusCode)
	}

	var parsed llmResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("%w: decoding llm response: %v", ErrTransient, err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("%w: llm response had no choices", ErrTransient)
	}

	return parsed.Choices[0].Message.Content, nil
}
