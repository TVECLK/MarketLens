package repositories_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

// setupReconcileDB opens an isolated in-memory SQLite database containing only
// the tables the reconcile query touches.
func setupReconcileDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get generic sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })

	if err := db.AutoMigrate(
		&models.Source{},
		&models.CrawlerRun{},
		&models.JobPost{},
		&models.JobMetaData{},
		&models.LshIndex{},
	); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}
	return db
}

func seedReconcileSource(t *testing.T, db *gorm.DB, name string) uint {
	t.Helper()
	src := models.Source{Name: name}
	if err := db.Create(&src).Error; err != nil {
		t.Fatalf("failed to create source %q: %v", name, err)
	}
	return src.ID
}

func seedReconcileRun(t *testing.T, db *gorm.DB) uint {
	t.Helper()
	run := models.CrawlerRun{Status: "RUNNING"}
	if err := db.Create(&run).Error; err != nil {
		t.Fatalf("failed to create crawler run: %v", err)
	}
	return run.ID
}

// seedReconcileJob creates an active job (plus one LSH row) for the given source and run.
func seedReconcileJob(t *testing.T, db *gorm.DB, role string, sourceID, runID uint) uint {
	t.Helper()
	job := models.JobPost{
		JobRole:       role,
		WorkMode:      models.WorkMode("onsite"),
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
	lsh := models.LshIndex{BucketKey: fmt.Sprintf("bucket_%d", job.ID), BandNo: 0, JobPostID: job.ID}
	if err := db.Create(&lsh).Error; err != nil {
		t.Fatalf("failed to create lsh row for %q: %v", role, err)
	}
	return job.ID
}

func reconcileEndDate(t *testing.T, db *gorm.DB, jobID uint) *time.Time {
	t.Helper()
	var meta models.JobMetaData
	if err := db.Where("job_post_id = ?", jobID).First(&meta).Error; err != nil {
		t.Fatalf("failed to load meta data for job %d: %v", jobID, err)
	}
	return meta.EndDate
}

func reconcileLshCount(t *testing.T, db *gorm.DB, jobID uint) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&models.LshIndex{}).Where("job_post_id = ?", jobID).Count(&n).Error; err != nil {
		t.Fatalf("failed to count lsh rows for job %d: %v", jobID, err)
	}
	return n
}

func TestStore_ReconcileStaleVacancies_OnlyListedSources(t *testing.T) {
	db := setupReconcileDB(t)
	store := repositories.NewStore(db)

	rooster := seedReconcileSource(t, db, "Rooster")
	ikman := seedReconcileSource(t, db, "Ikman")
	oldRun := seedReconcileRun(t, db)
	currentRun := seedReconcileRun(t, db)

	staleListed := seedReconcileJob(t, db, "stale rooster", rooster, oldRun)
	staleUnlisted := seedReconcileJob(t, db, "stale ikman", ikman, oldRun)
	currentListed := seedReconcileJob(t, db, "current rooster", rooster, currentRun)

	affected, err := store.ReconcileStaleVacancies(context.Background(), currentRun, []string{"Rooster"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 1 {
		t.Errorf("expected 1 reconciled job, got %d", affected)
	}

	if reconcileEndDate(t, db, staleListed) == nil {
		t.Error("stale job of a listed source should have been expired")
	}
	if n := reconcileLshCount(t, db, staleListed); n != 0 {
		t.Errorf("stale job of a listed source should lose its LSH rows, got %d", n)
	}

	if reconcileEndDate(t, db, staleUnlisted) != nil {
		t.Error("stale job of an unlisted source must stay active")
	}
	if n := reconcileLshCount(t, db, staleUnlisted); n != 1 {
		t.Errorf("stale job of an unlisted source must keep its LSH rows, got %d", n)
	}

	if reconcileEndDate(t, db, currentListed) != nil {
		t.Error("job seen in the current run must stay active")
	}
	if n := reconcileLshCount(t, db, currentListed); n != 1 {
		t.Errorf("job seen in the current run must keep its LSH rows, got %d", n)
	}
}

func TestStore_ReconcileStaleVacancies_UnknownSourceIsNoOp(t *testing.T) {
	db := setupReconcileDB(t)
	store := repositories.NewStore(db)

	rooster := seedReconcileSource(t, db, "Rooster")
	oldRun := seedReconcileRun(t, db)
	currentRun := seedReconcileRun(t, db)
	stale := seedReconcileJob(t, db, "stale rooster", rooster, oldRun)

	affected, err := store.ReconcileStaleVacancies(context.Background(), currentRun, []string{"DoesNotExist"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if affected != 0 {
		t.Errorf("expected 0 reconciled jobs, got %d", affected)
	}
	if reconcileEndDate(t, db, stale) != nil {
		t.Error("job must stay active when its source is not listed")
	}
}
