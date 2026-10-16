package crawler_test

import (
	"context"
	"testing"

	"marketlens-go-backend/crawler"
	"marketlens-go-backend/crawler/mocks"
)

func TestIndustryClassifier_WalksToLeaf(t *testing.T) {
	repo := mocks.NewMockRepository()
	llm := mocks.NewMockDeepSeek(t, func(prompt string) string {
		id := mocks.FirstListedID(prompt)
		return `{"id": ` + itoa(id) + `}`
	})

	classifier := crawler.NewIndustryClassifier(repo, llm)

	id, err := classifier.Classify(context.Background(), "Backend engineer building Go microservices")
	if err != nil {
		t.Fatalf("Classify() error: %v", err)
	}
	if id != 62011 {
		t.Errorf("expected leaf industry_subclass_id 62011, got %d", id)
	}
}

func TestOccupationClassifier_WalksToLeaf(t *testing.T) {
	repo := mocks.NewMockRepository()
	llm := mocks.NewMockDeepSeek(t, func(prompt string) string {
		id := mocks.FirstListedID(prompt)
		return `{"id": ` + itoa(id) + `}`
	})

	classifier := crawler.NewOccupationClassifier(repo, llm)

	id, err := classifier.Classify(context.Background(), "Backend engineer building Go microservices")
	if err != nil {
		t.Fatalf("Classify() error: %v", err)
	}
	if id != 25121 {
		t.Errorf("expected leaf occupation_group_id 25121, got %d", id)
	}
}

func TestAskToPick_RejectsHallucinatedIDNotInOptions(t *testing.T) {
	// The mock always answers with a fabricated id that was never
	// among the options offered, regardless of what's actually asked.
	llm := mocks.NewMockDeepSeek(t, func(prompt string) string {
		return `{"id": 999}`
	})

	options := []crawler.Option{{ID: 1, Name: "First"}, {ID: 2, Name: "Second"}}
	id, err := llm.AskToPick(context.Background(), "some job", options, "Test Level")
	if err != nil {
		t.Fatalf("AskToPick returned an error instead of treating the hallucination as no-match: %v", err)
	}
	if id != 0 {
		t.Errorf("expected hallucinated id 999 to be rejected (id=0), got %d", id)
	}
}

func TestAskToPick_AcceptsIDActuallyInOptions(t *testing.T) {
	llm := mocks.NewMockDeepSeek(t, func(prompt string) string {
		return `{"id": 2}`
	})

	options := []crawler.Option{{ID: 1, Name: "First"}, {ID: 2, Name: "Second"}}
	id, err := llm.AskToPick(context.Background(), "some job", options, "Test Level")
	if err != nil {
		t.Fatalf("AskToPick failed on a genuinely valid pick: %v", err)
	}
	if id != 2 {
		t.Errorf("expected valid id 2 to be accepted, got %d", id)
	}
}

func TestAskToPick_HandlesMarkdownFencedResponse(t *testing.T) {
	llm := mocks.NewMockDeepSeek(t, func(prompt string) string {
		return "```json\n{\"id\": 7}\n```"
	})

	id, err := llm.AskToPick(context.Background(), "some job", []crawler.Option{{ID: 7, Name: "Match"}}, "Test Level")
	if err != nil {
		t.Fatalf("AskToPick failed on fenced response: %v", err)
	}
	if id != 7 {
		t.Errorf("expected id 7, got %d", id)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}