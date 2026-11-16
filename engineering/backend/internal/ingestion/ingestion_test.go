package ingestion_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/ingestion/mocks"
	"marketlens-go-backend/internal/models"
)

// loadJobs returns every saved job, oldest first, with the relations the
// assertions read.
func loadJobs(t *testing.T, db *gorm.DB) []models.JobPost {
	t.Helper()

	var jobs []models.JobPost
	if err := db.Preload("MetaData").Preload("JobType").Preload("Skills").Order("id ASC").Find(&jobs).Error; err != nil {
		t.Fatalf("loading saved jobs: %v", err)
	}
	return jobs
}

// extractionResponder returns content for any prompt that looks like the
// extraction call (starts with the instruction template's opening line),
// and a classifier pick (mocks.FirstListedID) for anything else.
func extractionResponder(content string) mocks.Responder {
	return func(prompt string) string {
		if strings.HasPrefix(prompt, "Extract structural data from") {
			return content
		}
		return `{"id": ` + itoa(mocks.FirstListedID(prompt)) + `}`
	}
}

// validExtractionJSON is a complete, valid extraction response reused by
// tests that don't care about the extracted content itself.
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

func rawJob(jobID, employer, role, location, description, source string, crawlerRunID uint) ingestion.RawJobInput {
	return ingestion.RawJobInput{
		JobID:        jobID,
		Employer:     employer,
		JobRole:      role,
		Location:     location,
		Description:  description,
		CrawlerRunID: crawlerRunID,
		Source:       source,
	}
}

// ceylonJob is the job several cases ingest first, so a reworded repost
// of it (ceylonRepost) is detected as a duplicate.
var (
	ceylonJob = rawJob("seed-1", "Ceylon Software Solutions", "Senior Backend Engineer", "Colombo 03, Sri Lanka",
		"Go, PostgreSQL and Docker experience required for our growing backend team.", "Ikman", 100)
	ceylonRepost = rawJob("dup-1", "Ceylon Software Solutions", "Senior Backend Engineer", "Colombo 03",
		"Go, PostgreSQL, and Docker experience required for our growing backend team!", "Ikman", 200)
)

