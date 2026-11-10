package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

// ---------- GetJobsByBucketKeys ----------

func TestStore_GetJobsByBucketKeys(t *testing.T) {
	// The queries only read, so every case shares one data set:
	//   match - active job with bucket_a
	//   other - active job with bucket_b
	//   ended - ended job with bucket_a
	//   multi - active job with three buckets
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	sourceID := seedSource(t, db, "Rooster")
	runID := seedCrawlerRun(t, db)

	jobs := map[string]uint{
		"match": seedActiveJob(t, db, "match", 1, &sourceID, &runID, "bucket_a"),
		"other": seedActiveJob(t, db, "other", 1, &sourceID, &runID, "bucket_b"),
		"ended": seedActiveJob(t, db, "ended", 1, &sourceID, &runID, "bucket_a"),
		"multi": seedActiveJob(t, db, "multi", 1, &sourceID, &runID, "bucket_c", "bucket_d", "bucket_e"),
	}
	endJob(t, db, jobs["ended"])

	tests := []struct {
		name string
		keys []string
		want []string
	}{
		{"returns jobs with a matching key", []string{"bucket_a"}, []string{"match"}},
		{"excludes ended jobs", []string{"bucket_a"}, []string{"match"}},
		{"returns a job once when several keys match", []string{"bucket_c", "bucket_d", "bucket_e"}, []string{"multi"}},
		{"returns jobs from different keys", []string{"bucket_a", "bucket_b"}, []string{"match", "other"}},
		{"unknown key returns nothing", []string{"does_not_exist"}, nil},
		{"empty key list returns nothing", []string{}, nil},
		{"nil key list returns nothing", nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetJobsByBucketKeys(context.Background(), tt.keys)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("expected %d job(s), got %d: %v", len(tt.want), len(got), jobIDs(got))
			}
			ids := jobIDs(got)
			for _, label := range tt.want {
				if !ids[jobs[label]] {
					t.Errorf("expected %s (job %d) in the result, got %v", label, jobs[label], ids)
				}
			}
			for _, job := range got {
				if job.MetaData.JobPostID != job.ID {
					t.Errorf("expected MetaData to be preloaded for job %d, got job_post_id %d", job.ID, job.MetaData.JobPostID)
				}
			}
		})
	}
}

// ---------- SaveOneJob ----------

func newJobToSave(role string) *models.JobPost {
	return &models.JobPost{
		JobRole:       role,
		WorkMode:      models.WorkModeOnsite,
		NoOfVacancies: 1,
		Employer:      &models.Employer{Name: "Acme"},
		JobType:       &models.JobType{Type: "Full time"},
		Skills:        []models.Skill{{Skill: "Go"}, {Skill: "SQL"}},
		MetaData: models.JobMetaData{
			GeoData:   &models.GeoData{Province: "Western"},
			Source:    &models.Source{Name: "Rooster"},
			AiVersion: &models.AiVersion{Version: "v1"},
			PostedAt:  time.Now(),
		},
	}
}

func TestStore_SaveOneJob_PersistsJobWithLookupsAndLshRecords(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	geoID := seedProvince(t, db, "Western")
	runID := seedCrawlerRun(t, db)

	job := newJobToSave("Engineer")
	job.MetaData.CrawlerRunID = &runID
	lsh := []models.LshIndex{
		{BucketKey: "bucket_a", BandNo: 0},
		{BucketKey: "bucket_b", BandNo: 1},
	}

	if err := store.SaveOneJob(context.Background(), job, lsh); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID == 0 {
		t.Fatal("expected the job to get a database id")
	}

	var saved models.JobPost
	if err := db.Preload("Employer").Preload("JobType").Preload("Skills").
		Preload("MetaData").First(&saved, job.ID).Error; err != nil {
		t.Fatalf("failed to reload job: %v", err)
	}

	if saved.JobRole != "Engineer" {
		t.Errorf("expected job role %q, got %q", "Engineer", saved.JobRole)
	}
	if saved.Employer == nil || saved.Employer.Name != "Acme" {
		t.Errorf("expected employer Acme to be linked, got %+v", saved.Employer)
	}
	if saved.JobType == nil || saved.JobType.Type != "Full time" {
		t.Errorf("expected job type Full time to be linked, got %+v", saved.JobType)
	}
	if len(saved.Skills) != 2 {
		t.Errorf("expected 2 linked skills, got %d", len(saved.Skills))
	}

	meta := saved.MetaData
	if meta.GeoDataID == nil || *meta.GeoDataID != geoID {
		t.Errorf("expected geo data %d to be linked, got %v", geoID, meta.GeoDataID)
	}
	if meta.SourceID == nil || meta.AiVersionID == nil {
		t.Errorf("expected source and ai version to be linked, got source=%v ai_version=%v", meta.SourceID, meta.AiVersionID)
	}
	if meta.CrawlerRunID == nil || *meta.CrawlerRunID != runID {
		t.Errorf("expected crawler run %d, got %v", runID, meta.CrawlerRunID)
	}

	if n := countRows(t, db, &models.LshIndex{}, "job_post_id = ?", job.ID); n != 2 {
		t.Errorf("expected 2 lsh rows keyed to the new job id, got %d", n)
	}
}

