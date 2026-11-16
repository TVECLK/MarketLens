package ingestion_test

import (
	"context"
	"errors"
	"testing"

	"marketlens-go-backend/internal/ingestion"
)

const (
	numPerm  = 128
	numBands = 8
)

func TestGenerateMinHashAndLSH(t *testing.T) {
	acme := ingestion.JobData{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Build things"}

	tests := []struct {
		name string
		a, b ingestion.JobData
		// Bounds on how many of the 128 signature positions the two jobs
		// share.
		minMatches, maxMatches int
		wantSameBuckets        bool
	}{
		{
			name:            "same input gives the same signature and bucket keys",
			a:               acme,
			b:               acme,
			minMatches:      numPerm,
			maxMatches:      numPerm,
			wantSameBuckets: true,
		},
		{
			name:       "unrelated jobs give largely different signatures",
			a:          ingestion.JobData{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Build backend services"},
			b:          ingestion.JobData{Employer: "Globex", JobRole: "Store Manager", Location: "Kandy", Description: "Manage retail operations"},
			minMatches: 0,
			maxMatches: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sigA, lshA, err := ingestion.GenerateMinHashAndLSH(tt.a, numPerm, numBands)
			if err != nil {
				t.Fatalf("GenerateMinHashAndLSH(a) error: %v", err)
			}
			sigB, lshB, err := ingestion.GenerateMinHashAndLSH(tt.b, numPerm, numBands)
			if err != nil {
				t.Fatalf("GenerateMinHashAndLSH(b) error: %v", err)
			}

			if len(sigA) != numPerm {
				t.Fatalf("expected signature length %d, got %d", numPerm, len(sigA))
			}
			if len(lshA) != numBands {
				t.Fatalf("expected %d LSH bands, got %d", numBands, len(lshA))
			}

			matches := 0
			for i := range sigA {
				if sigA[i] == sigB[i] {
					matches++
				}
			}
			if matches < tt.minMatches || matches > tt.maxMatches {
				t.Errorf("expected between %d and %d matching positions, got %d", tt.minMatches, tt.maxMatches, matches)
			}

			sameBuckets := true
			for i := range lshA {
				if lshA[i].BucketKey != lshB[i].BucketKey {
					sameBuckets = false
				}
			}
			if sameBuckets != tt.wantSameBuckets {
				t.Errorf("expected same bucket keys=%t, got %t", tt.wantSameBuckets, sameBuckets)
			}
		})
	}
}

func TestGenerateMinHashAndLSH_TooLittleText(t *testing.T) {
	tests := []struct {
		name    string
		job     ingestion.JobData
		wantErr bool
	}{
		{"all fields empty", ingestion.JobData{}, true},
		{"only whitespace", ingestion.JobData{Employer: "  ", Description: "\n\t"}, true},
		{"one character", ingestion.JobData{Employer: "A"}, true},
		{"two characters", ingestion.JobData{Employer: "AB"}, true},
		{"two multi-byte runes", ingestion.JobData{Employer: "සි"}, true},
		{"exactly one shingle", ingestion.JobData{Employer: "abc"}, false},
		// The separator space between fields counts toward the length.
		{"one shingle from several fields", ingestion.JobData{Employer: "a", JobRole: "b"}, false},
		{"three multi-byte runes", ingestion.JobData{Employer: "සිං"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, lsh, err := ingestion.GenerateMinHashAndLSH(tt.job, numPerm, numBands)
			if tt.wantErr {
				if !errors.Is(err, ingestion.ErrTooLittleText) {
					t.Fatalf("expected ErrTooLittleText, got %v", err)
				}
				if sig != nil || lsh != nil {
					t.Errorf("expected nil signature and LSH on error, got sig=%v lsh=%v", sig, lsh)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(sig) != numPerm || len(lsh) != numBands {
				t.Errorf("expected %d perms and %d bands, got %d and %d", numPerm, numBands, len(sig), len(lsh))
			}
		})
	}
}

func TestPQArrayRoundTrip(t *testing.T) {
	generated, _, err := ingestion.GenerateMinHashAndLSH(
		ingestion.JobData{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Build things"},
		numPerm, numBands,
	)
	if err != nil {
		t.Fatalf("GenerateMinHashAndLSH() error: %v", err)
	}

	tests := []struct {
		name string
		sig  []uint32
	}{
		{"generated signature", generated},
		// Every value FNV-32 can produce (0 to 2^32-1) must survive the
		// round-trip through int32 without silently changing meaning.
		{"values at the int32 boundaries", []uint32{0, 1, 1 << 31, (1 << 32) - 1, 2147483647, 2147483648}},
		{"empty signature", []uint32{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recovered := ingestion.FromPQArray(ingestion.ToPQArray(tt.sig))

			if len(recovered) != len(tt.sig) {
				t.Fatalf("length mismatch: got %d, want %d", len(recovered), len(tt.sig))
			}
			for i, want := range tt.sig {
				if recovered[i] != want {
					t.Errorf("index %d: got %d after round-trip, want %d", i, recovered[i], want)
				}
			}
		})
	}
}

func TestCheckDuplicate(t *testing.T) {
	original := ingestion.JobData{
		Employer:    "Ceylon Software Solutions",
		JobRole:     "Senior Backend Engineer",
		Location:    "Colombo 03, Sri Lanka",
		Description: "Go, PostgreSQL and Docker experience required for our growing backend team.",
	}

	tests := []struct {
		name          string
		incoming      ingestion.JobData
		wantDuplicate bool
	}{
		{
			name: "reworded repost is caught as a near-duplicate",
			incoming: ingestion.JobData{
				Employer:    "Ceylon Software Solutions",
				JobRole:     "Senior Backend Engineer",
				Location:    "Colombo 03",
				Description: "Go, PostgreSQL, and Docker experience required for our growing backend team!",
			},
			wantDuplicate: true,
		},
		{
			name: "unrelated job is left alone",
			incoming: ingestion.JobData{
				Employer:    "Lanka Retail Group",
				JobRole:     "Store Manager",
				Location:    "Kandy, Sri Lanka",
				Description: "Manage daily operations, staff scheduling, and inventory.",
			},
			wantDuplicate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupDB(t)
			svcs := newServices(db)

			sig, lsh, err := ingestion.GenerateMinHashAndLSH(original, numPerm, numBands)
			if err != nil {
				t.Fatalf("GenerateMinHashAndLSH(original) error: %v", err)
			}
			bucketKeys := make([]string, len(lsh))
			for i, l := range lsh {
				bucketKeys[i] = l.BucketKey
			}
			seededID := seedJob(t, db, original.Location, sig, bucketKeys)

			newSig, newLSH, err := ingestion.GenerateMinHashAndLSH(tt.incoming, numPerm, numBands)
			if err != nil {
				t.Fatalf("GenerateMinHashAndLSH(incoming) error: %v", err)
			}
			isDuplicate, matchedID, err := ingestion.CheckDuplicate(context.Background(), svcs.JobPosts, newLSH, newSig, tt.incoming.Location, 0.65)
			if err != nil {
				t.Fatalf("CheckDuplicate() error: %v", err)
			}
			if isDuplicate != tt.wantDuplicate {
				t.Fatalf("expected duplicate=%t, got %t", tt.wantDuplicate, isDuplicate)
			}
			if tt.wantDuplicate && matchedID != int64(seededID) {
				t.Errorf("expected matched job ID %d, got %d", seededID, matchedID)
			}
			if !tt.wantDuplicate && matchedID != 0 {
				t.Errorf("expected no matched job ID, got %d", matchedID)
			}
		})
	}
}
