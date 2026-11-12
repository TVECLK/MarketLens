package ingestion_test

import (
	"context"
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
			sigA, lshA := ingestion.GenerateMinHashAndLSH(tt.a, numPerm, numBands)
			sigB, lshB := ingestion.GenerateMinHashAndLSH(tt.b, numPerm, numBands)

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

func TestPQArrayRoundTrip(t *testing.T) {
	generated, _ := ingestion.GenerateMinHashAndLSH(
		ingestion.JobData{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Build things"},
		numPerm, numBands,
	)

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

			sig, lsh := ingestion.GenerateMinHashAndLSH(original, numPerm, numBands)
			bucketKeys := make([]string, len(lsh))
			for i, l := range lsh {
				bucketKeys[i] = l.BucketKey
			}
			seededID := seedJob(t, db, original.Location, sig, bucketKeys)

			newSig, newLSH := ingestion.GenerateMinHashAndLSH(tt.incoming, numPerm, numBands)
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