func TestStore_SaveOneJob_ReusesExistingLookupRows(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	seedProvince(t, db, "Western")

	if err := store.SaveOneJob(context.Background(), newJobToSave("First"), nil); err != nil {
		t.Fatalf("unexpected error saving first job: %v", err)
	}
	if err := store.SaveOneJob(context.Background(), newJobToSave("Second"), nil); err != nil {
		t.Fatalf("unexpected error saving second job: %v", err)
	}

	for name, model := range map[string]interface{}{
		"employer":   &models.Employer{},
		"job_type":   &models.JobType{},
		"skills":     &models.Skill{},
		"source":     &models.Source{},
		"ai_version": &models.AiVersion{},
	} {
		want := int64(1)
		if name == "skills" {
			want = 2
		}
		if n := countRows(t, db, model, ""); n != want {
			t.Errorf("%s: expected %d row(s) after saving two jobs, got %d", name, want, n)
		}
	}
	if n := countRows(t, db, &models.JobPost{}, ""); n != 2 {
		t.Errorf("expected 2 jobs, got %d", n)
	}
}

func TestStore_SaveOneJob_SkipsEmptySkillNames(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	seedProvince(t, db, "Western")

	job := newJobToSave("Engineer")
	job.Skills = []models.Skill{{Skill: "Go"}, {Skill: ""}}

	if err := store.SaveOneJob(context.Background(), job, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if n := countRows(t, db, &models.Skill{}, ""); n != 1 {
		t.Errorf("expected only the non-empty skill to be created, got %d", n)
	}
}

func TestStore_SaveOneJob_UnknownProvinceIsPermanentFailureAndRollsBack(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	// "Western" is deliberately not seeded into geo_data.

	err := store.SaveOneJob(context.Background(), newJobToSave("Engineer"), []models.LshIndex{{BucketKey: "bucket_a"}})
	if !errors.Is(err, repositories.ErrPermanentSaveFailure) {
		t.Fatalf("expected ErrPermanentSaveFailure, got %v", err)
	}

	// The employer/job type/skills were created before the geo lookup failed;
	// the transaction must have rolled all of it back.
	for name, model := range map[string]interface{}{
		"job_post":  &models.JobPost{},
		"employer":  &models.Employer{},
		"job_type":  &models.JobType{},
		"skills":    &models.Skill{},
		"lsh_index": &models.LshIndex{},
	} {
		if n := countRows(t, db, model, ""); n != 0 {
			t.Errorf("%s: expected nothing to be saved after a failed save, got %d row(s)", name, n)
		}
	}
}

func TestStore_SaveOneJob_WithoutOptionalRelationsStillSaves(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)

	job := &models.JobPost{
		JobRole:       "Bare job",
		WorkMode:      models.WorkModeOnsite,
		NoOfVacancies: 1,
		MetaData:      models.JobMetaData{PostedAt: time.Now()},
	}

	if err := store.SaveOneJob(context.Background(), job, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta := loadMeta(t, db, job.ID)
	if meta.SourceID != nil || meta.GeoDataID != nil || meta.AiVersionID != nil {
		t.Errorf("expected no linked lookups, got source=%v geo=%v ai_version=%v", meta.SourceID, meta.GeoDataID, meta.AiVersionID)
	}
}

// ---------- UpdateDuplicateJob ----------

// missingJobID is an id no seeded job has.
const missingJobID = 9999

func TestStore_UpdateDuplicateJob(t *testing.T) {
	tests := []struct {
		name          string
		target        string // label of the job to re-stamp; "" means an unknown id
		wantRestamped []string
	}{
		{"re-stamps only the given job", "seen", []string{"seen"}},
		{"unknown job is a no-op", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := SetupTestDB(t)
			store := repositories.NewStore(db)
			sourceID := seedSource(t, db, "Rooster")
			oldRunID := seedCrawlerRun(t, db)
			newRunID := seedCrawlerRun(t, db)

			jobs := map[string]uint{
				"seen":      seedActiveJob(t, db, "seen", 1, &sourceID, &oldRunID),
				"untouched": seedActiveJob(t, db, "untouched", 1, &sourceID, &oldRunID),
			}

			targetID := uint(missingJobID)
			if tt.target != "" {
				targetID = jobs[tt.target]
			}
			if err := store.UpdateDuplicateJob(context.Background(), targetID, newRunID); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			restamped := map[string]bool{}
			for _, label := range tt.wantRestamped {
				restamped[label] = true
			}
			for label, id := range jobs {
				want := oldRunID
				if restamped[label] {
					want = newRunID
				}
				meta := loadMeta(t, db, id)
				if meta.CrawlerRunID == nil || *meta.CrawlerRunID != want {
					t.Errorf("%s: want run %d, got %v", label, want, meta.CrawlerRunID)
				}
				if meta.EndDate != nil {
					t.Errorf("%s: end_date must stay NULL, got %v", label, meta.EndDate)
				}
			}
		})
	}
}

