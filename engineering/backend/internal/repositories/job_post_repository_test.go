package repositories_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/testutil"
)

// ---------- GetJobsByBucketKeys ----------

func TestStore_GetJobsByBucketKeys(t *testing.T) {
	// The queries only read, so every case shares one data set:
	//   match - active job with bucket_a
	//   other - active job with bucket_b
	//   ended - ended job with bucket_a
	//   multi - active job with three buckets
	db := testutil.SetupTestDB(t)
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

// Not part of the table below: it checks every field of one saved job.
func TestStore_SaveOneJob_PersistsJobWithLookupsAndLshRecords(t *testing.T) {
	db := testutil.SetupTestDB(t)
	store := repositories.NewStore(db)
	geoID := seedProvince(t, db, "Western")
	sourceID := seedSource(t, db, "Rooster")
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
	if meta.SourceID == nil || *meta.SourceID != sourceID {
		t.Errorf("expected source %d to be linked, got %v", sourceID, meta.SourceID)
	}
	if meta.AiVersionID == nil {
		t.Errorf("expected ai version to be linked, got %v", meta.AiVersionID)
	}
	if meta.CrawlerRunID == nil || *meta.CrawlerRunID != runID {
		t.Errorf("expected crawler run %d, got %v", runID, meta.CrawlerRunID)
	}

	if n := countRows(t, db, &models.LshIndex{}, "job_post_id = ?", job.ID); n != 2 {
		t.Errorf("expected 2 lsh rows keyed to the new job id, got %d", n)
	}
}

// Not part of the table below: it needs two saves in a row.
func TestStore_SaveOneJob_ReusesExistingLookupRows(t *testing.T) {
	db := testutil.SetupTestDB(t)
	store := repositories.NewStore(db)
	seedProvince(t, db, "Western")
	seedSource(t, db, "Rooster")

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

// TestStore_SaveOneJob covers the scenarios that differ only in the job that is
// saved and the rows that end up in the database. The two tests around it stay
// separate: one checks every field of a single saved job, the other saves two
// jobs in a row.
func TestStore_SaveOneJob(t *testing.T) {
	// rowCounts is the number of rows expected in each table after the save
	type rowCounts struct {
		jobs, employers, jobTypes, skills, sources, aiVersions, lshRows int64
	}

	twoLshRows := []models.LshIndex{
		{BucketKey: "bucket_a", BandNo: 0},
		{BucketKey: "bucket_b", BandNo: 1},
	}

	// Every case seeds the "Rooster" source unless skipSourceSeed is set:
	// sources are seeded, never created by SaveOneJob, so the source row count
	// is 1 whether or not the job ends up linked to it.
	tests := []struct {
		name           string
		seedProvince   bool // whether the "Western" province exists in geo_data
		skipSourceSeed bool // whether to leave the "Rooster" source out
		sourceDisabled bool // whether the seeded "Rooster" source is disabled
		modify         func(job *models.JobPost)
		lshRecords     []models.LshIndex
		wantErr        error  // a sentinel the error must match
		wantErrText    string // text the error must contain, when there is no sentinel
		want           rowCounts
		sourceLinked   bool // whether the saved job should point at the source
	}{
		{
			name:         "saves a job with all its lookups and lsh rows",
			seedProvince: true,
			lshRecords:   twoLshRows,
			want:         rowCounts{jobs: 1, employers: 1, jobTypes: 1, skills: 2, sources: 1, aiVersions: 1, lshRows: 2},
			sourceLinked: true,
		},
		{
			name:         "saves a job without lsh rows",
			seedProvince: true,
			want:         rowCounts{jobs: 1, employers: 1, jobTypes: 1, skills: 2, sources: 1, aiVersions: 1},
			sourceLinked: true,
		},
		{
			name:         "skips empty skill names",
			seedProvince: true,
			modify: func(job *models.JobPost) {
				// The empty name comes first on purpose: if it were not skipped it
				// would be created as a blank skill. After "Go" exists, GORM would
				// just look up that row, and the bug would go unnoticed.
				job.Skills = []models.Skill{{Skill: ""}, {Skill: "Go"}}
			},
			want:         rowCounts{jobs: 1, employers: 1, jobTypes: 1, skills: 1, sources: 1, aiVersions: 1},
			sourceLinked: true,
		},
		{
			name:         "treats empty lookup names as absent",
			seedProvince: true,
			modify: func(job *models.JobPost) {
				job.Employer.Name = ""
				job.JobType.Type = ""
				job.MetaData.Source.Name = ""
				job.MetaData.AiVersion.Version = ""
			},
			want: rowCounts{jobs: 1, skills: 2, sources: 1},
		},
		{
			name: "saves a job without any optional relation",
			modify: func(job *models.JobPost) {
				job.Employer = nil
				job.JobType = nil
				job.Skills = nil
				job.MetaData.GeoData = nil
				job.MetaData.Source = nil
				job.MetaData.AiVersion = nil
			},
			want: rowCounts{jobs: 1, sources: 1},
		},
		{
			// The employer, job type and skills are created before the geo
			// lookup fails, so every count being 0 shows the rollback worked.
			name:       "unknown province is a permanent failure and rolls everything back",
			lshRecords: twoLshRows,
			wantErr:    repositories.ErrPermanentSaveFailure,
			want:       rowCounts{sources: 1},
		},
		{
			// Sources are seeded, never created on save. sources being 0 shows
			// no row was created; the rest being 0 shows the rollback worked.
			name:           "unknown source is a permanent failure and creates no source",
			seedProvince:   true,
			skipSourceSeed: true,
			lshRecords:     twoLshRows,
			wantErr:        repositories.ErrPermanentSaveFailure,
			want:           rowCounts{},
		},
		{
			// The source exists but is disabled. The text check tells this apart
			// from "not registered"; the counts show the rollback worked and the
			// source row was left alone.
			name:           "disabled source is a permanent failure and rolls everything back",
			seedProvince:   true,
			sourceDisabled: true,
			lshRecords:     twoLshRows,
			wantErr:        repositories.ErrPermanentSaveFailure,
			wantErrText:    "source 'Rooster' is disabled",
			want:           rowCounts{sources: 1},
		},
		{
			// Two records with the same bucket key break the (bucket_key,
			// job_post_id) primary key. The job row and the lookups are already
			// written by then, so every count being 0 shows the rollback worked.
			name:         "failing lsh insert rolls the job and its lookups back",
			seedProvince: true,
			lshRecords: []models.LshIndex{
				{BucketKey: "bucket_a", BandNo: 0},
				{BucketKey: "bucket_a", BandNo: 1},
			},
			wantErrText: "lsh_index insert failed",
			want:        rowCounts{sources: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := testutil.SetupTestDB(t)
			store := repositories.NewStore(db)
			if tt.seedProvince {
				seedProvince(t, db, "Western")
			}
			if !tt.skipSourceSeed {
				seedSourceOf(t, db, "Rooster", models.SourceTypeAuto, !tt.sourceDisabled)
			}

			job := newJobToSave("Engineer")
			if tt.modify != nil {
				tt.modify(job)
			}

			err := store.SaveOneJob(context.Background(), job, tt.lshRecords)
			wantFailure := tt.wantErr != nil || tt.wantErrText != ""
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErrText != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrText)) {
				t.Fatalf("expected an error containing %q, got %v", tt.wantErrText, err)
			}
			if !wantFailure && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := rowCounts{
				jobs:       countRows(t, db, &models.JobPost{}, ""),
				employers:  countRows(t, db, &models.Employer{}, ""),
				jobTypes:   countRows(t, db, &models.JobType{}, ""),
				skills:     countRows(t, db, &models.Skill{}, ""),
				sources:    countRows(t, db, &models.Source{}, ""),
				aiVersions: countRows(t, db, &models.AiVersion{}, ""),
				lshRows:    countRows(t, db, &models.LshIndex{}, ""),
			}
			if got != tt.want {
				t.Errorf("row counts: want %+v, got %+v", tt.want, got)
			}

			if wantFailure {
				return
			}
			// A lookup is only linked to the job if the job named it
			meta := loadMeta(t, db, job.ID)
			if (meta.SourceID != nil) != tt.sourceLinked {
				t.Errorf("source link: want linked=%t, got %v", tt.sourceLinked, meta.SourceID)
			}
			if (meta.AiVersionID != nil) != (tt.want.aiVersions > 0) {
				t.Errorf("ai version link: want linked=%t, got %v", tt.want.aiVersions > 0, meta.AiVersionID)
			}
		})
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
			db := testutil.SetupTestDB(t)
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
	//   staleDisabled  - job from the old run of a disabled auto source
	//   staleManual    - job from the old run of a manual (newspaper) source
	//   staleDeleted   - job from the old run of a soft-deleted source
	//
	// A NULL crawler run counts as stale, but only for a listed source. A job
	// without a source is never reconciled, because it can't be attributed to one.
	// A disabled or deleted source is never reconciled, even when listed. A listed
	// manual source is reconciled like any other.
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
			name:    "disabled source keeps its open jobs",
			sources: []string{"Disabled"},
		},
		{
			name:    "deleted source keeps its open jobs",
			sources: []string{"Deleted"},
		},
		{
			name:         "manual source is reconciled like any other",
			sources:      []string{"Daily News"},
			wantAffected: 1,
			wantExpired:  []string{"staleManual"},
		},
		{
			name:         "disabled and deleted sources are skipped alongside active ones",
			sources:      []string{"Rooster", "Disabled", "Deleted", "Daily News"},
			wantAffected: 3,
			wantExpired:  []string{"staleRooster", "nullRunRooster", "staleManual"},
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
			db := testutil.SetupTestDB(t)
			store := repositories.NewStore(db)

			roosterID := seedSource(t, db, "Rooster")
			ikmanID := seedSource(t, db, "Ikman")
			disabledID := seedSourceOf(t, db, "Disabled", models.SourceTypeAuto, false)
			manualID := seedSourceOf(t, db, "Daily News", models.SourceTypeManual, true)
			deletedID := seedSource(t, db, "Deleted")
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
				"staleDisabled":  seedActiveJob(t, db, "stale disabled", 1, &disabledID, &oldRunID, "bucket_stale_disabled"),
				"staleManual":    seedActiveJob(t, db, "stale manual", 1, &manualID, &oldRunID, "bucket_stale_manual"),
				"staleDeleted":   seedActiveJob(t, db, "stale deleted", 1, &deletedID, &oldRunID, "bucket_stale_deleted"),
			}
			endJob(t, db, jobs["endedRooster"])
			// Soft delete after its job exists, as in production.
			if err := db.Delete(&models.Source{}, deletedID).Error; err != nil {
				t.Fatalf("failed to soft-delete source: %v", err)
			}

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

			// Every job closed by one reconcile call must share a single end date.
			var sharedEndDate *time.Time
			var sharedLabel string

			for label, id := range jobs {
				endDate := loadMeta(t, db, id).EndDate
				lshRows := countRows(t, db, &models.LshIndex{}, "job_post_id = ?", id)

				if expired[label] {
					if endDate == nil {
						t.Errorf("%s: expected the job to be expired", label)
					} else if sharedEndDate == nil {
						sharedEndDate, sharedLabel = endDate, label
					} else if !endDate.Equal(*sharedEndDate) {
						t.Errorf("%s: expected the same end date as %s (%v), got %v", label, sharedLabel, *sharedEndDate, *endDate)
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
