package crawler_test

import (
	"strings"
	"testing"
	"time"

	"marketlens-go-backend/crawler"
	"marketlens-go-backend/crawler/mocks"
)

func TestMetadataBuilder_ProducesSchemaAndInstruction(t *testing.T) {
	repo := mocks.NewMockRepository()
	b := crawler.NewMetadataBuilder(repo, time.Minute)

	schema, instruction, err := b.Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}

	if schema["type"] != "object" {
		t.Errorf("expected schema type=object, got %v", schema["type"])
	}
	if !strings.Contains(instruction, "1: Formal") {
		t.Errorf("expected instruction to contain formatted formality option, got:\n%s", instruction)
	}
	if !strings.Contains(instruction, "2: Female") {
		t.Errorf("expected instruction to contain formatted gender option, got:\n%s", instruction)
	}
	if !strings.Contains(instruction, "3: Bachelor's Degree") {
		t.Errorf("expected instruction to contain formatted education level option, got:\n%s", instruction)
	}
}

func TestMetadataBuilder_CachesWithinTTL(t *testing.T) {
	repo := mocks.NewMockRepository()
	b := crawler.NewMetadataBuilder(repo, 50*time.Millisecond)

	if _, _, err := b.Build(); err != nil {
		t.Fatalf("first Build() error: %v", err)
	}
	if _, _, err := b.Build(); err != nil {
		t.Fatalf("second Build() error: %v", err)
	}

	if got := repo.FetchCounts["formalities"]; got != 1 {
		t.Errorf("expected 1 DB fetch within TTL window, got %d", got)
	}

	time.Sleep(60 * time.Millisecond)

	if _, _, err := b.Build(); err != nil {
		t.Fatalf("third Build() error: %v", err)
	}

	if got := repo.FetchCounts["formalities"]; got != 2 {
		t.Errorf("expected a second DB fetch after TTL expiry, got %d", got)
	}
}