package ingestion_test

import (
	"context"
	"errors"
	"testing"

	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/ingestion/mocks"
	"marketlens-go-backend/internal/llm"
	"marketlens-go-backend/internal/service"
)

// pickFirstListed answers every classifier prompt with the first option
// it was offered, so a walk reaches a real leaf id.
func pickFirstListed(prompt string) string {
	return `{"id": ` + itoa(mocks.FirstListedID(prompt)) + `}`
}

func TestClassifier_WalksToLeaf(t *testing.T) {
	const jobText = "Backend engineer building Go microservices"

	tests := []struct {
		name     string
		classify func(ctx context.Context, svcs *service.Services, client *llm.Client, jobText string) (uint, error)
		wantID   uint
	}{
		{
			name: "industry classifier reaches the subclass",
			classify: func(ctx context.Context, svcs *service.Services, client *llm.Client, jobText string) (uint, error) {
				return newIndustryClassifier(svcs, client).Classify(ctx, jobText)
			},
			wantID: 62011,
		},
		{
			name: "occupation classifier reaches the occupation group",
			classify: func(ctx context.Context, svcs *service.Services, client *llm.Client, jobText string) (uint, error) {
				return newOccupationClassifier(svcs, client).Classify(ctx, jobText)
			},
			wantID: 25121,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svcs := newServices(setupDB(t))
			client := mocks.NewMockLLM(t, pickFirstListed)

			id, err := tt.classify(context.Background(), svcs, client, jobText)
			if err != nil {
				t.Fatalf("Classify() error: %v", err)
			}
			if id != tt.wantID {
				t.Errorf("expected leaf id %d, got %d", tt.wantID, id)
			}
		})
	}
}

func TestAskToPick(t *testing.T) {
	tests := []struct {
		name    string
		options []ingestion.Option
		content string // what the mock LLM answers, whatever is asked
		wantID  uint
		wantErr error
	}{
		{
			name:    "accepts an id that was among the options",
			options: []ingestion.Option{{ID: 1, Name: "First"}, {ID: 2, Name: "Second"}},
			content: `{"id": 2}`,
			wantID:  2,
		},
		{
			// A fabricated id is treated as "no match", not trusted: it would
			// otherwise pass through as a plausible but invalid foreign key.
			name:    "rejects a hallucinated id not among the options",
			options: []ingestion.Option{{ID: 1, Name: "First"}, {ID: 2, Name: "Second"}},
			content: `{"id": 999}`,
			wantID:  0,
		},
		{
			name:    "handles a markdown-fenced response",
			options: []ingestion.Option{{ID: 7, Name: "Match"}},
			content: "```json\n{\"id\": 7}\n```",
			wantID:  7,
		},
		{
			name:    "unparseable content is a transient failure",
			options: []ingestion.Option{{ID: 1, Name: "First"}},
			content: "not json at all",
			wantErr: ingestion.ErrTransientExtractionFailure,
		},
		{
			name:    "no options means no LLM call and no match",
			options: nil,
			content: `{"id": 1}`,
			wantID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := mocks.NewMockLLM(t, func(string) string { return tt.content })

			id, err := ingestion.AskToPick(context.Background(), client, "some job", tt.options, "Test Level")
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tt.wantID {
				t.Errorf("expected id %d, got %d", tt.wantID, id)
			}
		})
	}
}
