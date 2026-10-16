package crawler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"marketlens-go-backend/crawler"
	"marketlens-go-backend/crawler/mocks"
	"marketlens-go-backend/models"
	"marketlens-go-backend/repositories"
)

// extractionResponder returns a full, valid extraction JSON for any
// prompt that looks like the extraction call (starts with the
// instruction template's opening line), and a classifier pick
// (mocks.FirstListedID) for anything else.
func extractionResponder(content string) mocks.Responder {
	return func(prompt string) string {
		if strings.HasPrefix(prompt, "Extract structural data from") {
			return content
		}
		id := mocks.FirstListedID(prompt)
		return `{"id": ` + itoa(id) + `}`
	}
}

func TestProcessBatch_DuplicateAndNewJobTogether(t *testing.T) {
	repo := mocks.NewMockRepository()
	llm := mocks.NewMockDeepSeek(t, extractionResponder(`{
		"job_type": {"type": "Full Time"},
		"work_mode": "onsite",
		"no_of_vacancies": 2,
		"meta_data": {
			"geo_data": {"province": "Western"},
			"posted_at": "2026-09-01T00:00:00Z",
			"confidence_score": 0.92,
			"formality_id": 1,
			"gender_id": 3,
			"vocational_education_id": 1,
			"employment_sector_id": 1,
			"education_level_id": 3,
			"experience_id": 2
		},
		"skills": [{"skill": "Go"}, {"skill": "PostgreSQL"}]
	}`))
	metaBuilder := crawler.NewMetadataBuilder(repo, 0)
	svc := crawler.NewIngestionService(repo, llm, metaBuilder, 0.65)

	firstJob := crawler.RawJobInput{
		Employer:     "Ceylon Software Solutions",
		JobRole:      "Senior Backend Engineer",
		Location:     "Colombo 03, Sri Lanka",
		Description:  "Go, PostgreSQL and Docker experience required for our growing backend team.",
		CrawlerRunID: 100,
		Source:       "Ikman",
	}

	// Seed the store with firstJob already saved, as if a previous
	// crawl run had inserted it.
	if failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{firstJob}); len(failed) != 0 {
		t.Fatalf("seeding first batch failed: %+v", failed)
	}

	duplicateJob := crawler.RawJobInput{
		Employer:     "Ceylon Software Solutions",
		JobRole:      "Senior Backend Engineer",
		Location:     "Colombo 03",
		Description:  "Go, PostgreSQL, and Docker experience required for our growing backend team!",
		CrawlerRunID: 200,
		Source:       "Ikman",
	}
	newJob := crawler.RawJobInput{
		Employer:     "Lanka Retail Group",
		JobRole:      "Store Manager",
		Location:     "Kandy, Sri Lanka",
		Description:  "Manage daily operations, staff scheduling, and inventory for our flagship retail outlet.",
		CrawlerRunID: 200,
		Source:       "TopJobs",
	}

	if failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{duplicateJob, newJob}); len(failed) != 0 {
		t.Fatalf("ProcessBatch failed: %+v", failed)
	}

	// The duplicate must trigger exactly one update, not a new save.
	if len(repo.UpdatedDuplicates) != 1 {
		t.Fatalf("expected 1 duplicate update, got %d", len(repo.UpdatedDuplicates))
	}
	if repo.UpdatedDuplicates[0].JobPostID != 1 {
		t.Errorf("expected duplicate update to target job ID 1, got %d", repo.UpdatedDuplicates[0].JobPostID)
	}
	if repo.UpdatedDuplicates[0].CrawlerRunID == nil || *repo.UpdatedDuplicates[0].CrawlerRunID != 200 {
		t.Errorf("expected duplicate's CrawlerRunID updated to 200, got %v", repo.UpdatedDuplicates[0].CrawlerRunID)
	}

	// The new job must trigger exactly one additional save (on top of
	// the first job saved during setup), fully extracted and
	// classified. Find it by location rather than assuming index 0,
	// since SavedJobs accumulates everything saved across both calls.
	if len(repo.SavedJobs) != 2 {
		t.Fatalf("expected 2 total jobs saved (1 setup + 1 new), got %d", len(repo.SavedJobs))
	}
	var saved *models.JobPost
	for i := range repo.SavedJobs {
		if repo.SavedJobs[i].Location == "Kandy, Sri Lanka" {
			saved = &repo.SavedJobs[i]
		}
	}
	if saved == nil {
		t.Fatal("new job (Kandy, Sri Lanka) not found among saved jobs")
	}
	if saved.JobType == nil || saved.JobType.Type != "Full Time" {
		t.Errorf("expected extracted job_type Full Time, got %+v", saved.JobType)
	}
	if len(saved.Skills) != 2 {
		t.Errorf("expected 2 extracted skills, got %d", len(saved.Skills))
	}
	if saved.MetaData.IndustrySubclassID == nil || *saved.MetaData.IndustrySubclassID == 0 {
		t.Errorf("expected a non-zero classified industry subclass id, got %v", saved.MetaData.IndustrySubclassID)
	}
	if saved.MetaData.OccupationGroupID == nil || *saved.MetaData.OccupationGroupID == 0 {
		t.Errorf("expected a non-zero classified occupation group id, got %v", saved.MetaData.OccupationGroupID)
	}
	if saved.MetaData.MinhashSignature == nil {
		t.Error("expected a stored MinHash signature on the new job")
	}
}

