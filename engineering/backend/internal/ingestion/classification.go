package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"marketlens-go-backend/internal/llm"
)

// ErrTransientExtractionFailure marks a failure in the extract/classify
// pipeline as retryable: a network problem, the LLM timing out (408) or
// rate limiting (429) or erroring (5xx), the LLM responding with
// unparseable/empty content, or a DB read failing while fetching
// reference data. None of these are about the job's own content — a
// retry later should succeed.
var ErrTransientExtractionFailure = errors.New("transient extraction failure")

// wrapLLMErr translates the generic llm package's transient-failure
// signal into this package's own retry signal, so downstream retry
// checks (errors.Is(err, ErrTransientExtractionFailure)) keep working
// regardless of which layer detected the transient failure.
func wrapLLMErr(err error) error {
	if errors.Is(err, llm.ErrTransient) {
		return fmt.Errorf("%w: %v", ErrTransientExtractionFailure, err)
	}
	return err
}

// Option is the common (id, name) shape every hierarchy level's rows are
// converted into before being shown to the LLM.
type Option struct {
	ID   uint
	Name string
}

// toOptions converts a slice of any model into []Option, given a
// function that knows how to pull the id/name off that specific type.
func toOptions[T any](items []T, extract func(T) (uint, string)) []Option {
	options := make([]Option, len(items))
	for i, item := range items {
		id, name := extract(item)
		options[i] = Option{ID: id, Name: name}
	}
	return options
}

// PipelineConfig holds tunable ingestion-pipeline settings.
type PipelineConfig struct {
	DedupThreshold   float64       // Jaccard similarity cutoff for duplicate job detection
	MetadataCacheTTL time.Duration // how long MetadataBuilder caches reference-data options
}

type pickResult struct {
	ID uint `json:"id"`
}

// stripJSONFences removes a leading/trailing markdown code fence
// (```json ... ``` or plain ``` ... ```) if present. Belt-and-suspenders
// alongside response_format: json_object — models occasionally wrap
// JSON in a fence anyway, and this costs nothing when there isn't one.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	if nl := strings.IndexByte(s, '\n'); nl != -1 && strings.TrimSpace(s[:nl]) != "" {
		// first line was a language tag like "json" — drop it
		s = s[nl+1:]
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// AskToPick asks the LLM to choose the single best matching option and
// returns its id (0 if there were no options to choose from).
func AskToPick(ctx context.Context, client *llm.Client, jobText string, options []Option, levelName string) (uint, error) {
	if len(options) == 0 {
		return 0, nil
	}

	prompt := buildAskToPickPrompt(jobText, options, levelName)

	content, err := client.Chat(ctx, prompt)
	if err != nil {
		return 0, wrapLLMErr(err)
	}

	var result pickResult
	if err := json.Unmarshal([]byte(stripJSONFences(content)), &result); err != nil {
		return 0, fmt.Errorf("%w: parsing chosen id from llm content: %v", ErrTransientExtractionFailure, err)
	}

	for _, o := range options {
		if o.ID == result.ID {
			return result.ID, nil
		}
	}

	// The LLM returned an id that wasn't actually one of the options it
	// was given — a hallucination, not a valid pick. Treated the same
	// as "no match" (0) rather than trusted, since a fabricated nonzero
	// id would otherwise pass straight through as a plausible-looking
	// but invalid foreign key.
	log.Printf("llm returned id %d for %q, which was not among the %d offered options — treating as no match", result.ID, levelName, len(options))
	return 0, nil
}

// LevelFetcher fetches the option list for one hierarchy level, given
// the id chosen at the previous level.
type LevelFetcher func(parentID uint) ([]Option, error)

// Level is one step of a classification walk.
type Level struct {
	Name  string
	Fetch LevelFetcher
}

// Walk asks the LLM to pick one option at each level in order, feeding
// the chosen id into the next level's Fetch. Stops early and returns 0
// the moment a level has no options or the LLM doesn't pick one.
func Walk(ctx context.Context, client *llm.Client, jobText string, levels []Level) (uint, error) {
	var parentID uint
	for _, level := range levels {
		options, err := level.Fetch(parentID)
		if err != nil {
			return 0, fmt.Errorf("%w: fetching %s options: %v", ErrTransientExtractionFailure, level.Name, err)
		}

		id, err := AskToPick(ctx, client, jobText, options, level.Name)
		if err != nil {
			return 0, fmt.Errorf("asking LLM to pick %s: %w", level.Name, err)
		}
		if id == 0 {
			return 0, nil
		}
		parentID = id
	}
	return parentID, nil
}
