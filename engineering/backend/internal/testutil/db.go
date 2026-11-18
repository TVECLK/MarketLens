// Package testutil holds helpers shared by the test suites of several
// packages (repositories, ingestion, ...). It is only imported from
// _test.go files, so none of it ends up in the production binaries.
package testutil

import (
	"fmt"
	"net/url"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"marketlens-go-backend/internal/models"
)

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
