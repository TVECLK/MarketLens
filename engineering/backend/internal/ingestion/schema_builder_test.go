package ingestion_test

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMetadataBuilder_Build(t *testing.T) {
	// Each case is one reference list that must show up in the
	// instruction as an "id: label" line.
	tests := []struct {
		name         string
		wantContains string
	}{
		{"formality option", "1: Formal"},
		{"gender option", "2: Female"},
		{"vocational education option", "1: NVQ Level 4"},
		{"employment sector option", "2: Government"},
		{"education level option", "3: Bachelor's Degree"},
		{"experience option", "2: 1-2 Years"},
	}

	b := newMetadataBuilder(newServices(setupDB(t)), time.Minute)
	schema, instruction, _, err := b.Build(context.Background())
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if schema["type"] != "object" {
		t.Errorf("expected schema type=object, got %v", schema["type"])
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(instruction, tt.wantContains) {
				t.Errorf("expected instruction to contain %q, got:\n%s", tt.wantContains, instruction)
			}
		})
	}
}

func TestMetadataBuilder_CachesWithinTTL(t *testing.T) {
	db := setupDB(t)
	formalityQueries := countQueries(t, db, "formality")
	b := newMetadataBuilder(newServices(db), time.Minute)
	ctx := context.Background()

	if _, _, _, err := b.Build(ctx); err != nil {
		t.Fatalf("first Build() error: %v", err)
	}
	// A refresh may run more than one query (e.g. count + list), so measure
	// what one refresh costs rather than hardcoding it.
	perRefresh := formalityQueries.Load()
	if perRefresh == 0 {
		t.Fatal("expected the first Build() to query formalities")
	}

	if _, _, _, err := b.Build(ctx); err != nil {
		t.Fatalf("second Build() error: %v", err)
	}
	if got := formalityQueries.Load(); got != perRefresh {
		t.Errorf("expected no extra queries within the TTL window, got %d (was %d)", got, perRefresh)
	}
}

func TestMetadataBuilder_RefreshesAfterTTL(t *testing.T) {
	db := setupDB(t)
	formalityQueries := countQueries(t, db, "formality")
	b := newMetadataBuilder(newServices(db), 50*time.Millisecond)
	ctx := context.Background()

	if _, _, _, err := b.Build(ctx); err != nil {
		t.Fatalf("first Build() error: %v", err)
	}
	perRefresh := formalityQueries.Load()
	if perRefresh == 0 {
		t.Fatal("expected the first Build() to query formalities")
	}

	// Sleeping past the TTL can only make expiry more certain, so a slow
	// runner can't make this test flake.
	time.Sleep(60 * time.Millisecond)

	if _, _, _, err := b.Build(ctx); err != nil {
		t.Fatalf("second Build() error: %v", err)
	}
	if got := formalityQueries.Load(); got != 2*perRefresh {
		t.Errorf("expected a second refresh after TTL expiry (%d queries), got %d", 2*perRefresh, got)
	}
}