// ---------- ReconcileStaleVacancies ----------

func TestStore_ReconcileStaleVacancies(t *testing.T) {
	// Every case runs against the same data set:
	//   staleRooster   - Rooster job last seen in the old run
	//   staleIkman     - Ikman job last seen in the old run
	//   currentRooster - Rooster job seen in the current run
	//   endedRooster   - Rooster job from the old run that is already ended
	//   nullRunRooster - Rooster job with no crawler run (NULL)
	//   nullRunIkman   - Ikman job with no crawler run (NULL)
	//   nullSource     - job from the old run with no source (NULL)
	//   nullSourceRun  - job with neither a source nor a crawler run (NULL)
	//
	// A NULL crawler run counts as stale, but only for a listed source. A job
	// without a source is never reconciled, because it can't be attributed to one.
	tests := []struct {
		name         string
		sources      []string
		unknownRun   bool
		wantAffected int64
		wantExpired  []string
		wantErr      error
	}{
		{
			name:         "only the listed source is reconciled",
			sources:      []string{"Rooster"},
			wantAffected: 2,
			wantExpired:  []string{"staleRooster", "nullRunRooster"},
		},
		{
			name:         "other sources stay untouched",
			sources:      []string{"Ikman"},
			wantAffected: 2,
			wantExpired:  []string{"staleIkman", "nullRunIkman"},
		},
		{
			name:         "several sources are reconciled together",
			sources:      []string{"Rooster", "Ikman"},
			wantAffected: 4,
			wantExpired:  []string{"staleRooster", "staleIkman", "nullRunRooster", "nullRunIkman"},
		},
		{
			name:    "unknown source is a no-op",
			sources: []string{"DoesNotExist"},
		},
		{
			name:    "nil sources is a no-op",
			sources: nil,
		},
		{
			name:    "empty sources is a no-op",
			sources: []string{},
		},
		{
			name:       "unknown crawler run returns an error",
			sources:    []string{"Rooster"},
			unknownRun: true,
			wantErr:    gorm.ErrRecordNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := SetupTestDB(t)
			store := repositories.NewStore(db)

			roosterID := seedSource(t, db, "Rooster")
			ikmanID := seedSource(t, db, "Ikman")
			oldRunID := seedCrawlerRun(t, db)
			currentRunID := seedCrawlerRun(t, db)

			jobs := map[string]uint{
				"staleRooster":   seedActiveJob(t, db, "stale rooster", 1, &roosterID, &oldRunID, "bucket_stale_rooster"),
				"staleIkman":     seedActiveJob(t, db, "stale ikman", 1, &ikmanID, &oldRunID, "bucket_stale_ikman"),
				"currentRooster": seedActiveJob(t, db, "current rooster", 1, &roosterID, &currentRunID, "bucket_current_rooster"),
				"endedRooster":   seedActiveJob(t, db, "ended rooster", 1, &roosterID, &oldRunID, "bucket_ended_rooster"),
				"nullRunRooster": seedActiveJob(t, db, "null run rooster", 1, &roosterID, nil, "bucket_null_run_rooster"),
				"nullRunIkman":   seedActiveJob(t, db, "null run ikman", 1, &ikmanID, nil, "bucket_null_run_ikman"),
				"nullSource":     seedActiveJob(t, db, "null source", 1, nil, &oldRunID, "bucket_null_source"),
				"nullSourceRun":  seedActiveJob(t, db, "null source and run", 1, nil, nil, "bucket_null_source_run"),
			}
			endJob(t, db, jobs["endedRooster"])

			endDateBefore := make(map[string]*time.Time, len(jobs))
			for label, id := range jobs {
				endDateBefore[label] = loadMeta(t, db, id).EndDate
			}

			runID := currentRunID
			if tt.unknownRun {
				runID = 9999
			}

			affected, err := store.ReconcileStaleVacancies(context.Background(), runID, tt.sources)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if affected != tt.wantAffected {
					t.Errorf("expected %d reconciled job(s), got %d", tt.wantAffected, affected)
				}
			}

			expired := make(map[string]bool, len(tt.wantExpired))
			for _, label := range tt.wantExpired {
				expired[label] = true
			}

			for label, id := range jobs {
				endDate := loadMeta(t, db, id).EndDate
				lshRows := countRows(t, db, &models.LshIndex{}, "job_post_id = ?", id)

				if expired[label] {
					if endDate == nil {
						t.Errorf("%s: expected the job to be expired", label)
					}
					if lshRows != 0 {
						t.Errorf("%s: expected its LSH rows to be removed, got %d", label, lshRows)
					}
					continue
				}

				before := endDateBefore[label]
				if (before == nil) != (endDate == nil) || (before != nil && !before.Equal(*endDate)) {
					t.Errorf("%s: end date must not change, was %v now %v", label, before, endDate)
				}
				if lshRows != 1 {
					t.Errorf("%s: expected its LSH rows to be kept, got %d", label, lshRows)
				}
			}
		})
	}
}