// TestProcessBatch_Succeeds covers batches where every job is handled
// without a failure: the saved rows (and what was left unsaved) are what
// differ between cases.
func TestProcessBatch_Succeeds(t *testing.T) {
	tests := []struct {
		name    string
		respond mocks.Responder
		seed    []ingestion.RawJobInput // ingested first, to act as "already saved by a previous crawl"
		batch   []ingestion.RawJobInput
		// wantJobs is the total number of job rows after seed + batch.
		wantJobs int
		check    func(t *testing.T, jobs []models.JobPost)
	}{
		{
			name: "duplicate updates the existing job while a new job is extracted and saved",
			respond: extractionResponder(`{
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
			}`),
			seed: []ingestion.RawJobInput{ceylonJob},
			batch: []ingestion.RawJobInput{
				ceylonRepost,
				rawJob("new-1", "Lanka Retail Group", "Store Manager", "Kandy, Sri Lanka",
					"Manage daily operations, staff scheduling, and inventory for our flagship retail outlet.", "TopJobs", 200),
			},
			wantJobs: 2, // the setup job + the new one; the repost adds nothing
			check: func(t *testing.T, jobs []models.JobPost) {
				// The duplicate re-stamps the original job with the new run.
				if run := jobs[0].MetaData.CrawlerRunID; run == nil || *run != 200 {
					t.Errorf("expected the duplicate to update the original job's CrawlerRunID to 200, got %v", run)
				}

				// The new job is fully extracted and classified.
				var saved *models.JobPost
				for i := range jobs {
					if jobs[i].Location == "Kandy, Sri Lanka" {
						saved = &jobs[i]
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
				if id := saved.MetaData.IndustrySubclassID; id == nil || *id != 62011 {
					t.Errorf("expected classified industry subclass id 62011, got %v", id)
				}
				if id := saved.MetaData.OccupationGroupID; id == nil || *id != 25121 {
					t.Errorf("expected classified occupation group id 25121, got %v", id)
				}
				if len(saved.MetaData.MinhashSignature) == 0 {
					t.Error("expected a stored MinHash signature on the new job")
				}
			},
		},
		{
			// Omitted ids and classifier misses must be stored as NULL, not as
			// a 0 that would break the foreign key.
			name: "omitted fields and classifier misses become nil, not a zero id",
			respond: func(prompt string) string {
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
			},
			batch: []ingestion.RawJobInput{
				rawJob("job-1", "Test Co", "Tester", "Colombo", "Testing incomplete extraction handling", "Test", 1),
			},
			wantJobs: 1,
			check: func(t *testing.T, jobs []models.JobPost) {
				meta := jobs[0].MetaData
				if meta.FormalityID != nil {
					t.Errorf("expected FormalityID nil, got pointer to %d", *meta.FormalityID)
				}
				if meta.GenderID != nil {
					t.Errorf("expected GenderID nil, got pointer to %d", *meta.GenderID)
				}
				if meta.IndustrySubclassID != nil {
					t.Errorf("expected IndustrySubclassID nil, got pointer to %d", *meta.IndustrySubclassID)
				}
				if meta.OccupationGroupID != nil {
					t.Errorf("expected OccupationGroupID nil, got pointer to %d", *meta.OccupationGroupID)
				}
				if jobs[0].WorkMode != models.WorkModeOnsite {
					t.Errorf("expected WorkMode to default to 'onsite' when omitted, got %q", jobs[0].WorkMode)
				}
			},
		},
		{
			name: "extraction response wrapped in a markdown code fence",
			respond: extractionResponder("```json\n" +
				strings.Replace(validExtractionJSON, `"onsite"`, `"remote"`, 1) + "\n```"),
			batch: []ingestion.RawJobInput{
				rawJob("job-1", "Fence Test Co", "Remote Analyst", "Colombo", "A job whose extraction response is wrapped in a markdown code fence.", "Test", 1),
			},
			wantJobs: 1,
			check: func(t *testing.T, jobs []models.JobPost) {
				if jobs[0].WorkMode != models.WorkModeRemote {
					t.Errorf("expected WorkMode 'remote' parsed from the fenced response, got %q", jobs[0].WorkMode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupDB(t)
			svc := newIngestionService(t, db, tt.respond)

			if len(tt.seed) > 0 {
				if failed := svc.ProcessBatch(context.Background(), tt.seed); len(failed) != 0 {
					t.Fatalf("seeding failed: %+v", failed)
				}
			}
			if failed := svc.ProcessBatch(context.Background(), tt.batch); len(failed) != 0 {
				t.Fatalf("ProcessBatch failed: %+v", failed)
			}

			jobs := loadJobs(t, db)
			if len(jobs) != tt.wantJobs {
				t.Fatalf("expected %d job(s) saved, got %d", tt.wantJobs, len(jobs))
			}
			tt.check(t, jobs)
		})
	}
}

// TestProcessBatch_Failures covers batches where some jobs fail: which
// status each failure maps to (4xx = don't retry, 503 = retry), and that
// the jobs that didn't fail were still saved.
func TestProcessBatch_Failures(t *testing.T) {
	tests := []struct {
		name    string
		respond mocks.Responder // nil means extractionResponder(validExtractionJSON)
		seed    []ingestion.RawJobInput
		// setup runs after seeding and before the batch, to break something.
		setup func(t *testing.T, db *gorm.DB)
		batch []ingestion.RawJobInput
		// wantFailures maps JobID to the status code it must fail with. Any
		// job not listed must not fail.
		wantFailures map[string]int
		// wantJobs is the total number of job rows after seed + batch.
		wantJobs int
	}{
		{
			// The critical batch requirement: one job's extraction failure
			// must not abort the rest. job-2's response is garbage; job-1 and
			// job-3 must still be saved.
			name: "extraction failure on one job does not stop the rest of the batch",
			respond: func(prompt string) string {
				if strings.HasPrefix(prompt, "Extract structural data from") {
					if strings.Contains(prompt, "BadCo") {
						return "not valid json"
					}
					return validExtractionJSON
				}
				return `{"id": ` + itoa(mocks.FirstListedID(prompt)) + `}`
			},
			batch: []ingestion.RawJobInput{
				rawJob("job-1", "GoodCo Alpha", "Engineer", "Colombo", "A perfectly normal backend engineering role.", "Test", 1),
				rawJob("job-2", "BadCo", "Analyst", "Kandy", "A role whose extraction response will come back as garbage.", "Test", 1),
				rawJob("job-3", "GoodCo Beta", "Designer", "Galle", "A completely unrelated design role at a different company.", "Test", 1),
			},
			wantFailures: map[string]int{"job-2": http.StatusServiceUnavailable},
			wantJobs:     2,
		},
		{
			// A province missing from geo_data fails identically on every
			// retry, so it must map to 422.
			name: "permanent save failure is 422",
			setup: func(t *testing.T, db *gorm.DB) {
				if err := db.Where("province = ?", "Western").Delete(&models.GeoData{}).Error; err != nil {
					t.Fatalf("removing province: %v", err)
				}
			},
			batch: []ingestion.RawJobInput{
				rawJob("job-perm", "Permanent Fail Co", "Tester", "Colombo", "A job whose save will fail permanently.", "Test", 1),
			},
			wantFailures: map[string]int{"job-perm": http.StatusUnprocessableEntity},
			wantJobs:     0,
		},
		{
			name: "transient save failure is 503",
			setup: func(t *testing.T, db *gorm.DB) {
				failWrites(t, db, "job_post", errors.New("connection reset by peer"))
			},
			batch: []ingestion.RawJobInput{
				rawJob("job-transient", "Transient Fail Co", "Tester", "Colombo", "A job whose save will fail transiently.", "Test", 1),
			},
			wantFailures: map[string]int{"job-transient": http.StatusServiceUnavailable},
			wantJobs:     0,
		},
		{
			// No permanent/transient split here: it's a single-column update
			// on an already-known-valid row, so it is always retryable.
			name: "failing to update a duplicate is 503",
			seed: []ingestion.RawJobInput{ceylonJob},
			setup: func(t *testing.T, db *gorm.DB) {
				failWrites(t, db, "meta_data", errors.New("db timeout"))
			},
			batch:        []ingestion.RawJobInput{ceylonRepost},
			wantFailures: map[string]int{"dup-1": http.StatusServiceUnavailable},
			wantJobs:     1,
		},
		{
			name: "unregistered source is 422",
			batch: []ingestion.RawJobInput{
				rawJob("job-1", "A Co", "Engineer", "Colombo", "Unknown source.", "NoSuchSource", 1),
			},
			wantFailures: map[string]int{"job-1": http.StatusUnprocessableEntity},
			wantJobs:     0,
		},
		{
			name: "disabled source is 422",
			setup: func(t *testing.T, db *gorm.DB) {
				if err := db.Model(&models.Source{}).Where("name = ?", "TopJobs").Update("is_active", false).Error; err != nil {
					t.Fatalf("disabling source: %v", err)
				}
			},
			batch: []ingestion.RawJobInput{
				rawJob("job-1", "B Co", "Designer", "Kandy", "Disabled source.", "TopJobs", 1),
			},
			wantFailures: map[string]int{"job-1": http.StatusUnprocessableEntity},
			wantJobs:     0,
		},
		{
			name: "job with no source is 400",
			batch: []ingestion.RawJobInput{
				rawJob("job-1", "C Co", "Analyst", "Galle", "No source given.", "", 1),
			},
			wantFailures: map[string]int{"job-1": http.StatusBadRequest},
			wantJobs:     0,
		},
		{
			name: "job with no usable text is 400",
			batch: []ingestion.RawJobInput{
				rawJob("job-1", "", "", "", "  ", "Test", 1),
			},
			wantFailures: map[string]int{"job-1": http.StatusBadRequest},
			wantJobs:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respond := tt.respond
			if respond == nil {
				respond = extractionResponder(validExtractionJSON)
			}

			db := setupDB(t)
			svc := newIngestionService(t, db, respond)

			if len(tt.seed) > 0 {
				if failed := svc.ProcessBatch(context.Background(), tt.seed); len(failed) != 0 {
					t.Fatalf("seeding failed: %+v", failed)
				}
			}
			if tt.setup != nil {
				tt.setup(t, db)
			}

			failed := svc.ProcessBatch(context.Background(), tt.batch)

			got := map[string]int{}
			for _, f := range failed {
				got[f.JobID] = f.StatusCode
			}
			if len(got) != len(tt.wantFailures) {
				t.Fatalf("expected failures %v, got %+v", tt.wantFailures, failed)
			}
			for jobID, wantStatus := range tt.wantFailures {
				if got[jobID] != wantStatus {
					t.Errorf("job %q: expected status %d, got %d (%+v)", jobID, wantStatus, got[jobID], failed)
				}
			}

			if n := countRows(t, db, &models.JobPost{}); n != int64(tt.wantJobs) {
				t.Errorf("expected %d job(s) saved, got %d", tt.wantJobs, n)
			}
		})
	}
}
