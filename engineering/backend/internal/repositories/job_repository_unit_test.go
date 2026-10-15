package repositories_test

import (
	"errors"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

// ---------- Test DB setup ----------

// SetupTestDB opens a fresh, isolated in-memory SQLite database for a single
// test, runs AutoMigrate against it (in dependency order), and registers
// cleanup so the connection is closed when the test finishes.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// A unique DSN per test avoids two tests sharing the same in-memory
	// database when run in parallel.
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", t.Name())

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

	if err := migrateAll(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	t.Cleanup(func() {
		sqlDB.Close()
	})

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

// ---------- CreateEducationLevel ----------

func TestStore_CreateEducationLevel(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.EducationLevel{Level: "Bachelor's Degree"}

	if err := repo.CreateEducationLevel(item); err != nil {
		t.Fatalf("CreateEducationLevel returned error: %v", err)
	}

	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}

	// Confirm it actually landed in the DB, not just in the in-memory struct
	var fromDB models.EducationLevel
	if err := db.First(&fromDB, item.ID).Error; err != nil {
		t.Fatalf("expected to find created row in DB, got error: %v", err)
	}
	if fromDB.Level != "Bachelor's Degree" {
		t.Fatalf("expected Level %q, got %q", "Bachelor's Degree", fromDB.Level)
	}
}

// ---------- GetAllEducationLevels ----------

func TestStore_GetAllEducationLevels(t *testing.T) {
	t.Run("empty table returns empty slice, not error", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		items, err := repo.GetAllEducationLevels()
		if err != nil {
			t.Fatalf("expected no error on empty table, got: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected 0 items, got %d", len(items))
		}
	})

	t.Run("returns all seeded rows", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seed := []models.EducationLevel{
			{Level: "Diploma"},
			{Level: "Bachelor's Degree"},
			{Level: "Master's Degree"},
		}
		for i := range seed {
			if err := db.Create(&seed[i]).Error; err != nil {
				t.Fatalf("failed to seed row: %v", err)
			}
		}

		items, err := repo.GetAllEducationLevels()
		if err != nil {
			t.Fatalf("GetAllEducationLevels returned error: %v", err)
		}
		if len(items) != len(seed) {
			t.Fatalf("expected %d items, got %d", len(seed), len(items))
		}
	})
}

// ---------- GetEducationLevelByID ----------

