package crawler

import "marketlens-go-backend/models"

// DuplicateLookupRepository is the read access CheckDuplicate needs to
// find candidate jobs sharing an LSH bucket key.
type DuplicateLookupRepository interface {
	GetJobsByBucketKeys(bucketKeys []string) ([]models.JobPost, error)
}

// IndustryRepository is the read access IndustryClassifier needs to walk
// Industry Sector -> Division -> Group -> Class -> Subclass.
type IndustryRepository interface {
	GetAllIndustrySectors() ([]models.IndustrySector, error)
	GetIndustryDivisionsByIndustrySector(industrySectorID uint) ([]models.IndustryDivision, error)
	GetIndustryGroupsByIndustryDivision(industryDivisionID uint) ([]models.IndustryGroup, error)
	GetIndustryClassesByIndustryGroup(industryGroupID uint) ([]models.IndustryClass, error)
	GetIndustrySubclassesByIndustryClass(industryClassID uint) ([]models.IndustrySubclass, error)
}

// OccupationRepository is the read access OccupationClassifier needs to
// walk Major Group -> Sub Major Group -> Minor Group -> Unit Group ->
// Occupation Group.
type OccupationRepository interface {
	GetAllMajorGroups() ([]models.MajorGroup, error)
	GetSubMajorGroupsByMajorGroup(majorGroupID uint) ([]models.SubMajorGroup, error)
	GetMinorGroupsBySubMajorGroup(subMajorGroupID uint) ([]models.MinorGroup, error)
	GetUnitGroupsByMinorGroup(minorGroupID uint) ([]models.UnitGroup, error)
	GetOccupationGroupsByUnitGroup(unitGroupID uint) ([]models.OccupationGroup, error)
}

// MetadataRepository is the read access MetadataBuilder needs to fetch
// the reference option lists (formality, gender, etc.) used in the
// extraction prompt.
type MetadataRepository interface {
	GetAllFormalities() ([]models.Formality, error)
	GetAllGenders() ([]models.Gender, error)
	GetAllVocationalEducations() ([]models.VocationalEducation, error)
	GetAllEmploymentSectors() ([]models.EmploymentSector, error)
	GetAllEducationLevels() ([]models.EducationLevel, error)
	GetAllExperiences() ([]models.Experience, error)
}

// Repository is the full set of data access IngestionService needs: all
// of the above, plus the per-job write operations. This is the only
// interface main.go needs to know about — everything else is internal
// wiring.
type Repository interface {
	DuplicateLookupRepository
	IndustryRepository
	OccupationRepository
	MetadataRepository

	SaveOneJob(job *models.JobPost, lshIndexRecords []models.LshIndex) error
	UpdateDuplicateJob(jobPostID uint, crawlerRunID uint) error
}