func TestProcessBatch_OmittedFieldsBecomeNilNotFKBreakingZero(t *testing.T) {
	repo := mocks.NewMockRepository()
	// Deliberately omit work_mode, formality_id, gender_id from the
	// extraction response, and make the classifier picks return 0 (no
	// match) by using an id that isn't in any offered option list.
	llm := mocks.NewMockDeepSeek(t, func(prompt string) string {
		if strings.HasPrefix(prompt, "Extract structural data from") {
			return `{
				"job_type": {"type": "Full Time"},
				"no_of_vacancies": 1,
				"meta_data": {
					"geo_data": {"province": "Western"},
					"posted_at": "2026-09-01T00:00:00Z",
					"confidence_score": 0.5
				},
				"skills": []
			}`
		}
		return `{"id": 0}`
	})
	metaBuilder := crawler.NewMetadataBuilder(repo, 0)
	svc := crawler.NewIngestionService(repo, llm, metaBuilder, 0.65)

	raw := crawler.RawJobInput{
		Employer:     "Test Co",
		JobRole:      "Tester",
		Location:     "Colombo",
		Description:  "Testing incomplete extraction handling",
		CrawlerRunID: 1,
		Source:       "Test",
	}

	if failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{raw}); len(failed) != 0 {
		t.Fatalf("ProcessBatch failed: %+v", failed)
	}

	if len(repo.SavedJobs) != 1 {
		t.Fatalf("expected 1 job saved, got %d", len(repo.SavedJobs))
	}
	saved := repo.SavedJobs[0]

	if saved.MetaData.FormalityID != nil {
		t.Errorf("expected FormalityID nil (omitted by LLM), got pointer to %d", *saved.MetaData.FormalityID)
	}
	if saved.MetaData.GenderID != nil {
		t.Errorf("expected GenderID nil (omitted by LLM), got pointer to %d", *saved.MetaData.GenderID)
	}
	if saved.MetaData.IndustrySubclassID != nil {
		t.Errorf("expected IndustrySubclassID nil (no classification match), got pointer to %d", *saved.MetaData.IndustrySubclassID)
	}
	if saved.MetaData.OccupationGroupID != nil {
		t.Errorf("expected OccupationGroupID nil (no classification match), got pointer to %d", *saved.MetaData.OccupationGroupID)
	}
	if saved.WorkMode != "onsite" {
		t.Errorf("expected WorkMode to default to 'onsite' when omitted, got %q", saved.WorkMode)
	}
}

func TestProcessBatch_HandlesMarkdownFencedExtractionResponse(t *testing.T) {
	repo := mocks.NewMockRepository()
	fenced := "```json\n" + `{
		"job_type": {"type": "Part Time"},
		"work_mode": "remote",
		"no_of_vacancies": 1,
		"meta_data": {
			"geo_data": {"province": "Western"},
			"posted_at": "2026-09-01T00:00:00Z",
			"confidence_score": 0.8,
			"formality_id": 1,
			"gender_id": 3,
			"vocational_education_id": 1,
			"employment_sector_id": 1,
			"education_level_id": 3,
			"experience_id": 2
		},
		"skills": [{"skill": "Python"}]
	}` + "\n```"

	llm := mocks.NewMockDeepSeek(t, extractionResponder(fenced))
	metaBuilder := crawler.NewMetadataBuilder(repo, 0)
	svc := crawler.NewIngestionService(repo, llm, metaBuilder, 0.65)

	raw := crawler.RawJobInput{
		Employer:     "Fence Test Co",
		JobRole:      "Remote Analyst",
		Location:     "Colombo",
		Description:  "A job whose extraction response is wrapped in a markdown code fence.",
		CrawlerRunID: 1,
		Source:       "Test",
	}

	if failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{raw}); len(failed) != 0 {
		t.Fatalf("ProcessBatch failed on fenced extraction response: %+v", failed)
	}

	if len(repo.SavedJobs) != 1 {
		t.Fatalf("expected 1 job saved, got %d", len(repo.SavedJobs))
	}
	if repo.SavedJobs[0].WorkMode != "remote" {
		t.Errorf("expected WorkMode 'remote' parsed correctly from fenced response, got %q", repo.SavedJobs[0].WorkMode)
	}
}