func TestStore_GetEducationLevelByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.EducationLevel{Level: "PhD"}
		if err := db.Create(&seeded).Error; err != nil {
			t.Fatalf("failed to seed row: %v", err)
		}

		got, err := repo.GetEducationLevelByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetEducationLevelByID returned error: %v", err)
		}
		if got.ID != seeded.ID || got.Level != "PhD" {
			t.Fatalf("expected %+v, got %+v", seeded, got)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetEducationLevelByID(999999)
		if err == nil {
			t.Fatalf("expected an error for non-existent ID, got nil")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

// ---------- UpdateEducationLevel ----------

func TestStore_UpdateEducationLevel(t *testing.T) {
	t.Run("updates an existing row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.EducationLevel{Level: "Diploma"}
		if err := db.Create(&seeded).Error; err != nil {
			t.Fatalf("failed to seed row: %v", err)
		}

		updated, err := repo.UpdateEducationLevel(seeded.ID, map[string]interface{}{
			"level": "Advanced Diploma",
		})
		if err != nil {
			t.Fatalf("UpdateEducationLevel returned error: %v", err)
		}
		if updated.Level != "Advanced Diploma" {
			t.Fatalf("expected updated Level %q, got %q", "Advanced Diploma", updated.Level)
		}

		// Confirm the change actually persisted, not just returned in-memory
		var fromDB models.EducationLevel
		if err := db.First(&fromDB, seeded.ID).Error; err != nil {
			t.Fatalf("failed to reload row: %v", err)
		}
		if fromDB.Level != "Advanced Diploma" {
			t.Fatalf("expected persisted Level %q, got %q", "Advanced Diploma", fromDB.Level)
		}
	})

	t.Run("non-existent ID returns error and does not panic", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.UpdateEducationLevel(999999, map[string]interface{}{
			"level": "Should Not Apply",
		})
		if err == nil {
			t.Fatalf("expected an error for non-existent ID, got nil")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

// ---------- CreateFormality ----------

func TestStore_CreateFormality(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.Formality{FormalityType: "Formal"}

	if err := repo.CreateFormality(item); err != nil {
		t.Fatalf("CreateFormality returned error: %v", err)
	}

	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}

	var fromDB models.Formality
	if err := db.First(&fromDB, item.ID).Error; err != nil {
		t.Fatalf("expected to find created row in DB, got error: %v", err)
	}
	if fromDB.FormalityType != "Formal" {
		t.Fatalf("expected FormalityType %q, got %q", "Formal", fromDB.FormalityType)
	}
}

// ---------- GetAllFormalities ----------

func TestStore_GetAllFormalities(t *testing.T) {
	t.Run("empty table returns empty slice, not error", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		items, err := repo.GetAllFormalities()
		if err != nil {
			t.Fatalf("expected no error on empty table, got: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected 0 items, got %d", len(items))
		}
	})

	t.Run("returns all seeded rows", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seed := []models.Formality{
			{FormalityType: "Formal"},
			{FormalityType: "Informal"},
		}
		for i := range seed {
			if err := db.Create(&seed[i]).Error; err != nil {
				t.Fatalf("failed to seed row: %v", err)
			}
		}

		items, err := repo.GetAllFormalities()
		if err != nil {
			t.Fatalf("GetAllFormalities returned error: %v", err)
		}
		if len(items) != len(seed) {
			t.Fatalf("expected %d items, got %d", len(seed), len(items))
		}
	})
}

// ---------- GetFormalityByID ----------

func TestStore_GetFormalityByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.Formality{FormalityType: "Formal"}
		if err := db.Create(&seeded).Error; err != nil {
			t.Fatalf("failed to seed row: %v", err)
		}

		got, err := repo.GetFormalityByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetFormalityByID returned error: %v", err)
		}
		if got.ID != seeded.ID || got.FormalityType != "Formal" {
			t.Fatalf("expected %+v, got %+v", seeded, got)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetFormalityByID(999999)
		if err == nil {
			t.Fatalf("expected an error for non-existent ID, got nil")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

// ---------- UpdateFormality ----------

func TestStore_UpdateFormality(t *testing.T) {
	t.Run("updates an existing row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.Formality{FormalityType: "Formal"}
		if err := db.Create(&seeded).Error; err != nil {
			t.Fatalf("failed to seed row: %v", err)
		}

		updated, err := repo.UpdateFormality(seeded.ID, map[string]interface{}{
			"formality_type": "Semi-Formal",
		})
		if err != nil {
			t.Fatalf("UpdateFormality returned error: %v", err)
		}
		if updated.FormalityType != "Semi-Formal" {
			t.Fatalf("expected updated FormalityType %q, got %q", "Semi-Formal", updated.FormalityType)
		}

		var fromDB models.Formality
		if err := db.First(&fromDB, seeded.ID).Error; err != nil {
			t.Fatalf("failed to reload row: %v", err)
		}
		if fromDB.FormalityType != "Semi-Formal" {
			t.Fatalf("expected persisted FormalityType %q, got %q", "Semi-Formal", fromDB.FormalityType)
		}
	})

	t.Run("non-existent ID returns error and does not panic", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.UpdateFormality(999999, map[string]interface{}{
			"formality_type": "Should Not Apply",
		})
		if err == nil {
			t.Fatalf("expected an error for non-existent ID, got nil")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

// ---------- DeleteFormality ----------

func TestStore_DeleteFormality(t *testing.T) {
	t.Run("deletes an existing row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.Formality{FormalityType: "Formal"}
		if err := db.Create(&seeded).Error; err != nil {
			t.Fatalf("failed to seed row: %v", err)
		}

		if err := repo.DeleteFormality(seeded.ID); err != nil {
			t.Fatalf("DeleteFormality returned error: %v", err)
		}

		var count int64
		db.Model(&models.Formality{}).Where("id = ?", seeded.ID).Count(&count)
		if count != 0 {
			t.Fatalf("expected row to be deleted, but %d rows still match id %d", count, seeded.ID)
		}
	})

	t.Run("non-existent ID does not return an error (GORM delete is idempotent)", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		if err := repo.DeleteFormality(999999); err != nil {
			t.Fatalf("expected no error deleting a non-existent ID, got: %v", err)
		}
	})
}

// ================= Gender =================

func TestStore_CreateGender(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.Gender{GenderType: "Male"}
	if err := repo.CreateGender(item); err != nil {
		t.Fatalf("CreateGender returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}

	var fromDB models.Gender
	if err := db.First(&fromDB, item.ID).Error; err != nil {
		t.Fatalf("expected to find created row in DB, got error: %v", err)
	}
	if fromDB.GenderType != "Male" {
		t.Fatalf("expected GenderType %q, got %q", "Male", fromDB.GenderType)
	}
}

func TestStore_GetAllGenders(t *testing.T) {
	t.Run("empty table returns empty slice, not error", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		items, err := repo.GetAllGenders()
		if err != nil {
			t.Fatalf("expected no error on empty table, got: %v", err)
		}
		if len(items) != 0 {
			t.Fatalf("expected 0 items, got %d", len(items))
		}
	})

	t.Run("returns all seeded rows", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seed := []models.Gender{{GenderType: "Male"}, {GenderType: "Female"}}
		for i := range seed {
			if err := db.Create(&seed[i]).Error; err != nil {
				t.Fatalf("failed to seed row: %v", err)
			}
		}

		items, err := repo.GetAllGenders()
		if err != nil {
			t.Fatalf("GetAllGenders returned error: %v", err)
		}
		if len(items) != len(seed) {
			t.Fatalf("expected %d items, got %d", len(seed), len(items))
		}
	})
}

func TestStore_GetGenderByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.Gender{GenderType: "Male"}
		db.Create(&seeded)

		got, err := repo.GetGenderByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetGenderByID returned error: %v", err)
		}
		if got.ID != seeded.ID || got.GenderType != "Male" {
			t.Fatalf("expected %+v, got %+v", seeded, got)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetGenderByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateGender(t *testing.T) {
	t.Run("updates an existing row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.Gender{GenderType: "Male"}
		db.Create(&seeded)

		updated, err := repo.UpdateGender(seeded.ID, map[string]interface{}{"gender_type": "Female"})
		if err != nil {
			t.Fatalf("UpdateGender returned error: %v", err)
		}
		if updated.GenderType != "Female" {
			t.Fatalf("expected GenderType %q, got %q", "Female", updated.GenderType)
		}

		var fromDB models.Gender
		db.First(&fromDB, seeded.ID)
		if fromDB.GenderType != "Female" {
			t.Fatalf("expected persisted GenderType %q, got %q", "Female", fromDB.GenderType)
		}
	})

	t.Run("non-existent ID returns error and does not panic", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.UpdateGender(999999, map[string]interface{}{"gender_type": "X"})
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_DeleteGender(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.Gender{GenderType: "Male"}
	db.Create(&seeded)

	if err := repo.DeleteGender(seeded.ID); err != nil {
		t.Fatalf("DeleteGender returned error: %v", err)
	}

	var count int64
	db.Model(&models.Gender{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, but %d rows still match id %d", count, seeded.ID)
	}
}

// ================= EmploymentSector =================

func TestStore_CreateEmploymentSector(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.EmploymentSector{Sector: "Private"}
	if err := repo.CreateEmploymentSector(item); err != nil {
		t.Fatalf("CreateEmploymentSector returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllEmploymentSectors(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	db.Create(&models.EmploymentSector{Sector: "Private"})
	db.Create(&models.EmploymentSector{Sector: "Government"})

	items, err := repo.GetAllEmploymentSectors()
	if err != nil {
		t.Fatalf("GetAllEmploymentSectors returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestStore_GetEmploymentSectorByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.EmploymentSector{Sector: "Private"}
		db.Create(&seeded)

		got, err := repo.GetEmploymentSectorByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetEmploymentSectorByID returned error: %v", err)
		}
		if got.Sector != "Private" {
			t.Fatalf("expected Sector %q, got %q", "Private", got.Sector)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetEmploymentSectorByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateEmploymentSector(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.EmploymentSector{Sector: "Private"}
	db.Create(&seeded)

	updated, err := repo.UpdateEmploymentSector(seeded.ID, map[string]interface{}{"sector": "Government"})
	if err != nil {
		t.Fatalf("UpdateEmploymentSector returned error: %v", err)
	}
	if updated.Sector != "Government" {
		t.Fatalf("expected Sector %q, got %q", "Government", updated.Sector)
	}
}

func TestStore_DeleteEmploymentSector(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.EmploymentSector{Sector: "Private"}
	db.Create(&seeded)

	if err := repo.DeleteEmploymentSector(seeded.ID); err != nil {
		t.Fatalf("DeleteEmploymentSector returned error: %v", err)
	}

	var count int64
	db.Model(&models.EmploymentSector{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= VocationalEducation =================

func TestStore_CreateVocationalEducation(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.VocationalEducation{Level: "NVQ 3"}
	if err := repo.CreateVocationalEducation(item); err != nil {
		t.Fatalf("CreateVocationalEducation returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllVocationalEducations(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	db.Create(&models.VocationalEducation{Level: "NVQ 3"})
	db.Create(&models.VocationalEducation{Level: "NVQ 4"})

	items, err := repo.GetAllVocationalEducations()
	if err != nil {
		t.Fatalf("GetAllVocationalEducations returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestStore_GetVocationalEducationByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.VocationalEducation{Level: "NVQ 3"}
		db.Create(&seeded)

		got, err := repo.GetVocationalEducationByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetVocationalEducationByID returned error: %v", err)
		}
		if got.Level != "NVQ 3" {
			t.Fatalf("expected Level %q, got %q", "NVQ 3", got.Level)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetVocationalEducationByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateVocationalEducation(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.VocationalEducation{Level: "NVQ 3"}
	db.Create(&seeded)

	updated, err := repo.UpdateVocationalEducation(seeded.ID, map[string]interface{}{"level": "NVQ 5"})
	if err != nil {
		t.Fatalf("UpdateVocationalEducation returned error: %v", err)
	}
	if updated.Level != "NVQ 5" {
		t.Fatalf("expected Level %q, got %q", "NVQ 5", updated.Level)
	}
}

func TestStore_DeleteVocationalEducation(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.VocationalEducation{Level: "NVQ 3"}
	db.Create(&seeded)

	if err := repo.DeleteVocationalEducation(seeded.ID); err != nil {
		t.Fatalf("DeleteVocationalEducation returned error: %v", err)
	}

	var count int64
	db.Model(&models.VocationalEducation{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= Experience =================
// NOTE: your pasted code has no GetAllExperiences method, so there's no
// "get all" test here — only Create/GetByID/Update/Delete are covered.

func TestStore_CreateExperience(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.Experience{Name: "Entry Level"}
	if err := repo.CreateExperience(item); err != nil {
		t.Fatalf("CreateExperience returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetExperienceByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.Experience{Name: "Entry Level"}
		db.Create(&seeded)

		got, err := repo.GetExperienceByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetExperienceByID returned error: %v", err)
		}
		if got.Name != "Entry Level" {
			t.Fatalf("expected Name %q, got %q", "Entry Level", got.Name)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetExperienceByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateExperience(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.Experience{Name: "Entry Level"}
	db.Create(&seeded)

	updated, err := repo.UpdateExperience(seeded.ID, map[string]interface{}{"name": "Senior Level"})
	if err != nil {
		t.Fatalf("UpdateExperience returned error: %v", err)
	}
	if updated.Name != "Senior Level" {
		t.Fatalf("expected Name %q, got %q", "Senior Level", updated.Name)
	}
}

func TestStore_DeleteExperience(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.Experience{Name: "Entry Level"}
	db.Create(&seeded)

	if err := repo.DeleteExperience(seeded.ID); err != nil {
		t.Fatalf("DeleteExperience returned error: %v", err)
	}

	var count int64
	db.Model(&models.Experience{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= MajorGroup =================

func TestStore_CreateMajorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.MajorGroup{Name: "Managers", Code: "1"}
	if err := repo.CreateMajorGroup(item); err != nil {
		t.Fatalf("CreateMajorGroup returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllMajorGroups(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	db.Create(&models.MajorGroup{Name: "Managers", Code: "1"})
	db.Create(&models.MajorGroup{Name: "Professionals", Code: "2"})

	items, err := repo.GetAllMajorGroups()
	if err != nil {
		t.Fatalf("GetAllMajorGroups returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestStore_GetMajorGroupByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.MajorGroup{Name: "Managers", Code: "1"}
		db.Create(&seeded)

		got, err := repo.GetMajorGroupByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetMajorGroupByID returned error: %v", err)
		}
		if got.Name != "Managers" {
			t.Fatalf("expected Name %q, got %q", "Managers", got.Name)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetMajorGroupByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateMajorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&seeded)

	updated, err := repo.UpdateMajorGroup(seeded.ID, map[string]interface{}{"name": "Senior Managers"})
	if err != nil {
		t.Fatalf("UpdateMajorGroup returned error: %v", err)
	}
	if updated.Name != "Senior Managers" {
		t.Fatalf("expected Name %q, got %q", "Senior Managers", updated.Name)
	}
}

func TestStore_DeleteMajorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&seeded)

	if err := repo.DeleteMajorGroup(seeded.ID); err != nil {
		t.Fatalf("DeleteMajorGroup returned error: %v", err)
	}

	var count int64
	db.Model(&models.MajorGroup{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= SubMajorGroup (has Preload("MajorGroup")) =================

func TestStore_CreateSubMajorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)

	item := &models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"}
	if err := repo.CreateSubMajorGroup(item); err != nil {
		t.Fatalf("CreateSubMajorGroup returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllSubMajorGroups(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	db.Create(&models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"})

	items, err := repo.GetAllSubMajorGroups()
	if err != nil {
		t.Fatalf("GetAllSubMajorGroups returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	// Preload("MajorGroup") should populate the parent, not just MajorGroupID
	if items[0].MajorGroup == nil {
		t.Fatalf("expected MajorGroup to be preloaded, got nil")
	}
	if items[0].MajorGroup.Name != "Managers" {
		t.Fatalf("expected preloaded MajorGroup.Name %q, got %q", "Managers", items[0].MajorGroup.Name)
	}
}

func TestStore_GetSubMajorGroupByID(t *testing.T) {
	t.Run("existing ID returns the row with MajorGroup preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		parent := models.MajorGroup{Name: "Managers", Code: "1"}
		db.Create(&parent)
		seeded := models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"}
		db.Create(&seeded)

		got, err := repo.GetSubMajorGroupByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetSubMajorGroupByID returned error: %v", err)
		}
		if got.MajorGroup == nil || got.MajorGroup.Name != "Managers" {
			t.Fatalf("expected preloaded MajorGroup.Name %q, got %+v", "Managers", got.MajorGroup)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetSubMajorGroupByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateSubMajorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	seeded := models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&seeded)

	updated, err := repo.UpdateSubMajorGroup(seeded.ID, map[string]interface{}{"name": "Senior Executives"})
	if err != nil {
		t.Fatalf("UpdateSubMajorGroup returned error: %v", err)
	}
	if updated.Name != "Senior Executives" {
		t.Fatalf("expected Name %q, got %q", "Senior Executives", updated.Name)
	}
}

func TestStore_DeleteSubMajorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	parent := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&parent)
	seeded := models.SubMajorGroup{MajorGroupID: parent.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&seeded)

	if err := repo.DeleteSubMajorGroup(seeded.ID); err != nil {
		t.Fatalf("DeleteSubMajorGroup returned error: %v", err)
	}

	var count int64
	db.Model(&models.SubMajorGroup{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= MinorGroup (has Preload("SubMajorGroup")) =================

func TestStore_CreateMinorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)

	item := &models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"}
	if err := repo.CreateMinorGroup(item); err != nil {
		t.Fatalf("CreateMinorGroup returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllMinorGroups(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	db.Create(&models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"})

	items, err := repo.GetAllMinorGroups()
	if err != nil {
		t.Fatalf("GetAllMinorGroups returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].SubMajorGroup == nil || items[0].SubMajorGroup.Name != "Chief Executives" {
		t.Fatalf("expected preloaded SubMajorGroup.Name %q, got %+v", "Chief Executives", items[0].SubMajorGroup)
	}
}

func TestStore_GetMinorGroupByID(t *testing.T) {
	t.Run("existing ID returns the row with SubMajorGroup preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		major := models.MajorGroup{Name: "Managers", Code: "1"}
		db.Create(&major)
		sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
		db.Create(&sub)
		seeded := models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"}
		db.Create(&seeded)

		got, err := repo.GetMinorGroupByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetMinorGroupByID returned error: %v", err)
		}
		if got.SubMajorGroup == nil || got.SubMajorGroup.Name != "Chief Executives" {
			t.Fatalf("expected preloaded SubMajorGroup.Name %q, got %+v", "Chief Executives", got.SubMajorGroup)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetMinorGroupByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateMinorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	seeded := models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"}
	db.Create(&seeded)

	updated, err := repo.UpdateMinorGroup(seeded.ID, map[string]interface{}{"name": "Senior Legislators"})
	if err != nil {
		t.Fatalf("UpdateMinorGroup returned error: %v", err)
	}
	if updated.Name != "Senior Legislators" {
		t.Fatalf("expected Name %q, got %q", "Senior Legislators", updated.Name)
	}
}

func TestStore_DeleteMinorGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	seeded := models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"}
	db.Create(&seeded)

	if err := repo.DeleteMinorGroup(seeded.ID); err != nil {
		t.Fatalf("DeleteMinorGroup returned error: %v", err)
	}

	var count int64
	db.Model(&models.MinorGroup{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= UnitGroup (has Preload("MinorGroup")) =================

func seedUnitGroupChain(t *testing.T, db *gorm.DB) models.MinorGroup {
	t.Helper()
	major := models.MajorGroup{Name: "Managers", Code: "1"}
	db.Create(&major)
	sub := models.SubMajorGroup{MajorGroupID: major.ID, Name: "Chief Executives", Code: "11"}
	db.Create(&sub)
	minor := models.MinorGroup{SubMajorGroupID: sub.ID, Name: "Legislators", Code: "111"}
	db.Create(&minor)
	return minor
}

func TestStore_CreateUnitGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	minor := seedUnitGroupChain(t, db)

	item := &models.UnitGroup{MinorGroupID: minor.ID, Name: "Senior Officials", Code: "1111"}
	if err := repo.CreateUnitGroup(item); err != nil {
		t.Fatalf("CreateUnitGroup returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllUnitGroups(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	minor := seedUnitGroupChain(t, db)
	db.Create(&models.UnitGroup{MinorGroupID: minor.ID, Name: "Senior Officials", Code: "1111"})

	items, err := repo.GetAllUnitGroups()
	if err != nil {
		t.Fatalf("GetAllUnitGroups returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].MinorGroup == nil || items[0].MinorGroup.Name != "Legislators" {
		t.Fatalf("expected preloaded MinorGroup.Name %q, got %+v", "Legislators", items[0].MinorGroup)
	}
}

func TestStore_GetUnitGroupByID(t *testing.T) {
	t.Run("existing ID returns the row with MinorGroup preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		minor := seedUnitGroupChain(t, db)
		seeded := models.UnitGroup{MinorGroupID: minor.ID, Name: "Senior Officials", Code: "1111"}
		db.Create(&seeded)

		got, err := repo.GetUnitGroupByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetUnitGroupByID returned error: %v", err)
		}
		if got.MinorGroup == nil || got.MinorGroup.Name != "Legislators" {
			t.Fatalf("expected preloaded MinorGroup.Name %q, got %+v", "Legislators", got.MinorGroup)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetUnitGroupByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateUnitGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	minor := seedUnitGroupChain(t, db)
	seeded := models.UnitGroup{MinorGroupID: minor.ID, Name: "Senior Officials", Code: "1111"}
	db.Create(&seeded)

	updated, err := repo.UpdateUnitGroup(seeded.ID, map[string]interface{}{"name": "Senior Government Officials"})
	if err != nil {
		t.Fatalf("UpdateUnitGroup returned error: %v", err)
	}
	if updated.Name != "Senior Government Officials" {
		t.Fatalf("expected Name %q, got %q", "Senior Government Officials", updated.Name)
	}
}

func TestStore_DeleteUnitGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	minor := seedUnitGroupChain(t, db)
	seeded := models.UnitGroup{MinorGroupID: minor.ID, Name: "Senior Officials", Code: "1111"}
	db.Create(&seeded)

	if err := repo.DeleteUnitGroup(seeded.ID); err != nil {
		t.Fatalf("DeleteUnitGroup returned error: %v", err)
	}

	var count int64
	db.Model(&models.UnitGroup{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= OccupationGroup (Preload("UnitGroup") + limit/offset/total) =================

func seedOccupationGroupChain(t *testing.T, db *gorm.DB) models.UnitGroup {
	t.Helper()
	minor := seedUnitGroupChain(t, db)
	unit := models.UnitGroup{MinorGroupID: minor.ID, Name: "Senior Officials", Code: "1111"}
	db.Create(&unit)
	return unit
}

func TestStore_CreateOccupationGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	unit := seedOccupationGroupChain(t, db)

	item := &models.OccupationGroup{UnitGroupID: unit.ID, Name: "Legislator", Code: "11111"}
	if err := repo.CreateOccupationGroup(item); err != nil {
		t.Fatalf("CreateOccupationGroup returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllOccupationGroups(t *testing.T) {
	t.Run("total reflects all rows regardless of limit", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		unit := seedOccupationGroupChain(t, db)
		for i := 0; i < 5; i++ {
			db.Create(&models.OccupationGroup{UnitGroupID: unit.ID, Name: "Group", Code: "X"})
		}

		items, total, err := repo.GetAllOccupationGroups(2, 0)
		if err != nil {
			t.Fatalf("GetAllOccupationGroups returned error: %v", err)
		}
		if total != 5 {
			t.Fatalf("expected total 5, got %d", total)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items with limit=2, got %d", len(items))
		}
	})

	t.Run("offset skips the correct number of rows, ordered by id", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		unit := seedOccupationGroupChain(t, db)
		var seeded []models.OccupationGroup
		for i := 0; i < 3; i++ {
			g := models.OccupationGroup{UnitGroupID: unit.ID, Name: "Group", Code: "X"}
			db.Create(&g)
			seeded = append(seeded, g)
		}

		items, _, err := repo.GetAllOccupationGroups(0, 2)
		if err != nil {
			t.Fatalf("GetAllOccupationGroups returned error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item after offsetting past 2 of 3, got %d", len(items))
		}
		if items[0].ID != seeded[2].ID {
			t.Fatalf("expected remaining item to be the 3rd seeded row (id %d), got id %d", seeded[2].ID, items[0].ID)
		}
	})

	t.Run("limit=0 and offset=0 returns all rows", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		unit := seedOccupationGroupChain(t, db)
		for i := 0; i < 4; i++ {
			db.Create(&models.OccupationGroup{UnitGroupID: unit.ID, Name: "Group", Code: "X"})
		}

		items, total, err := repo.GetAllOccupationGroups(0, 0)
		if err != nil {
			t.Fatalf("GetAllOccupationGroups returned error: %v", err)
		}
		if len(items) != 4 || total != 4 {
			t.Fatalf("expected 4 items and total 4, got %d items, total %d", len(items), total)
		}
	})

	t.Run("UnitGroup is preloaded on each item", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		unit := seedOccupationGroupChain(t, db)
		db.Create(&models.OccupationGroup{UnitGroupID: unit.ID, Name: "Legislator", Code: "11111"})

		items, _, err := repo.GetAllOccupationGroups(0, 0)
		if err != nil {
			t.Fatalf("GetAllOccupationGroups returned error: %v", err)
		}
		if len(items) != 1 || items[0].UnitGroup == nil || items[0].UnitGroup.Name != "Senior Officials" {
			t.Fatalf("expected preloaded UnitGroup.Name %q, got %+v", "Senior Officials", items)
		}
	})
}

func TestStore_GetOccupationGroupByID(t *testing.T) {
	t.Run("existing ID returns the row with UnitGroup preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		unit := seedOccupationGroupChain(t, db)
		seeded := models.OccupationGroup{UnitGroupID: unit.ID, Name: "Legislator", Code: "11111"}
		db.Create(&seeded)

		got, err := repo.GetOccupationGroupByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetOccupationGroupByID returned error: %v", err)
		}
		if got.UnitGroup == nil || got.UnitGroup.Name != "Senior Officials" {
			t.Fatalf("expected preloaded UnitGroup.Name %q, got %+v", "Senior Officials", got.UnitGroup)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetOccupationGroupByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateOccupationGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	unit := seedOccupationGroupChain(t, db)
	seeded := models.OccupationGroup{UnitGroupID: unit.ID, Name: "Legislator", Code: "11111"}
	db.Create(&seeded)

	updated, err := repo.UpdateOccupationGroup(seeded.ID, map[string]interface{}{"name": "Senior Legislator"})
	if err != nil {
		t.Fatalf("UpdateOccupationGroup returned error: %v", err)
	}
	if updated.Name != "Senior Legislator" {
		t.Fatalf("expected Name %q, got %q", "Senior Legislator", updated.Name)
	}
}

func TestStore_DeleteOccupationGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	unit := seedOccupationGroupChain(t, db)
	seeded := models.OccupationGroup{UnitGroupID: unit.ID, Name: "Legislator", Code: "11111"}
	db.Create(&seeded)

	if err := repo.DeleteOccupationGroup(seeded.ID); err != nil {
		t.Fatalf("DeleteOccupationGroup returned error: %v", err)
	}

	var count int64
	db.Model(&models.OccupationGroup{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= IndustrySector =================

func TestStore_CreateIndustrySector(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	item := &models.IndustrySector{Name: "Agriculture", Code: "A"}
	if err := repo.CreateIndustrySector(item); err != nil {
		t.Fatalf("CreateIndustrySector returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllIndustrySectors(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	db.Create(&models.IndustrySector{Name: "Agriculture", Code: "A"})
	db.Create(&models.IndustrySector{Name: "Mining", Code: "B"})

	items, err := repo.GetAllIndustrySectors()
	if err != nil {
		t.Fatalf("GetAllIndustrySectors returned error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestStore_GetIndustrySectorByID(t *testing.T) {
	t.Run("existing ID returns the correct row", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		seeded := models.IndustrySector{Name: "Agriculture", Code: "A"}
		db.Create(&seeded)

		got, err := repo.GetIndustrySectorByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetIndustrySectorByID returned error: %v", err)
		}
		if got.Name != "Agriculture" {
			t.Fatalf("expected Name %q, got %q", "Agriculture", got.Name)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetIndustrySectorByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateIndustrySector(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&seeded)

	updated, err := repo.UpdateIndustrySector(seeded.ID, map[string]interface{}{"name": "Agri & Fisheries"})
	if err != nil {
		t.Fatalf("UpdateIndustrySector returned error: %v", err)
	}
	if updated.Name != "Agri & Fisheries" {
		t.Fatalf("expected Name %q, got %q", "Agri & Fisheries", updated.Name)
	}
}

func TestStore_DeleteIndustrySector(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)

	seeded := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&seeded)

	if err := repo.DeleteIndustrySector(seeded.ID); err != nil {
		t.Fatalf("DeleteIndustrySector returned error: %v", err)
	}

	var count int64
	db.Model(&models.IndustrySector{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= IndustryDivision (Preload("IndustrySector")) =================

func TestStore_CreateIndustryDivision(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)

	item := &models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	if err := repo.CreateIndustryDivision(item); err != nil {
		t.Fatalf("CreateIndustryDivision returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllIndustryDivisions(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	db.Create(&models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"})

	items, err := repo.GetAllIndustryDivisions()
	if err != nil {
		t.Fatalf("GetAllIndustryDivisions returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].IndustrySector == nil || items[0].IndustrySector.Name != "Agriculture" {
		t.Fatalf("expected preloaded IndustrySector.Name %q, got %+v", "Agriculture", items[0].IndustrySector)
	}
}

func TestStore_GetIndustryDivisionByID(t *testing.T) {
	t.Run("existing ID returns the row with IndustrySector preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
		db.Create(&sector)
		seeded := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
		db.Create(&seeded)

		got, err := repo.GetIndustryDivisionByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetIndustryDivisionByID returned error: %v", err)
		}
		if got.IndustrySector == nil || got.IndustrySector.Name != "Agriculture" {
			t.Fatalf("expected preloaded IndustrySector.Name %q, got %+v", "Agriculture", got.IndustrySector)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetIndustryDivisionByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateIndustryDivision(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	seeded := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&seeded)

	updated, err := repo.UpdateIndustryDivision(seeded.ID, map[string]interface{}{"name": "Arable Farming"})
	if err != nil {
		t.Fatalf("UpdateIndustryDivision returned error: %v", err)
	}
	if updated.Name != "Arable Farming" {
		t.Fatalf("expected Name %q, got %q", "Arable Farming", updated.Name)
	}
}

func TestStore_DeleteIndustryDivision(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	seeded := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&seeded)

	if err := repo.DeleteIndustryDivision(seeded.ID); err != nil {
		t.Fatalf("DeleteIndustryDivision returned error: %v", err)
	}

	var count int64
	db.Model(&models.IndustryDivision{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= IndustryGroup (Preload("IndustryDivision")) =================

func seedIndustryGroupChain(t *testing.T, db *gorm.DB) models.IndustryDivision {
	t.Helper()
	sector := models.IndustrySector{Name: "Agriculture", Code: "A"}
	db.Create(&sector)
	division := models.IndustryDivision{IndustrySectorID: sector.ID, Name: "Crop Farming", Code: "01"}
	db.Create(&division)
	return division
}

func TestStore_CreateIndustryGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	division := seedIndustryGroupChain(t, db)

	item := &models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"}
	if err := repo.CreateIndustryGroup(item); err != nil {
		t.Fatalf("CreateIndustryGroup returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllIndustryGroups(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	division := seedIndustryGroupChain(t, db)
	db.Create(&models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"})

	items, err := repo.GetAllIndustryGroups()
	if err != nil {
		t.Fatalf("GetAllIndustryGroups returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].IndustryDivision == nil || items[0].IndustryDivision.Name != "Crop Farming" {
		t.Fatalf("expected preloaded IndustryDivision.Name %q, got %+v", "Crop Farming", items[0].IndustryDivision)
	}
}

func TestStore_GetIndustryGroupByID(t *testing.T) {
	t.Run("existing ID returns the row with IndustryDivision preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		division := seedIndustryGroupChain(t, db)
		seeded := models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"}
		db.Create(&seeded)

		got, err := repo.GetIndustryGroupByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetIndustryGroupByID returned error: %v", err)
		}
		if got.IndustryDivision == nil || got.IndustryDivision.Name != "Crop Farming" {
			t.Fatalf("expected preloaded IndustryDivision.Name %q, got %+v", "Crop Farming", got.IndustryDivision)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetIndustryGroupByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateIndustryGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	division := seedIndustryGroupChain(t, db)
	seeded := models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"}
	db.Create(&seeded)

	updated, err := repo.UpdateIndustryGroup(seeded.ID, map[string]interface{}{"name": "Grain Growing"})
	if err != nil {
		t.Fatalf("UpdateIndustryGroup returned error: %v", err)
	}
	if updated.Name != "Grain Growing" {
		t.Fatalf("expected Name %q, got %q", "Grain Growing", updated.Name)
	}
}

func TestStore_DeleteIndustryGroup(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	division := seedIndustryGroupChain(t, db)
	seeded := models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"}
	db.Create(&seeded)

	if err := repo.DeleteIndustryGroup(seeded.ID); err != nil {
		t.Fatalf("DeleteIndustryGroup returned error: %v", err)
	}

	var count int64
	db.Model(&models.IndustryGroup{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= IndustryClass (Preload("IndustryGroup")) =================

func seedIndustryClassChain(t *testing.T, db *gorm.DB) models.IndustryGroup {
	t.Helper()
	division := seedIndustryGroupChain(t, db)
	group := models.IndustryGroup{IndustryDivisionID: division.ID, Name: "Cereal Growing", Code: "011"}
	db.Create(&group)
	return group
}

func TestStore_CreateIndustryClass(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	group := seedIndustryClassChain(t, db)

	item := &models.IndustryClass{IndustryGroupID: group.ID, Name: "Rice Growing", Code: "0111"}
	if err := repo.CreateIndustryClass(item); err != nil {
		t.Fatalf("CreateIndustryClass returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllIndustryClasses(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	group := seedIndustryClassChain(t, db)
	db.Create(&models.IndustryClass{IndustryGroupID: group.ID, Name: "Rice Growing", Code: "0111"})

	items, err := repo.GetAllIndustryClasses()
	if err != nil {
		t.Fatalf("GetAllIndustryClasses returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].IndustryGroup == nil || items[0].IndustryGroup.Name != "Cereal Growing" {
		t.Fatalf("expected preloaded IndustryGroup.Name %q, got %+v", "Cereal Growing", items[0].IndustryGroup)
	}
}

func TestStore_GetIndustryClassByID(t *testing.T) {
	t.Run("existing ID returns the row with IndustryGroup preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		group := seedIndustryClassChain(t, db)
		seeded := models.IndustryClass{IndustryGroupID: group.ID, Name: "Rice Growing", Code: "0111"}
		db.Create(&seeded)

		got, err := repo.GetIndustryClassByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetIndustryClassByID returned error: %v", err)
		}
		if got.IndustryGroup == nil || got.IndustryGroup.Name != "Cereal Growing" {
			t.Fatalf("expected preloaded IndustryGroup.Name %q, got %+v", "Cereal Growing", got.IndustryGroup)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetIndustryClassByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateIndustryClass(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	group := seedIndustryClassChain(t, db)
	seeded := models.IndustryClass{IndustryGroupID: group.ID, Name: "Rice Growing", Code: "0111"}
	db.Create(&seeded)

	updated, err := repo.UpdateIndustryClass(seeded.ID, map[string]interface{}{"name": "Paddy Growing"})
	if err != nil {
		t.Fatalf("UpdateIndustryClass returned error: %v", err)
	}
	if updated.Name != "Paddy Growing" {
		t.Fatalf("expected Name %q, got %q", "Paddy Growing", updated.Name)
	}
}

func TestStore_DeleteIndustryClass(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	group := seedIndustryClassChain(t, db)
	seeded := models.IndustryClass{IndustryGroupID: group.ID, Name: "Rice Growing", Code: "0111"}
	db.Create(&seeded)

	if err := repo.DeleteIndustryClass(seeded.ID); err != nil {
		t.Fatalf("DeleteIndustryClass returned error: %v", err)
	}

	var count int64
	db.Model(&models.IndustryClass{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}

// ================= IndustrySubclass (Preload("IndustryClass") + limit/offset/total) =================

func seedIndustrySubclassChain(t *testing.T, db *gorm.DB) models.IndustryClass {
	t.Helper()
	group := seedIndustryClassChain(t, db)
	class := models.IndustryClass{IndustryGroupID: group.ID, Name: "Rice Growing", Code: "0111"}
	db.Create(&class)
	return class
}

func TestStore_CreateIndustrySubclass(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	class := seedIndustrySubclassChain(t, db)

	item := &models.IndustrySubclass{IndustryClassID: class.ID, Name: "Rice Milling", Code: "01111"}
	if err := repo.CreateIndustrySubclass(item); err != nil {
		t.Fatalf("CreateIndustrySubclass returned error: %v", err)
	}
	if item.ID == 0 {
		t.Fatalf("expected ID to be populated after create, got 0")
	}
}

func TestStore_GetAllIndustrySubclasses(t *testing.T) {
	t.Run("total reflects all rows regardless of limit", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		class := seedIndustrySubclassChain(t, db)
		for i := 0; i < 5; i++ {
			db.Create(&models.IndustrySubclass{IndustryClassID: class.ID, Name: "Subclass", Code: "X"})
		}

		items, total, err := repo.GetAllIndustrySubclasses(2, 0)
		if err != nil {
			t.Fatalf("GetAllIndustrySubclasses returned error: %v", err)
		}
		if total != 5 {
			t.Fatalf("expected total 5, got %d", total)
		}
		if len(items) != 2 {
			t.Fatalf("expected 2 items with limit=2, got %d", len(items))
		}
	})

	t.Run("offset skips the correct number of rows, ordered by id", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		class := seedIndustrySubclassChain(t, db)
		var seeded []models.IndustrySubclass
		for i := 0; i < 3; i++ {
			s := models.IndustrySubclass{IndustryClassID: class.ID, Name: "Subclass", Code: "X"}
			db.Create(&s)
			seeded = append(seeded, s)
		}

		items, _, err := repo.GetAllIndustrySubclasses(0, 2)
		if err != nil {
			t.Fatalf("GetAllIndustrySubclasses returned error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item after offsetting past 2 of 3, got %d", len(items))
		}
		if items[0].ID != seeded[2].ID {
			t.Fatalf("expected remaining item to be the 3rd seeded row (id %d), got id %d", seeded[2].ID, items[0].ID)
		}
	})

	t.Run("IndustryClass is preloaded on each item", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		class := seedIndustrySubclassChain(t, db)
		db.Create(&models.IndustrySubclass{IndustryClassID: class.ID, Name: "Rice Milling", Code: "01111"})

		items, _, err := repo.GetAllIndustrySubclasses(0, 0)
		if err != nil {
			t.Fatalf("GetAllIndustrySubclasses returned error: %v", err)
		}
		if len(items) != 1 || items[0].IndustryClass == nil || items[0].IndustryClass.Name != "Rice Growing" {
			t.Fatalf("expected preloaded IndustryClass.Name %q, got %+v", "Rice Growing", items)
		}
	})
}

func TestStore_GetIndustrySubclassByID(t *testing.T) {
	t.Run("existing ID returns the row with IndustryClass preloaded", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)
		class := seedIndustrySubclassChain(t, db)
		seeded := models.IndustrySubclass{IndustryClassID: class.ID, Name: "Rice Milling", Code: "01111"}
		db.Create(&seeded)

		got, err := repo.GetIndustrySubclassByID(seeded.ID)
		if err != nil {
			t.Fatalf("GetIndustrySubclassByID returned error: %v", err)
		}
		if got.IndustryClass == nil || got.IndustryClass.Name != "Rice Growing" {
			t.Fatalf("expected preloaded IndustryClass.Name %q, got %+v", "Rice Growing", got.IndustryClass)
		}
	})

	t.Run("non-existent ID returns gorm.ErrRecordNotFound", func(t *testing.T) {
		db := SetupTestDB(t)
		repo := repositories.NewStore(db)

		_, err := repo.GetIndustrySubclassByID(999999)
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("expected gorm.ErrRecordNotFound, got: %v", err)
		}
	})
}

func TestStore_UpdateIndustrySubclass(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	class := seedIndustrySubclassChain(t, db)
	seeded := models.IndustrySubclass{IndustryClassID: class.ID, Name: "Rice Milling", Code: "01111"}
	db.Create(&seeded)

	updated, err := repo.UpdateIndustrySubclass(seeded.ID, map[string]interface{}{"name": "Paddy Milling"})
	if err != nil {
		t.Fatalf("UpdateIndustrySubclass returned error: %v", err)
	}
	if updated.Name != "Paddy Milling" {
		t.Fatalf("expected Name %q, got %q", "Paddy Milling", updated.Name)
	}
}

func TestStore_DeleteIndustrySubclass(t *testing.T) {
	db := SetupTestDB(t)
	repo := repositories.NewStore(db)
	class := seedIndustrySubclassChain(t, db)
	seeded := models.IndustrySubclass{IndustryClassID: class.ID, Name: "Rice Milling", Code: "01111"}
	db.Create(&seeded)

	if err := repo.DeleteIndustrySubclass(seeded.ID); err != nil {
		t.Fatalf("DeleteIndustrySubclass returned error: %v", err)
	}

	var count int64
	db.Model(&models.IndustrySubclass{}).Where("id = ?", seeded.ID).Count(&count)
	if count != 0 {
		t.Fatalf("expected row to be deleted, got count %d", count)
	}
}