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

func TestStore_GetJobsByBucketKeys_ReturnsMatchingActiveJobsWithMetaData(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	sourceID := seedSource(t, db, "Rooster")
	runID := seedCrawlerRun(t, db)

	matchID := seedActiveJob(t, db, "match", sourceID, runID, "bucket_a")
	otherID := seedActiveJob(t, db, "other", sourceID, runID, "bucket_b")

	jobs, err := store.GetJobsByBucketKeys(context.Background(), []string{"bucket_a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := jobIDs(jobs)
	if len(jobs) != 1 || !ids[matchID] {
		t.Fatalf("expected only job %d, got %v", matchID, ids)
	}
	if ids[otherID] {
		t.Errorf("job %d has a different bucket key and must not be returned", otherID)
	}
	if jobs[0].MetaData.JobPostID != matchID {
		t.Errorf("expected MetaData to be preloaded for job %d, got job_post_id %d", matchID, jobs[0].MetaData.JobPostID)
	}
}

func TestStore_GetJobsByBucketKeys_ExcludesEndedJobs(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	sourceID := seedSource(t, db, "Rooster")
	runID := seedCrawlerRun(t, db)

	activeID := seedActiveJob(t, db, "active", sourceID, runID, "bucket_a")
	endedID := seedActiveJob(t, db, "ended", sourceID, runID, "bucket_a")
	endJob(t, db, endedID)

	jobs, err := store.GetJobsByBucketKeys(context.Background(), []string{"bucket_a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := jobIDs(jobs)
	if len(jobs) != 1 || !ids[activeID] {
		t.Errorf("expected only the active job %d, got %v", activeID, ids)
	}
}

func TestStore_GetJobsByBucketKeys_ReturnsAJobOnceWhenSeveralKeysMatch(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	sourceID := seedSource(t, db, "Rooster")
	runID := seedCrawlerRun(t, db)

	jobID := seedActiveJob(t, db, "multi band", sourceID, runID, "bucket_a", "bucket_b", "bucket_c")

	jobs, err := store.GetJobsByBucketKeys(context.Background(), []string{"bucket_a", "bucket_b", "bucket_c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 1 || jobs[0].ID != jobID {
		t.Errorf("expected job %d exactly once, got %d result(s)", jobID, len(jobs))
	}
}

func TestStore_GetJobsByBucketKeys_NoMatchReturnsEmpty(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	sourceID := seedSource(t, db, "Rooster")
	runID := seedCrawlerRun(t, db)
	seedActiveJob(t, db, "job", sourceID, runID, "bucket_a")

	for name, keys := range map[string][]string{
		"unknown key": {"does_not_exist"},
		"empty list":  {},
	} {
		t.Run(name, func(t *testing.T) {
			jobs, err := store.GetJobsByBucketKeys(context.Background(), keys)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(jobs) != 0 {
				t.Errorf("expected no jobs, got %d", len(jobs))
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

func TestStore_UpdateDuplicateJob_RestampsOnlyTheGivenJob(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	sourceID := seedSource(t, db, "Rooster")
	oldRunID := seedCrawlerRun(t, db)
	newRunID := seedCrawlerRun(t, db)

	seenID := seedActiveJob(t, db, "seen again", sourceID, oldRunID)
	untouchedID := seedActiveJob(t, db, "not seen", sourceID, oldRunID)

	if err := store.UpdateDuplicateJob(context.Background(), seenID, newRunID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := loadMeta(t, db, seenID).CrawlerRunID; got == nil || *got != newRunID {
		t.Errorf("expected job %d to move to run %d, got %v", seenID, newRunID, got)
	}
	if got := loadMeta(t, db, untouchedID).CrawlerRunID; got == nil || *got != oldRunID {
		t.Errorf("job %d must keep run %d, got %v", untouchedID, oldRunID, got)
	}
}

func TestStore_UpdateDuplicateJob_UnknownJobIsNoOp(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	sourceID := seedSource(t, db, "Rooster")
	oldRunID := seedCrawlerRun(t, db)
	newRunID := seedCrawlerRun(t, db)
	existingID := seedActiveJob(t, db, "existing", sourceID, oldRunID)

	if err := store.UpdateDuplicateJob(context.Background(), 9999, newRunID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := loadMeta(t, db, existingID).CrawlerRunID; got == nil || *got != oldRunID {
		t.Errorf("job %d must keep run %d, got %v", existingID, oldRunID, got)
	}
}

// ---------- ReconcileStaleVacancies ----------

func TestStore_ReconcileStaleVacancies(t *testing.T) {
	// Every case runs against the same data set:
	//   staleRooster   - Rooster job last seen in the old run
	//   staleIkman     - Ikman job last seen in the old run
	//   currentRooster - Rooster job seen in the current run
	//   endedRooster   - Rooster job from the old run that is already ended
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
			wantAffected: 1,
			wantExpired:  []string{"staleRooster"},
		},
		{
			name:         "other sources stay untouched",
			sources:      []string{"Ikman"},
			wantAffected: 1,
			wantExpired:  []string{"staleIkman"},
		},
		{
			name:         "several sources are reconciled together",
			sources:      []string{"Rooster", "Ikman"},
			wantAffected: 2,
			wantExpired:  []string{"staleRooster", "staleIkman"},
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
				"staleRooster":   seedActiveJob(t, db, "stale rooster", roosterID, oldRunID, "bucket_stale_rooster"),
				"staleIkman":     seedActiveJob(t, db, "stale ikman", ikmanID, oldRunID, "bucket_stale_ikman"),
				"currentRooster": seedActiveJob(t, db, "current rooster", roosterID, currentRunID, "bucket_current_rooster"),
				"endedRooster":   seedActiveJob(t, db, "ended rooster", roosterID, oldRunID, "bucket_ended_rooster"),
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
