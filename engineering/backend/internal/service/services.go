package service

import (
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

type Services struct {
	Analytics   *AnalyticsService
	JobPosts    *JobPostService
	CrawlerRuns *CrawlerRunService

	EducationLevels      *CRUDService[models.EducationLevel]
	Formalities          *CRUDService[models.Formality]
	Genders              *CRUDService[models.Gender]
	EmploymentSectors    *CRUDService[models.EmploymentSector]
	VocationalEducations *CRUDService[models.VocationalEducation]
	Experiences          *CRUDService[models.Experience]
	GeoData              *CRUDService[models.GeoData]
	JobTypes             *CRUDService[models.JobType]

	MajorGroups      *CRUDService[models.MajorGroup]
	SubMajorGroups   *CRUDService[models.SubMajorGroup]
	MinorGroups      *CRUDService[models.MinorGroup]
	UnitGroups       *CRUDService[models.UnitGroup]
	OccupationGroups *CRUDService[models.OccupationGroup]

	IndustrySectors    *CRUDService[models.IndustrySector]
	IndustryDivisions  *CRUDService[models.IndustryDivision]
	IndustryGroups     *CRUDService[models.IndustryGroup]
	IndustryClasses    *CRUDService[models.IndustryClass]
	IndustrySubclasses *CRUDService[models.IndustrySubclass]
}

func NewServices(store *repositories.Store) *Services {
	return &Services{
		Analytics:   NewAnalyticsService(store),
		JobPosts:    NewJobPostService(store),
		CrawlerRuns: NewCrawlerRunService(store),

		EducationLevels:      NewEducationLevelService(store),
		Formalities:          NewFormalityService(store),
		Genders:              NewGenderService(store),
		EmploymentSectors:    NewEmploymentSectorService(store),
		VocationalEducations: NewVocationalEducationService(store),
		Experiences:          NewExperienceService(store),
		GeoData:              NewGeoDataService(store),
		JobTypes:             NewJobTypeService(store),

		MajorGroups:      NewMajorGroupService(store),
		SubMajorGroups:   NewSubMajorGroupService(store),
		MinorGroups:      NewMinorGroupService(store),
		UnitGroups:       NewUnitGroupService(store),
		OccupationGroups: NewOccupationGroupService(store),

		IndustrySectors:    NewIndustrySectorService(store),
		IndustryDivisions:  NewIndustryDivisionService(store),
		IndustryGroups:     NewIndustryGroupService(store),
		IndustryClasses:    NewIndustryClassService(store),
		IndustrySubclasses: NewIndustrySubclassService(store),
	}
}
