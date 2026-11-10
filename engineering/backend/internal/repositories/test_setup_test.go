package repositories_test

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"marketlens-go-backend/internal/models"
)

// Shared test setup for the repositories package. The file name ends in
// _test.go so none of this is compiled into the production binary.

// ---------- Test DB setup ----------

// SetupTestDB opens a fresh, isolated in-memory SQLite database for a single
// test, runs AutoMigrate against it (in dependency order), and registers
// cleanup so the connection is closed when the test finishes.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// A unique DSN per test avoids two tests sharing the same in-memory
	// database when run in parallel.
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", url.PathEscape(t.Name()))

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}

	// Pin to a single connection — SQLite's in-memory DB only persists as
	// long as at least one connection stays open, and GORM pools connections
	// by default.
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get generic sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	// Register the cleanup before migrating: if the migration fails, t.Fatalf
	// stops the test, and a cleanup registered after it would never be added
	t.Cleanup(func() { sqlDB.Close() })

	if err := migrateAll(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	return db
}

// migrateAll runs AutoMigrate in dependency order: parent/lookup tables
// first, then tables that hold a foreign key into them.
func migrateAll(db *gorm.DB) error {
	return db.AutoMigrate(
		// --- Lookup / reference tables (no FK dependencies) ---
		&models.Employer{},
		&models.JobType{},
		&models.Skill{},
		&models.AiVersion{},
		&models.EducationLevel{},
		&models.Source{},
		&models.Experience{},
		&models.GeoData{},
		&models.Formality{},
		&models.Gender{},
		&models.VocationalEducation{},
		&models.EmploymentSector{},
		&models.CrawlerRun{},

		// --- Occupation classification hierarchy (parent before child) ---
		&models.MajorGroup{},
		&models.SubMajorGroup{},
		&models.MinorGroup{},
		&models.UnitGroup{},
		&models.OccupationGroup{},

		// --- Industry classification hierarchy (parent before child) ---
		&models.IndustrySector{},
		&models.IndustryDivision{},
		&models.IndustryGroup{},
		&models.IndustryClass{},
		&models.IndustrySubclass{},

		// --- Core job tables (depend on everything above) ---
		&models.JobPost{},
		&models.JobMetaData{},
		&models.LshIndex{},
	)
}

// ---------- Seed and query helpers ----------

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
