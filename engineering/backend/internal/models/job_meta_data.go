package models

import (
	"time"

	"github.com/lib/pq"
)

type JobMetaData struct {
	ID                 uint                  `json:"-"                  gorm:"primaryKey"`
	JobPostID          uint                  `json:"job_post_id"        gorm:"unique;not null"`
	AiVersionID        *uint                 `json:"ai_version_id"`
	AiVersion          *AiVersion            `json:"ai_version"         gorm:"foreignKey:AiVersionID"`
	EducationLevelID   *uint                 `json:"education_level_id"`
	EducationLevel     *EducationLevel       `json:"education_level"    gorm:"foreignKey:EducationLevelID"`
	CrawlerRunID       *uint                 `json:"crawler_run_id"`
	CrawlerRun         *CrawlerRun           `json:"-"                  gorm:"foreignKey:CrawlerRunID"`
	GeoDataID          *uint                 `json:"geo_data_id"`
	GeoData            *GeoData              `json:"geo_data"           gorm:"foreignKey:GeoDataID"`
	SourceID           *uint                 `json:"source_id"`
	Source             *Source               `json:"source"             gorm:"foreignKey:SourceID"`
	ExperienceID       *uint                 `json:"experience_id"`
	Experience         *Experience           `json:"experience"         gorm:"foreignKey:ExperienceID"`
	OccupationGroupID  *uint                 `json:"occupation_group_id"`
	OccupationGroup    *OccupationGroup      `json:"occupation_group"   gorm:"foreignKey:OccupationGroupID"`
	IndustrySubclassID *uint                 `json:"industry_subclass_id"`
	IndustrySubclass   *IndustrySubclass     `json:"industry_subclass"   gorm:"foreignKey:IndustrySubclassID"`
	FormalityID        *uint                 `json:"formality_id"`
	Formality          *Formality            `json:"formality"   gorm:"foreignKey:FormalityID"`
	GenderID           *uint                 `json:"gender_id"`
	Gender             *Gender               `json:"gender"    gorm:"foreignKey:GenderID"`
	VocationalEducationID *uint              `json:"vocational_education_id"`
	VocationalEducation   *VocationalEducation `json:"vocational_education"   gorm:"foreignKey:VocationalEducationID"`
	EmploymentSectorID *uint             `json:"employment_sector_id"`
	EmploymentSector   *EmploymentSector `json:"employment_sector"   gorm:"foreignKey:EmploymentSectorID"`
	PostedAt         *time.Time    `json:"posted_at"`
	EndDate          *time.Time    `json:"end_date"`
	MinhashSignature pq.Int64Array `json:"minhash_signature"  gorm:"type:integer[]"`
	ConfidenceScore  float64       `json:"confidence_score"   gorm:"type:decimal(5,4)"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

func (JobMetaData) TableName() string { return "meta_data" }