// validExtractionJSON is a complete, valid extraction response reused by
// tests that don't care about the extracted content itself, only about
// what happens around it (failure handling, status codes, etc.).
const validExtractionJSON = `{
	"job_type": {"type": "Full Time"},
	"work_mode": "onsite",
	"no_of_vacancies": 1,
	"meta_data": {
		"geo_data": {"province": "Western"},
		"posted_at": "2026-09-01T00:00:00Z",
		"confidence_score": 0.9,
		"formality_id": 1,
		"gender_id": 3,
		"vocational_education_id": 1,
		"employment_sector_id": 1,
		"education_level_id": 3,
		"experience_id": 2
	},
	"skills": [{"skill": "Go"}]
}`

// TestProcessBatch_ContinuesAfterExtractionFailure proves the CRITICAL
// batch-processing requirement: one job's extraction failure must not
// abort the rest of the batch. job-2's extraction response is
// deliberately garbage; job-1 and job-3 must still be saved.
func TestProcessBatch_ContinuesAfterExtractionFailure(t *testing.T) {
	repo := mocks.NewMockRepository()
	llm := mocks.NewMockDeepSeek(t, func(prompt string) string {
		if strings.HasPrefix(prompt, "Extract structural data from") {
			if strings.Contains(prompt, "BadCo") {
				return "not valid json"
			}
			return validExtractionJSON
		}
		id := mocks.FirstListedID(prompt)
		return `{"id": ` + itoa(id) + `}`
	})
	metaBuilder := crawler.NewMetadataBuilder(repo, 0)
	svc := crawler.NewIngestionService(repo, llm, metaBuilder, 0.65)

	jobs := []crawler.RawJobInput{
		{
			JobID:        "job-1",
			Employer:     "GoodCo Alpha",
			JobRole:      "Engineer",
			Location:     "Colombo",
			Description:  "A perfectly normal backend engineering role.",
			CrawlerRunID: 1,
			Source:       "Test",
		},
		{
			JobID:        "job-2",
			Employer:     "BadCo",
			JobRole:      "Analyst",
			Location:     "Kandy",
			Description:  "A role whose extraction response will come back as garbage.",
			CrawlerRunID: 1,
			Source:       "Test",
		},
		{
			JobID:        "job-3",
			Employer:     "GoodCo Beta",
			JobRole:      "Designer",
			Location:     "Galle",
			Description:  "A completely unrelated design role at a different company.",
			CrawlerRunID: 1,
			Source:       "Test",
		},
	}

	failed := svc.ProcessBatch(context.Background(), jobs)

	if len(failed) != 1 {
		t.Fatalf("expected exactly 1 failure, got %d: %+v", len(failed), failed)
	}
	if failed[0].JobID != "job-2" {
		t.Errorf("expected the failure to be job-2, got %q", failed[0].JobID)
	}
	if failed[0].StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected malformed extraction content to be transient (503), got %d", failed[0].StatusCode)
	}

	// job-1 and job-3 must still have been saved — one bad job must not
	// stop the rest of the batch from being processed.
	if len(repo.SavedJobs) != 2 {
		t.Fatalf("expected 2 jobs saved (job-1 and job-3), got %d", len(repo.SavedJobs))
	}
	for _, saved := range repo.SavedJobs {
		if saved.Location == "Kandy" {
			t.Error("job-2 should not have been saved, but a Kandy-located job was found among SavedJobs")
		}
	}
}

