package ingestion_test

import (
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/ingestion/mocks"
	"marketlens-go-backend/internal/llm"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"
	"marketlens-go-backend/internal/testutil"
)

// setupDB returns a fresh in-memory DB seeded with the reference data the
// ingestion pipeline reads: formality/gender/etc. options, one industry
// and one occupation hierarchy path, the "Western" province, and the
// active sources the tests ingest from.
func setupDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testutil.SetupTestDB(t)

	rows := []any{
		&models.Formality{ID: 1, FormalityType: "Formal"},
		&models.Formality{ID: 2, FormalityType: "Informal"},
		&models.Gender{ID: 1, GenderType: "Male"},
		&models.Gender{ID: 2, GenderType: "Female"},
		&models.Gender{ID: 3, GenderType: "Any"},
		&models.VocationalEducation{ID: 1, Level: "NVQ Level 4"},
		&models.EmploymentSector{ID: 1, Sector: "Private"},
		&models.EmploymentSector{ID: 2, Sector: "Government"},
		&models.EducationLevel{ID: 1, Level: "O/L"},
		&models.EducationLevel{ID: 2, Level: "A/L"},
		&models.EducationLevel{ID: 3, Level: "Bachelor's Degree"},
		&models.Experience{ID: 1, Name: "Entry Level"},
		&models.Experience{ID: 2, Name: "1-2 Years"},

		&models.IndustrySector{ID: 1, Name: "Information and Communication", Code: "J"},
		&models.IndustryDivision{ID: 62, IndustrySectorID: 1, Name: "Computer Programming Activities", Code: "62"},
		&models.IndustryGroup{ID: 620, IndustryDivisionID: 62, Name: "Computer Programming", Code: "620"},
		&models.IndustryClass{ID: 6201, IndustryGroupID: 620, Name: "Computer Programming Activities", Code: "6201"},
		&models.IndustrySubclass{ID: 62011, IndustryClassID: 6201, Name: "Custom Software Development", Code: "62011"},

		&models.MajorGroup{ID: 2, Name: "Professionals", Code: "2"},
		&models.SubMajorGroup{ID: 25, MajorGroupID: 2, Name: "ICT Professionals", Code: "25"},
		&models.MinorGroup{ID: 251, SubMajorGroupID: 25, Name: "Software and Applications Developers", Code: "251"},
		&models.UnitGroup{ID: 2512, MinorGroupID: 251, Name: "Software Developers", Code: "2512"},
		&models.OccupationGroup{ID: 25121, UnitGroupID: 2512, Name: "Backend Software Developer", Code: "25121"},

		&models.GeoData{Province: "Western"},

		// Crawler runs the tests reference through RawJobInput.CrawlerRunID.
		&models.CrawlerRun{ID: 1, Status: "RUNNING"},
		&models.CrawlerRun{ID: 100, Status: "RUNNING"},
		&models.CrawlerRun{ID: 200, Status: "RUNNING"},

		&models.Source{Name: "Ikman", Type: models.SourceTypeAuto, IsActive: true},
		&models.Source{Name: "TopJobs", Type: models.SourceTypeAuto, IsActive: true},
		&models.Source{Name: "Test", Type: models.SourceTypeAuto, IsActive: true},
	}
	for _, row := range rows {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("failed to seed %T: %v", row, err)
		}
	}
	return db
}

func newServices(db *gorm.DB) *service.Services {
	return service.NewServices(repositories.NewStore(db))
}

func newMetadataBuilder(svcs *service.Services, ttl time.Duration) *ingestion.MetadataBuilder {
	return ingestion.NewMetadataBuilder(
		svcs.Formalities,
		svcs.Genders,
		svcs.VocationalEducations,
		svcs.EmploymentSectors,
		svcs.EducationLevels,
		svcs.Experiences,
		ttl,
	)
}

func newIndustryClassifier(svcs *service.Services, client *llm.Client) *ingestion.IndustryClassifier {
	return ingestion.NewIndustryClassifier(
		svcs.IndustrySectors, svcs.IndustryDivisions, svcs.IndustryGroups,
		svcs.IndustryClasses, svcs.IndustrySubclasses, client,
	)
}

func newOccupationClassifier(svcs *service.Services, client *llm.Client) *ingestion.OccupationClassifier {
	return ingestion.NewOccupationClassifier(
		svcs.MajorGroups, svcs.SubMajorGroups, svcs.MinorGroups,
		svcs.UnitGroups, svcs.OccupationGroups, client,
	)
}

// newIngestionService wires the full pipeline against db, with the LLM
// client talking to a mock server driven by respond.
func newIngestionService(t *testing.T, db *gorm.DB, respond mocks.Responder) *ingestion.IngestionService {
	t.Helper()

	svcs := newServices(db)
	client := mocks.NewMockLLM(t, respond)

	return ingestion.NewIngestionService(
		svcs.JobPosts,
		svcs.Sources,
		client,
		"test-model",
		newIndustryClassifier(svcs, client),
		newOccupationClassifier(svcs, client),
		newMetadataBuilder(svcs, 0),
		0.65,
	)
}

// failWrites makes every create/update GORM runs against table return err
// for the rest of the test, to force a DB failure on one table without a
// real outage. Callbacks are registered per db, and each test has its own.
func failWrites(t *testing.T, db *gorm.DB, table string, err error) {
	t.Helper()

	fail := func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			_ = tx.AddError(err)
		}
	}
	name := "test:fail:" + table
	if e := db.Callback().Create().Before("gorm:create").Register(name, fail); e != nil {
		t.Fatalf("registering create callback: %v", e)
	}
	if e := db.Callback().Update().Before("gorm:update").Register(name, fail); e != nil {
		t.Fatalf("registering update callback: %v", e)
	}
}

// countQueries returns a counter of the SELECTs GORM runs against table.
func countQueries(t *testing.T, db *gorm.DB, table string) *atomic.Int64 {
	t.Helper()

	var n atomic.Int64
	err := db.Callback().Query().Before("gorm:query").Register("test:count:"+table, func(tx *gorm.DB) {
		if tx.Statement.Table == table {
			n.Add(1)
		}
	})
	if err != nil {
		t.Fatalf("registering query callback: %v", err)
	}
	return &n
}

// seedJob inserts an already-saved, active job with the given location and
// MinHash signature, indexed under bucketKeys, as if a previous crawl run
// had ingested it. It returns the new job's id.
func seedJob(t *testing.T, db *gorm.DB, location string, sig []uint32, bucketKeys []string) uint {
	t.Helper()

	job := models.JobPost{
		JobRole:  "Seeded Job",
		Location: location,
		WorkMode: models.WorkModeOnsite,
		MetaData: models.JobMetaData{
			PostedAt:         time.Now(),
			MinhashSignature: ingestion.ToPQArray(sig),
		},
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("failed to seed job: %v", err)
	}
	for i, key := range bucketKeys {
		lsh := models.LshIndex{BucketKey: key, BandNo: i, JobPostID: job.ID}
		if err := db.Create(&lsh).Error; err != nil {
			t.Fatalf("failed to seed lsh row %q: %v", key, err)
		}
	}
	return job.ID
}

func countRows(t *testing.T, db *gorm.DB, model any) int64 {
	t.Helper()

	var n int64
	if err := db.Model(model).Count(&n).Error; err != nil {
		t.Fatalf("counting rows: %v", err)
	}
	return n
}

func itoa(n int) string { return strconv.Itoa(n) }
