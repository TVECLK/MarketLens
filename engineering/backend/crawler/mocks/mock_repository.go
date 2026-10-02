package mocks

import (
	"sync"

	"marketlens-go-backend/models"
)

// MockRepository is an in-memory stand-in for the real GORM-backed
// repository. It satisfies crawler.Repository structurally — nothing
// needs to import crawler.Repository explicitly here, Go checks this at
// the call site where a *MockRepository is passed in.
//
// Reference lists (formalities, genders, industry/occupation hierarchy)
// come pre-seeded with a small realistic sample; override them via the
// exported fields before use if a test needs different data. Writes
// (SaveOneJob, UpdateDuplicateJob) are captured into SavedJobs /
// SavedLSH / UpdatedDuplicates so tests can assert on exactly what the
// code under test tried to persist.
type MockRepository struct {
	mu sync.Mutex

	// Existing jobs, keyed by id, and their LSH bucket-key index —
	// what GetJobsByBucketKeys searches over. Use SeedJob to populate.
	jobs   map[uint]*models.JobPost
	lsh    map[string][]uint
	nextID uint

	// Captured writes, for assertions after calling code under test.
	SavedJobs         []models.JobPost
	SavedLSH          []models.LshIndex
	UpdatedDuplicates []models.JobMetaData

	// Error injection hooks for tests exercising ProcessBatch's failure
	// handling — nil (the default) means "succeed normally." Set to force
	// SaveOneJob/UpdateDuplicateJob to fail, e.g. wrapping
	// repositories.ErrPermanentSaveFailure to simulate a permanent save
	// error, or a plain error to simulate a transient one.
	SaveOneJobErr      func(job *models.JobPost) error
	UpdateDuplicateErr func(jobPostID uint) error

	// Reference data returned by the metadata/classifier Get* methods.
	// Pre-seeded with a small sample; replace before use for different
	// test data.
	Formalities           []models.Formality
	Genders               []models.Gender
	VocationalEducations  []models.VocationalEducation
	EmploymentSectors     []models.EmploymentSector
	EducationLevels       []models.EducationLevel
	Experiences           []models.Experience
	IndustrySectors       []models.IndustrySector
	IndustryDivisions     []models.IndustryDivision
	IndustryGroups        []models.IndustryGroup
	IndustryClasses       []models.IndustryClass
	IndustrySubclasses    []models.IndustrySubclass
	MajorGroups           []models.MajorGroup
	SubMajorGroups        []models.SubMajorGroup
	MinorGroups            []models.MinorGroup
	UnitGroups             []models.UnitGroup
	OccupationGroups       []models.OccupationGroup

	// FetchCounts lets tests verify caching behavior (e.g.
	// MetadataBuilder's TTL cache) without a real clock dependency.
	FetchCounts map[string]int
}

// NewMockRepository returns a MockRepository pre-seeded with a small,
// realistic sample of reference data — enough for a classifier walk to
// reach a real leaf id, or for MetadataBuilder to produce a non-empty
// instruction.
func NewMockRepository() *MockRepository {
	return &MockRepository{
		jobs:        map[uint]*models.JobPost{},
		lsh:         map[string][]uint{},
		FetchCounts: map[string]int{},

		Formalities: []models.Formality{
			{ID: 1, FormalityType: "Formal"},
			{ID: 2, FormalityType: "Informal"},
		},
		Genders: []models.Gender{
			{ID: 1, GenderType: "Male"},
			{ID: 2, GenderType: "Female"},
			{ID: 3, GenderType: "Any"},
		},
		VocationalEducations: []models.VocationalEducation{
			{ID: 1, Level: "NVQ Level 4"},
		},
		EmploymentSectors: []models.EmploymentSector{
			{ID: 1, Sector: "Private"},
			{ID: 2, Sector: "Government"},
		},
		EducationLevels: []models.EducationLevel{
			{ID: 1, Level: "O/L"},
			{ID: 2, Level: "A/L"},
			{ID: 3, Level: "Bachelor's Degree"},
		},
		Experiences: []models.Experience{
			{ID: 1, Name: "Entry Level"},
			{ID: 2, Name: "1-2 Years"},
		},

		IndustrySectors:    []models.IndustrySector{{ID: 1, Name: "Information and Communication", Code: "J"}},
		IndustryDivisions:  []models.IndustryDivision{{ID: 62, IndustrySectorID: 1, Name: "Computer Programming Activities", Code: "62"}},
		IndustryGroups:     []models.IndustryGroup{{ID: 620, IndustryDivisionID: 62, Name: "Computer Programming", Code: "620"}},
		IndustryClasses:    []models.IndustryClass{{ID: 6201, IndustryGroupID: 620, Name: "Computer Programming Activities", Code: "6201"}},
		IndustrySubclasses: []models.IndustrySubclass{{ID: 62011, IndustryClassID: 6201, Name: "Custom Software Development", Code: "62011"}},

		MajorGroups:      []models.MajorGroup{{ID: 2, Name: "Professionals", Code: "2"}},
		SubMajorGroups:   []models.SubMajorGroup{{ID: 25, MajorGroupID: 2, Name: "ICT Professionals", Code: "25"}},
		MinorGroups:      []models.MinorGroup{{ID: 251, SubMajorGroupID: 25, Name: "Software and Applications Developers", Code: "251"}},
		UnitGroups:       []models.UnitGroup{{ID: 2512, MinorGroupID: 251, Name: "Software Developers", Code: "2512"}},
		OccupationGroups: []models.OccupationGroup{{ID: 25121, UnitGroupID: 2512, Name: "Backend Software Developer", Code: "25121"}},
	}
}