// TestProcessBatch_SaveFailure_PermanentReturns422 proves that a SaveOneJob
// error wrapping repositories.ErrPermanentSaveFailure is reported as 422
// (don't retry).
func TestProcessBatch_SaveFailure_PermanentReturns422(t *testing.T) {
	repo := mocks.NewMockRepository()
	repo.SaveOneJobErr = func(job *models.JobPost) error {
		return fmt.Errorf("%w: test-forced permanent failure", repositories.ErrPermanentSaveFailure)
	}

	llm := mocks.NewMockDeepSeek(t, extractionResponder(validExtractionJSON))
	metaBuilder := crawler.NewMetadataBuilder(repo, 0)
	svc := crawler.NewIngestionService(repo, llm, metaBuilder, 0.65)

	raw := crawler.RawJobInput{
		JobID:        "job-perm",
		Employer:     "Permanent Fail Co",
		JobRole:      "Tester",
		Location:     "Colombo",
		Description:  "A job whose save will fail permanently.",
		CrawlerRunID: 1,
		Source:       "Test",
	}

	failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{raw})

	if len(failed) != 1 {
		t.Fatalf("expected exactly 1 failure, got %d: %+v", len(failed), failed)
	}
	if failed[0].JobID != "job-perm" {
		t.Errorf("expected failure JobID 'job-perm', got %q", failed[0].JobID)
	}
	if failed[0].StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected ErrPermanentSaveFailure to map to 422, got %d", failed[0].StatusCode)
	}
	if len(repo.SavedJobs) != 0 {
		t.Errorf("expected no jobs saved, got %d", len(repo.SavedJobs))
	}
}

// TestProcessBatch_SaveFailure_TransientReturns503 proves that any
// SaveOneJob error NOT wrapping ErrPermanentSaveFailure is reported as
// 503 (retryable).
func TestProcessBatch_SaveFailure_TransientReturns503(t *testing.T) {
	repo := mocks.NewMockRepository()
	repo.SaveOneJobErr = func(job *models.JobPost) error {
		return errors.New("connection reset by peer")
	}

	llm := mocks.NewMockDeepSeek(t, extractionResponder(validExtractionJSON))
	metaBuilder := crawler.NewMetadataBuilder(repo, 0)
	svc := crawler.NewIngestionService(repo, llm, metaBuilder, 0.65)

	raw := crawler.RawJobInput{
		JobID:        "job-transient",
		Employer:     "Transient Fail Co",
		JobRole:      "Tester",
		Location:     "Colombo",
		Description:  "A job whose save will fail transiently.",
		CrawlerRunID: 1,
		Source:       "Test",
	}

	failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{raw})

	if len(failed) != 1 {
		t.Fatalf("expected exactly 1 failure, got %d: %+v", len(failed), failed)
	}
	if failed[0].JobID != "job-transient" {
		t.Errorf("expected failure JobID 'job-transient', got %q", failed[0].JobID)
	}
	if failed[0].StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected a plain save error to map to 503, got %d", failed[0].StatusCode)
	}
}

// TestProcessBatch_DuplicateUpdateFailure_Returns503 proves that an
// UpdateDuplicateJob error is reported as 503 (retryable) — there's no
// permanent/transient split on this path since it's just a single-column
// update against an already-known-valid row.
func TestProcessBatch_DuplicateUpdateFailure_Returns503(t *testing.T) {
	repo := mocks.NewMockRepository()
	llm := mocks.NewMockDeepSeek(t, extractionResponder(validExtractionJSON))
	metaBuilder := crawler.NewMetadataBuilder(repo, 0)
	svc := crawler.NewIngestionService(repo, llm, metaBuilder, 0.65)

	seedJob := crawler.RawJobInput{
		JobID:        "seed-1",
		Employer:     "Ceylon Software Solutions",
		JobRole:      "Senior Backend Engineer",
		Location:     "Colombo 03, Sri Lanka",
		Description:  "Go, PostgreSQL and Docker experience required for our growing backend team.",
		CrawlerRunID: 100,
		Source:       "Ikman",
	}
	if failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{seedJob}); len(failed) != 0 {
		t.Fatalf("seeding failed: %+v", failed)
	}

	// Make the duplicate-update path fail, now that the seed job exists.
	repo.UpdateDuplicateErr = func(jobPostID uint) error {
		return errors.New("db timeout")
	}

	duplicateJob := crawler.RawJobInput{
		JobID:        "dup-1",
		Employer:     "Ceylon Software Solutions",
		JobRole:      "Senior Backend Engineer",
		Location:     "Colombo 03",
		Description:  "Go, PostgreSQL, and Docker experience required for our growing backend team!",
		CrawlerRunID: 200,
		Source:       "Ikman",
	}

	failed := svc.ProcessBatch(context.Background(), []crawler.RawJobInput{duplicateJob})

	if len(failed) != 1 {
		t.Fatalf("expected exactly 1 failure, got %d: %+v", len(failed), failed)
	}
	if failed[0].JobID != "dup-1" {
		t.Errorf("expected failure JobID 'dup-1', got %q", failed[0].JobID)
	}
	if failed[0].StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected a duplicate-update failure to map to 503, got %d", failed[0].StatusCode)
	}
}