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

// ---------- Helpers ----------

func seedSource(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()
	src := models.Source{Name: name}
	if err := db.Create(&src).Error; err != nil {
		t.Fatalf("failed to create source %q: %v", name, err)
	}
	return src.ID
}

func seedCrawlerRun(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	run := models.CrawlerRun{Status: "RUNNING"}
	if err := db.Create(&run).Error; err != nil {
		t.Fatalf("failed to create crawler run: %v", err)
	}
	return run.ID
}

// seedActiveJob creates an active (end_date IS NULL) job for the given source
// and crawler run, with one LSH row per bucket key.
func seedActiveJob(t *testing.T, db *gorm.DB, role string, sourceID, runID uint, bucketKeys ...string) uint {
	t.Helper()
	job := models.JobPost{
		JobRole:       role,
		WorkMode:      models.WorkModeOnsite,
		NoOfVacancies: 1,
		MetaData: models.JobMetaData{
			SourceID:     &sourceID,
			CrawlerRunID: &runID,
			PostedAt:     time.Now(),
		},
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("failed to create job %q: %v", role, err)
	}
	for i, key := range bucketKeys {
		lsh := models.LshIndex{BucketKey: key, BandNo: i, JobPostID: job.ID}
		if err := db.Create(&lsh).Error; err != nil {
			t.Fatalf("failed to create lsh row %q for job %q: %v", key, role, err)
		}
	}
	return job.ID
}

func endJob(t *testing.T, db *gorm.DB, jobID uint) {
	t.Helper()
	now := time.Now()
	if err := db.Model(&models.JobMetaData{}).
		Where("job_post_id = ?", jobID).
		Update("end_date", &now).Error; err != nil {
		t.Fatalf("failed to end job %d: %v", jobID, err)
	}
}

func loadMeta(t *testing.T, db *gorm.DB, jobID uint) models.JobMetaData {
	t.Helper()
	var meta models.JobMetaData
	if err := db.Where("job_post_id = ?", jobID).First(&meta).Error; err != nil {
		t.Fatalf("failed to load meta data for job %d: %v", jobID, err)
	}
	return meta
}

func countRows(t *testing.T, db *gorm.DB, model interface{}, query string, args ...interface{}) int64 {
	t.Helper()
	var n int64
	tx := db.Model(model)
	if query != "" {
		tx = tx.Where(query, args...)
	}
	if err := tx.Count(&n).Error; err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	return n
}

func jobIDs(jobs []models.JobPost) map[uint]bool {
	ids := make(map[uint]bool, len(jobs))
	for _, j := range jobs {
		ids[j.ID] = true
	}
	return ids
}

// ---------- GetJobsByBucketKeys ----------

