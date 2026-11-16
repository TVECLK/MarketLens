package ingestion_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/ingestion"
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

// TestMetadataBuilder_CallerGivesUpDuringRefresh checks that a caller whose
// context ends while the reference-data refresh is still running gets a
// transient error (a 503 the crawler retries), not a bare context error
// that ProcessBatch would report as the job's own fault (422).
func TestMetadataBuilder_CallerGivesUpDuringRefresh(t *testing.T) {
	tests := []struct {
		name    string
		newCtx  func() (context.Context, context.CancelFunc)
		giveUp  func(cancel context.CancelFunc) // ends the context once the refresh is running
		wantCtx error
	}{
		{
			name:    "cancelled",
			newCtx:  func() (context.Context, context.CancelFunc) { return context.WithCancel(context.Background()) },
			giveUp:  func(cancel context.CancelFunc) { cancel() },
			wantCtx: context.Canceled,
		},
		{
			name: "deadline exceeded",
			newCtx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 50*time.Millisecond)
			},
			giveUp:  func(context.CancelFunc) {}, // the deadline ends it on its own
			wantCtx: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupDB(t)
			started, release := blockQueries(t, db, "formality")
			b := newMetadataBuilder(newServices(db), time.Minute)

			ctx, cancel := tt.newCtx()
			defer cancel()

			errCh := make(chan error, 1)
			go func() {
				_, _, _, err := b.Build(ctx)
				errCh <- err
			}()

			// Wait until the refresh is genuinely in flight, so the cache
			// is cold and Build can only return by giving up on the wait.
			select {
			case <-started:
			case <-time.After(5 * time.Second):
				t.Fatal("refresh never started")
			}
			tt.giveUp(cancel)

			var err error
			select {
			case err = <-errCh:
			case <-time.After(5 * time.Second):
				t.Fatal("Build() did not return after its context ended")
			}
			release()

			if !errors.Is(err, ingestion.ErrTransientExtractionFailure) {
				t.Errorf("expected a transient failure, got %v", err)
			}
			if !errors.Is(err, tt.wantCtx) {
				t.Errorf("expected the context error %v to stay in the chain, got %v", tt.wantCtx, err)
			}
		})
	}
}

// blockQueries makes every SELECT GORM runs against table wait until
// release is called, so a test can hold a refresh in flight. started is
// closed when the first such query begins. release is safe to call more
// than once and is also run at cleanup so no goroutine is left blocked.
func blockQueries(t *testing.T, db *gorm.DB, table string) (started <-chan struct{}, release func()) {
	t.Helper()

	startedCh := make(chan struct{})
	releaseCh := make(chan struct{})
	var startOnce, releaseOnce sync.Once

	err := db.Callback().Query().Before("gorm:query").Register("test:block:"+table, func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			startOnce.Do(func() { close(startedCh) })
			<-releaseCh
		}
	})
	if err != nil {
		t.Fatalf("registering query callback: %v", err)
	}

	release = func() { releaseOnce.Do(func() { close(releaseCh) }) }
	t.Cleanup(release)
	return startedCh, release
}
