package repositories_test

import (
	"testing"
	"time"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/models"
)

// Seed and query helpers for the repositories package. The DB itself
// comes from testutil.SetupTestDB, shared with other packages' tests. The
// file name ends in _test.go so none of this is compiled into the
// production binary.

// ---------- Seed and query helpers ----------

// seedSource creates an active auto (crawler) source.
func seedSource(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()
	return seedSourceOf(t, db, name, models.SourceTypeAuto, true)
}

func seedSourceOf(t *testing.T, db *gorm.DB, name string, sourceType models.SourceType, isActive bool) uint {
	t.Helper()
	src := models.Source{Name: name, Type: sourceType}
	if err := db.Create(&src).Error; err != nil {
		t.Fatalf("failed to create source %q: %v", name, err)
	}
	// Set separately: Create skips is_active = false because of its default:true tag.
	if err := db.Model(&src).Update("is_active", isActive).Error; err != nil {
		t.Fatalf("failed to set is_active on source %q: %v", name, err)
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

// seedActiveJob creates an active (end_date IS NULL) job with the given number
// of vacancies for the given source and crawler run, with one LSH row per bucket
// key. A nil sourceID or runID leaves that column NULL.
func seedActiveJob(t *testing.T, db *gorm.DB, role string, noOfVacancies int, sourceID, runID *uint, bucketKeys ...string) uint {
	t.Helper()
	job := models.JobPost{
		JobRole:       role,
		WorkMode:      models.WorkModeOnsite,
		NoOfVacancies: noOfVacancies,
		MetaData: models.JobMetaData{
			SourceID:     sourceID,
			CrawlerRunID: runID,
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

func seedProvince(t *testing.T, db *gorm.DB, province string) uint {
	t.Helper()
	geo := models.GeoData{Province: province}
	if err := db.Create(&geo).Error; err != nil {
		t.Fatalf("failed to create geo data %q: %v", province, err)
	}
	return geo.ID
}
