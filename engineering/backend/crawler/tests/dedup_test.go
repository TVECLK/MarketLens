package crawler_test

import (
	"testing"

	"marketlens-go-backend/crawler"
	"marketlens-go-backend/crawler/mocks"
)

func TestGenerateMinHashAndLSH_SameInputSameOutput(t *testing.T) {
	job := crawler.JobData{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Build things"}

	sig1, lsh1 := crawler.GenerateMinHashAndLSH(job, 128, 8)
	sig2, lsh2 := crawler.GenerateMinHashAndLSH(job, 128, 8)

	if len(sig1) != 128 {
		t.Fatalf("expected signature length 128, got %d", len(sig1))
	}
	for i := range sig1 {
		if sig1[i] != sig2[i] {
			t.Fatalf("signature not deterministic at index %d: %d != %d", i, sig1[i], sig2[i])
		}
	}
	if len(lsh1) != 8 {
		t.Fatalf("expected 8 LSH bands, got %d", len(lsh1))
	}
	for i := range lsh1 {
		if lsh1[i].BucketKey != lsh2[i].BucketKey {
			t.Fatalf("bucket key not deterministic at band %d", i)
		}
	}
}

func TestGenerateMinHashAndLSH_DifferentJobsDifferentSignatures(t *testing.T) {
	jobA := crawler.JobData{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Build backend services"}
	jobB := crawler.JobData{Employer: "Globex", JobRole: "Store Manager", Location: "Kandy", Description: "Manage retail operations"}

	sigA, _ := crawler.GenerateMinHashAndLSH(jobA, 128, 8)
	sigB, _ := crawler.GenerateMinHashAndLSH(jobB, 128, 8)

	matches := 0
	for i := range sigA {
		if sigA[i] == sigB[i] {
			matches++
		}
	}
	if matches > 20 {
		t.Errorf("expected largely different signatures for unrelated jobs, got %d/128 matching positions", matches)
	}
}

func TestPQArrayRoundTrip(t *testing.T) {
	job := crawler.JobData{Employer: "Acme", JobRole: "Engineer", Location: "Colombo", Description: "Build things"}
	sig, _ := crawler.GenerateMinHashAndLSH(job, 128, 8)

	stored := crawler.ToPQArray(sig)
	recovered := crawler.FromPQArray(stored)

	if len(recovered) != len(sig) {
		t.Fatalf("length mismatch: got %d, want %d", len(recovered), len(sig))
	}
	for i := range sig {
		if sig[i] != recovered[i] {
			t.Fatalf("round-trip mismatch at index %d: got %d, want %d", i, recovered[i], sig[i])
		}
	}
}

func TestPQArrayFitsSignedInt32Range(t *testing.T) {
	// Every value FNV-32 can produce (0 to 2^32-1) must survive the
	// round-trip through int32 without silently changing meaning.
	extremes := []uint32{0, 1, 1 << 31, (1 << 32) - 1, 2147483647, 2147483648}
	stored := crawler.ToPQArray(extremes)
	recovered := crawler.FromPQArray(stored)
	for i, want := range extremes {
		if recovered[i] != want {
			t.Errorf("value %d: got %d after round-trip", want, recovered[i])
		}
	}
}

func TestCheckDuplicate_CatchesNearDuplicate(t *testing.T) {
	repo := mocks.NewMockRepository()

	original := crawler.JobData{
		Employer:    "Ceylon Software Solutions",
		JobRole:     "Senior Backend Engineer",
		Location:    "Colombo 03, Sri Lanka",
		Description: "Go, PostgreSQL and Docker experience required for our growing backend team.",
	}
	sig, lsh := crawler.GenerateMinHashAndLSH(original, 128, 8)
	bucketKeys := make([]string, len(lsh))
	for i, l := range lsh {
		bucketKeys[i] = l.BucketKey
	}
	repo.SeedJob(1, original.Location, sig, bucketKeys)

	reworded := crawler.JobData{
		Employer:    "Ceylon Software Solutions",
		JobRole:     "Senior Backend Engineer",
		Location:    "Colombo 03",
		Description: "Go, PostgreSQL, and Docker experience required for our growing backend team!",
	}
	newSig, newLSH := crawler.GenerateMinHashAndLSH(reworded, 128, 8)

	isDuplicate, matchedID := crawler.CheckDuplicate(repo, newLSH, newSig, reworded.Location, 0.65)
	if !isDuplicate {
		t.Fatal("expected near-duplicate job to be caught")
	}
	if matchedID != 1 {
		t.Errorf("expected matched job ID 1, got %d", matchedID)
	}
}

func TestCheckDuplicate_LeavesUnrelatedJobsAlone(t *testing.T) {
	repo := mocks.NewMockRepository()

	original := crawler.JobData{
		Employer:    "Ceylon Software Solutions",
		JobRole:     "Senior Backend Engineer",
		Location:    "Colombo 03, Sri Lanka",
		Description: "Go, PostgreSQL and Docker experience required for our growing backend team.",
	}
	sig, lsh := crawler.GenerateMinHashAndLSH(original, 128, 8)
	bucketKeys := make([]string, len(lsh))
	for i, l := range lsh {
		bucketKeys[i] = l.BucketKey
	}
	repo.SeedJob(1, original.Location, sig, bucketKeys)

	different := crawler.JobData{
		Employer:    "Lanka Retail Group",
		JobRole:     "Store Manager",
		Location:    "Kandy, Sri Lanka",
		Description: "Manage daily operations, staff scheduling, and inventory.",
	}
	newSig, newLSH := crawler.GenerateMinHashAndLSH(different, 128, 8)

	isDuplicate, _ := crawler.CheckDuplicate(repo, newLSH, newSig, different.Location, 0.65)
	if isDuplicate {
		t.Fatal("expected unrelated job NOT to be flagged as a duplicate")
	}
}