// SeedJob directly inserts a job into the store, as if it had been
// saved by a previous crawl run — for setting up "this job already
// exists" scenarios without going through BatchSaveNewJobs.
func (r *MockRepository) SeedJob(id uint, location string, minhashSignature []uint32, lshBucketKeys []string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	sig := make([]int64, len(minhashSignature))
	for i, v := range minhashSignature {
		sig[i] = int64(int32(v))
	}

	r.jobs[id] = &models.JobPost{
		ID:       id,
		Location: location,
		MetaData: models.JobMetaData{
			JobPostID:        id,
			MinhashSignature: sig,
		},
	}
	if id >= r.nextID {
		r.nextID = id
	}
	for _, key := range lshBucketKeys {
		r.lsh[key] = append(r.lsh[key], id)
	}
}

// --- DuplicateLookupRepository / batch writes ---

func (r *MockRepository) GetJobsByBucketKeys(bucketKeys []string) ([]models.JobPost, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	seen := map[uint]bool{}
	var out []models.JobPost
	for _, key := range bucketKeys {
		for _, id := range r.lsh[key] {
			if seen[id] {
				continue
			}
			seen[id] = true
			if job, ok := r.jobs[id]; ok {
				out = append(out, *job)
			}
		}
	}
	return out, nil
}

func (r *MockRepository) SaveOneJob(job *models.JobPost, lshIndexRecords []models.LshIndex) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.SaveOneJobErr != nil {
		if err := r.SaveOneJobErr(job); err != nil {
			return err
		}
	}

	r.nextID++
	job.ID = r.nextID
	jobCopy := *job
	r.jobs[job.ID] = &jobCopy

	for i := range lshIndexRecords {
		lshIndexRecords[i].JobPostID = job.ID
		r.lsh[lshIndexRecords[i].BucketKey] = append(r.lsh[lshIndexRecords[i].BucketKey], job.ID)
	}

	r.SavedJobs = append(r.SavedJobs, jobCopy)
	r.SavedLSH = append(r.SavedLSH, lshIndexRecords...)
	return nil
}

func (r *MockRepository) UpdateDuplicateJob(jobPostID uint, crawlerRunID uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.UpdateDuplicateErr != nil {
		if err := r.UpdateDuplicateErr(jobPostID); err != nil {
			return err
		}
	}

	if job, ok := r.jobs[jobPostID]; ok {
		job.MetaData.CrawlerRunID = &crawlerRunID
	}
	r.UpdatedDuplicates = append(r.UpdatedDuplicates, models.JobMetaData{
		JobPostID:    jobPostID,
		CrawlerRunID: &crawlerRunID,
	})
	return nil
}

// --- IndustryRepository ---

func (r *MockRepository) GetAllIndustrySectors() ([]models.IndustrySector, error) { return r.IndustrySectors, nil }

func (r *MockRepository) GetIndustryDivisionsByIndustrySector(uint) ([]models.IndustryDivision, error) {
	return r.IndustryDivisions, nil
}

func (r *MockRepository) GetIndustryGroupsByIndustryDivision(uint) ([]models.IndustryGroup, error) {
	return r.IndustryGroups, nil
}

func (r *MockRepository) GetIndustryClassesByIndustryGroup(uint) ([]models.IndustryClass, error) {
	return r.IndustryClasses, nil
}

func (r *MockRepository) GetIndustrySubclassesByIndustryClass(uint) ([]models.IndustrySubclass, error) {
	return r.IndustrySubclasses, nil
}

// --- OccupationRepository ---

func (r *MockRepository) GetAllMajorGroups() ([]models.MajorGroup, error) { return r.MajorGroups, nil }

func (r *MockRepository) GetSubMajorGroupsByMajorGroup(uint) ([]models.SubMajorGroup, error) {
	return r.SubMajorGroups, nil
}

func (r *MockRepository) GetMinorGroupsBySubMajorGroup(uint) ([]models.MinorGroup, error) {
	return r.MinorGroups, nil
}

func (r *MockRepository) GetUnitGroupsByMinorGroup(uint) ([]models.UnitGroup, error) {
	return r.UnitGroups, nil
}

func (r *MockRepository) GetOccupationGroupsByUnitGroup(uint) ([]models.OccupationGroup, error) {
	return r.OccupationGroups, nil
}

// --- MetadataRepository ---

func (r *MockRepository) GetAllFormalities() ([]models.Formality, error) {
	r.mu.Lock()
	r.FetchCounts["formalities"]++
	r.mu.Unlock()
	return r.Formalities, nil
}

func (r *MockRepository) GetAllGenders() ([]models.Gender, error) { return r.Genders, nil }

func (r *MockRepository) GetAllVocationalEducations() ([]models.VocationalEducation, error) {
	return r.VocationalEducations, nil
}

func (r *MockRepository) GetAllEmploymentSectors() ([]models.EmploymentSector, error) {
	return r.EmploymentSectors, nil
}

func (r *MockRepository) GetAllEducationLevels() ([]models.EducationLevel, error) {
	return r.EducationLevels, nil
}

func (r *MockRepository) GetAllExperiences() ([]models.Experience, error) { return r.Experiences, nil }