func TestStore_GetJobsByBucketKeys_ReturnsMatchingActiveJobsWithMetaData(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	source := seedSource(t, db, "Rooster")
	run := seedCrawlerRun(t, db)

	match := seedActiveJob(t, db, "match", source, run, "bucket_a")
	other := seedActiveJob(t, db, "other", source, run, "bucket_b")

	jobs, err := store.GetJobsByBucketKeys(context.Background(), []string{"bucket_a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := jobIDs(jobs)
	if len(jobs) != 1 || !ids[match] {
		t.Fatalf("expected only job %d, got %v", match, ids)
	}
	if ids[other] {
		t.Errorf("job %d has a different bucket key and must not be returned", other)
	}
	if jobs[0].MetaData.JobPostID != match {
		t.Errorf("expected MetaData to be preloaded for job %d, got job_post_id %d", match, jobs[0].MetaData.JobPostID)
	}
}

func TestStore_GetJobsByBucketKeys_ExcludesEndedJobs(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	source := seedSource(t, db, "Rooster")
	run := seedCrawlerRun(t, db)

	active := seedActiveJob(t, db, "active", source, run, "bucket_a")
	ended := seedActiveJob(t, db, "ended", source, run, "bucket_a")
	endJob(t, db, ended)

	jobs, err := store.GetJobsByBucketKeys(context.Background(), []string{"bucket_a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ids := jobIDs(jobs)
	if len(jobs) != 1 || !ids[active] {
		t.Errorf("expected only the active job %d, got %v", active, ids)
	}
}

func TestStore_GetJobsByBucketKeys_ReturnsAJobOnceWhenSeveralKeysMatch(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	source := seedSource(t, db, "Rooster")
	run := seedCrawlerRun(t, db)

	job := seedActiveJob(t, db, "multi band", source, run, "bucket_a", "bucket_b", "bucket_c")

	jobs, err := store.GetJobsByBucketKeys(context.Background(), []string{"bucket_a", "bucket_b", "bucket_c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 1 || jobs[0].ID != job {
		t.Errorf("expected job %d exactly once, got %d result(s)", job, len(jobs))
	}
}

func TestStore_GetJobsByBucketKeys_NoMatchReturnsEmpty(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	source := seedSource(t, db, "Rooster")
	run := seedCrawlerRun(t, db)
	seedActiveJob(t, db, "job", source, run, "bucket_a")

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

func seedProvince(t *testing.T, db *gorm.DB, province string) uint {
	t.Helper()
	geo := models.GeoData{Province: province}
	if err := db.Create(&geo).Error; err != nil {
		t.Fatalf("failed to create geo data %q: %v", province, err)
	}
	return geo.ID
}

func TestStore_SaveOneJob_PersistsJobWithLookupsAndLshRecords(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	geoID := seedProvince(t, db, "Western")
	run := seedCrawlerRun(t, db)

	job := newJobToSave("Engineer")
	job.MetaData.CrawlerRunID = &run
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
	if meta.CrawlerRunID == nil || *meta.CrawlerRunID != run {
		t.Errorf("expected crawler run %d, got %v", run, meta.CrawlerRunID)
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
	source := seedSource(t, db, "Rooster")
	oldRun := seedCrawlerRun(t, db)
	newRun := seedCrawlerRun(t, db)

	seen := seedActiveJob(t, db, "seen again", source, oldRun)
	untouched := seedActiveJob(t, db, "not seen", source, oldRun)

	if err := store.UpdateDuplicateJob(context.Background(), seen, newRun); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := loadMeta(t, db, seen).CrawlerRunID; got == nil || *got != newRun {
		t.Errorf("expected job %d to move to run %d, got %v", seen, newRun, got)
	}
	if got := loadMeta(t, db, untouched).CrawlerRunID; got == nil || *got != oldRun {
		t.Errorf("job %d must keep run %d, got %v", untouched, oldRun, got)
	}
}

func TestStore_UpdateDuplicateJob_UnknownJobIsNoOp(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)
	source := seedSource(t, db, "Rooster")
	oldRun := seedCrawlerRun(t, db)
	newRun := seedCrawlerRun(t, db)
	existing := seedActiveJob(t, db, "existing", source, oldRun)

	if err := store.UpdateDuplicateJob(context.Background(), 9999, newRun); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := loadMeta(t, db, existing).CrawlerRunID; got == nil || *got != oldRun {
		t.Errorf("job %d must keep run %d, got %v", existing, oldRun, got)
	}
}

// ---------- ReconcileStaleVacancies ----------

func TestStore_ReconcileStaleVacancies_OnlyListedSources(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)

	rooster := seedSource(t, db, "Rooster")
	ikman := seedSource(t, db, "Ikman")
	oldRun := seedCrawlerRun(t, db)
	currentRun := seedCrawlerRun(t, db)

	staleListed := seedActiveJob(t, db, "stale rooster", rooster, oldRun, "bucket_stale_listed")
	staleUnlisted := seedActiveJob(t, db, "stale ikman", ikman, oldRun, "bucket_stale_unlisted")
	currentListed := seedActiveJob(t, db, "current rooster", rooster, currentRun, "bucket_current")

	affected, err := store.ReconcileStaleVacancies(context.Background(), currentRun, []string{"Rooster"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 reconciled job, got %d", affected)
	}

	lshCount := func(jobID uint) int64 {
		return countRows(t, db, &models.LshIndex{}, "job_post_id = ?", jobID)
	}

	if loadMeta(t, db, staleListed).EndDate == nil {
		t.Error("stale job of a listed source should have been expired")
	}
	if n := lshCount(staleListed); n != 0 {
		t.Errorf("stale job of a listed source should lose its LSH rows, got %d", n)
	}

	if loadMeta(t, db, staleUnlisted).EndDate != nil {
		t.Error("stale job of an unlisted source must stay active")
	}
	if n := lshCount(staleUnlisted); n != 1 {
		t.Errorf("stale job of an unlisted source must keep its LSH rows, got %d", n)
	}

	if loadMeta(t, db, currentListed).EndDate != nil {
		t.Error("job seen in the current run must stay active")
	}
	if n := lshCount(currentListed); n != 1 {
		t.Errorf("job seen in the current run must keep its LSH rows, got %d", n)
	}
}

func TestStore_ReconcileStaleVacancies_UnknownSourceIsNoOp(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)

	rooster := seedSource(t, db, "Rooster")
	oldRun := seedCrawlerRun(t, db)
	currentRun := seedCrawlerRun(t, db)
	stale := seedActiveJob(t, db, "stale rooster", rooster, oldRun)

	affected, err := store.ReconcileStaleVacancies(context.Background(), currentRun, []string{"DoesNotExist"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 0 {
		t.Errorf("expected 0 reconciled jobs, got %d", affected)
	}
	if loadMeta(t, db, stale).EndDate != nil {
		t.Error("job must stay active when its source is not listed")
	}
}

func TestStore_ReconcileStaleVacancies_SkipsAlreadyEndedJobs(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)

	rooster := seedSource(t, db, "Rooster")
	oldRun := seedCrawlerRun(t, db)
	currentRun := seedCrawlerRun(t, db)
	ended := seedActiveJob(t, db, "already ended", rooster, oldRun)
	endJob(t, db, ended)
	originalEnd := loadMeta(t, db, ended).EndDate

	affected, err := store.ReconcileStaleVacancies(context.Background(), currentRun, []string{"Rooster"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 0 {
		t.Errorf("expected 0 reconciled jobs, got %d", affected)
	}
	if got := loadMeta(t, db, ended).EndDate; got == nil || !got.Equal(*originalEnd) {
		t.Errorf("end date of an already ended job must not change, got %v want %v", got, originalEnd)
	}
}

func TestStore_ReconcileStaleVacancies_UnknownRunReturnsError(t *testing.T) {
	db := SetupTestDB(t)
	store := repositories.NewStore(db)

	_, err := store.ReconcileStaleVacancies(context.Background(), 9999, []string{"Rooster"})
	if err == nil {
		t.Fatal("expected an error for a crawler run that does not exist")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}
