package models

import (
	"time"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Lookup / reference tables 

// Employers model for jobs
type Employer struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Name      string    `json:"name"       gorm:"size:255;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Employer) TableName() string { return "employer" }

// Type of the job like full time, part time etc
type JobType struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Type      string    `json:"type"       gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (JobType) TableName() string       { return "job_type" }

// Skill model for each job
type Skill struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Skill     string    `json:"skill"      gorm:"size:100;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Skill) TableName() string         { return "skills" }

// AI version model
type AiVersion struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Version   string    `json:"version"    gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AiVersion) TableName() string { return "ai_version" }

// Education level model
type EducationLevel struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Level     string    `json:"level"      gorm:"size:100;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (EducationLevel) TableName() string { return "education_level" }

// Source model like ikman jobs, rooster
type Source struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Name    string    `json:"name"     gorm:"size:255;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Source) TableName() string { return "source" }

// Experience level model
type Experience struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Name      string    `json:"name"       gorm:"size:100;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Experience) TableName() string { return "experience" }

// Geo data model (provinces)
type GeoData struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Longitude float64   `json:"lng"        gorm:"type:decimal(9,6)"`
	Latitude  float64   `json:"lat"        gorm:"type:decimal(9,6)"`
	Province  string    `json:"province"   gorm:"size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (GeoData) TableName() string { return "geo_data" }

// Formal / Informal model
type Formality struct {
	ID				uint		`json:"id"					gorm:"primaryKey"` 
	FormalityType	string		`json:"formality_type"		gorm:"size:50;not null"`
	CreatedAt		time.Time	`json:"created_at"`
	UpdatedAt		time.Time	`json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Formality) TableName() string { return "formality" } 

// Gender model
type Gender struct {
	ID				uint		`json:"id"				gorm:"primaryKey"`
	GenderType		string		`json:"gender_type"		gorm:"size:50;not null"` 
	CreatedAt		time.Time	`json:"created_at"`
	UpdatedAt		time.Time	`json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Gender) TableName() string { return "gender" }

// Vocational education model
type VocationalEducation struct {
	ID				uint		`json:"id"				gorm:"primaryKey"`
	Level			string		`json:"level"			gorm:"size:50;not null"`
	CreatedAt		time.Time	`json:"created_at"`
	UpdatedAt		time.Time	`json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (VocationalEducation) TableName() string { return "vocational_education" }

// Employment sector model like private, goverment
type EmploymentSector struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Sector    string    `json:"sector"     gorm:"size:100;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (EmploymentSector) TableName() string { return "employment_sector" }

// Occupation classification tables
type MajorGroup struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Name      string    `json:"name"       gorm:"size:255;not null"`
	Code      string    `json:"code"       gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (MajorGroup) TableName() string { return "major_group" }

type SubMajorGroup struct {
	ID           uint        `json:"id"             gorm:"primaryKey"`
	MajorGroupID uint        `json:"major_group_id" gorm:"not null"`
	MajorGroup   *MajorGroup `json:"major_group"    gorm:"foreignKey:MajorGroupID;constraint:OnDelete:CASCADE;"`
	Name         string      `json:"name"           gorm:"size:255;not null"`
	Code         string      `json:"code"           gorm:"size:50;not null"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SubMajorGroup) TableName() string { return "sub_major_group" }

type MinorGroup struct {
	ID              uint           `json:"id"                 gorm:"primaryKey"`
	SubMajorGroupID uint           `json:"sub_major_group_id" gorm:"not null"`
	SubMajorGroup   *SubMajorGroup `json:"sub_major_group"    gorm:"foreignKey:SubMajorGroupID;constraint:OnDelete:CASCADE;"`
	Name            string         `json:"name"               gorm:"size:255;not null"`
	Code            string         `json:"code"               gorm:"size:50;not null"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (MinorGroup) TableName() string { return "minor_group" }

type UnitGroup struct {
	ID           uint        `json:"id"             gorm:"primaryKey"`
	MinorGroupID uint        `json:"minor_group_id" gorm:"not null"`
	MinorGroup   *MinorGroup `json:"minor_group"    gorm:"foreignKey:MinorGroupID;constraint:OnDelete:CASCADE;"`
	Name         string      `json:"name"           gorm:"size:255;not null"`
	Code         string      `json:"code"           gorm:"size:50;not null"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (UnitGroup) TableName() string { return "unit_group" }

type OccupationGroup struct {
	ID          uint       `json:"id"            gorm:"primaryKey"`
	UnitGroupID uint       `json:"unit_group_id" gorm:"not null"`
	UnitGroup   *UnitGroup `json:"unit_group"    gorm:"foreignKey:UnitGroupID;constraint:OnDelete:CASCADE;"`
	Name        string     `json:"name"          gorm:"size:255;not null"`
	Code        string     `json:"code"          gorm:"size:50;not null"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (OccupationGroup) TableName() string { return "occupation_group" }

// Industry classification tables
type IndustrySector struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Name      string    `json:"name"       gorm:"size:255;not null"`
	Code      string    `json:"code"       gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (IndustrySector) TableName() string { return "industry_sector" }

type IndustryDivision struct {
	ID               uint            `json:"id"                 gorm:"primaryKey"`
	IndustrySectorID uint            `json:"industry_sector_id" gorm:"not null"`
	IndustrySector   *IndustrySector `json:"industry_sector"    gorm:"foreignKey:IndustrySectorID;constraint:OnDelete:CASCADE;"`
	Name             string          `json:"name"               gorm:"size:255;not null"`
	Code             string          `json:"code"               gorm:"size:50;not null"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (IndustryDivision) TableName() string { return "industry_division" }

type IndustryGroup struct {
	ID                 uint              `json:"id"                   gorm:"primaryKey"`
	IndustryDivisionID uint              `json:"industry_division_id" gorm:"not null"`
	IndustryDivision   *IndustryDivision `json:"industry_division"    gorm:"foreignKey:IndustryDivisionID;constraint:OnDelete:CASCADE;"`
	Name               string            `json:"name"                 gorm:"size:255;not null"`
	Code               string            `json:"code"                 gorm:"size:50;not null"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (IndustryGroup) TableName() string { return "industry_group" }

type IndustryClass struct {
	ID              uint           `json:"id"                gorm:"primaryKey"`
	IndustryGroupID uint           `json:"industry_group_id" gorm:"not null"`
	IndustryGroup   *IndustryGroup `json:"industry_group"    gorm:"foreignKey:IndustryGroupID;constraint:OnDelete:CASCADE;"`
	Name            string         `json:"name"              gorm:"size:255;not null"`
	Code            string         `json:"code"              gorm:"size:50;not null"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (IndustryClass) TableName() string { return "industry_class" }

type IndustrySubclass struct {
	ID              uint           `json:"id"                gorm:"primaryKey"`
	IndustryClassID uint           `json:"industry_class_id" gorm:"not null"`
	IndustryClass   *IndustryClass `json:"industry_class"    gorm:"foreignKey:IndustryClassID;constraint:OnDelete:CASCADE;"`
	Name            string         `json:"name"              gorm:"size:255;not null"`
	Code            string         `json:"code"              gorm:"size:50;not null"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
 
func (IndustrySubclass) TableName() string { return "industry_subclass" }

// Crawler run

type CrawlerRun struct {
	ID         uint       `json:"id"          gorm:"primaryKey"`
	StartedAt  *time.Time `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Status     string     `json:"status"      gorm:"size:50"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (CrawlerRun) TableName() string { return "crawler_runs" }

type WorkMode string

const (
	WorkModeRemote WorkMode = "remote"
	WorkModeOnsite WorkMode = "onsite"
	WorkModeHybrid WorkMode = "hybrid"
)

// Core job tables

type JobPost struct {
	ID             uint       `json:"id"              gorm:"primaryKey"`
	EmployerID     *uint      `json:"employer_id"`
	Employer       *Employer  `json:"employer"        gorm:"foreignKey:EmployerID"`
	JobTypeID      *uint      `json:"job_type_id"`
	JobType        *JobType   `json:"job_type"        gorm:"foreignKey:JobTypeID"`
	JobRole        string     `json:"job_role"        gorm:"size:255;not null"`
	WorkMode       WorkMode    `json:"work_mode"       gorm:"type:work_mode_enum;not null;default:onsite"`
	JobDescription string     `json:"job_description" gorm:"type:text"`
	Location       string     `json:"location"        gorm:"size:255"`
	NoOfVacancies  int        `json:"no_of_vacancies" gorm:"not null;default:1"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	MetaData       JobMetaData `json:"meta_data" gorm:"foreignKey:JobPostID;constraint:OnDelete:CASCADE;"`
	Skills         []Skill     `json:"skills"    gorm:"many2many:job_post_skills;constraint:OnDelete:CASCADE;"`
}

func (JobPost) TableName() string { return "job_post" }

type JobMetaData struct {
	ID               uint            `json:"-"                  gorm:"primaryKey"`
	JobPostID        uint            `json:"job_post_id"        gorm:"unique;not null"`
	AiVersionID      *uint           `json:"ai_version_id"`
	AiVersion        *AiVersion      `json:"ai_version"         gorm:"foreignKey:AiVersionID"`
	EducationLevelID *uint           `json:"education_level_id"`
	EducationLevel   *EducationLevel `json:"education_level"    gorm:"foreignKey:EducationLevelID"`
	CrawlerRunID     *uint           `json:"crawler_run_id"`
	CrawlerRun       *CrawlerRun     `json:"-"                  gorm:"foreignKey:CrawlerRunID"`
	GeoDataID        *uint           `json:"geo_data_id"`
	GeoData          *GeoData        `json:"geo_data"           gorm:"foreignKey:GeoDataID"`
	// IndustryID       *uint           `json:"industry_id"`
	// Industry         *Industry       `json:"industry"           gorm:"foreignKey:IndustryID"`
	// OccupationID     *uint           `json:"occupation_id"`
	// Occupation       *Occupation     `json:"occupation"         gorm:"foreignKey:OccupationID"`
	SourceID         *uint           `json:"source_id"`
	Source           *Source         `json:"source"             gorm:"foreignKey:SourceID"`
	ExperienceID     *uint           `json:"experience_id"`
	Experience       *Experience     `json:"experience"         gorm:"foreignKey:ExperienceID"`
	OccupationGroupID *uint            `json:"occupation_group_id"`
	OccupationGroup   *OccupationGroup `json:"occupation_group"   gorm:"foreignKey:OccupationGroupID"`
	IndustrySubclassID *uint             `json:"industry_subclass_id"`
	IndustrySubclass   *IndustrySubclass `json:"industry_subclass"   gorm:"foreignKey:IndustrySubclassID"`
	FormalityID *uint      `json:"formality_id"`
	Formality   *Formality `json:"formality"   gorm:"foreignKey:FormalityID"`
	GenderID *uint   `json:"gender_id"`
	Gender   *Gender `json:"gender"    gorm:"foreignKey:GenderID"`
	VocationalEducationID *uint                 `json:"vocational_education_id"`
	VocationalEducation   *VocationalEducation  `json:"vocational_education"   gorm:"foreignKey:VocationalEducationID"`
	EmploymentSectorID *uint             `json:"employment_sector_id"`
	EmploymentSector   *EmploymentSector `json:"employment_sector"   gorm:"foreignKey:EmploymentSectorID"`
	PostedAt         *time.Time      `json:"posted_at"`
	EndDate          *time.Time      `json:"end_date"`
	MinhashSignature pq.Int64Array   `json:"minhash_signature"  gorm:"type:integer[]"`
	ConfidenceScore  float64         `json:"confidence_score"   gorm:"type:decimal(5,4)"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (JobMetaData) TableName() string { return "meta_data" }

// LSH dedup index

type LshIndex struct {
	BucketKey string   `json:"bucket_key" gorm:"primaryKey;size:64;not null"`
	BandNo    int      `json:"band_no"    gorm:"not null"`
	JobPostID uint     `json:"job_post_id" gorm:"primaryKey;not null"`
	JobPost   *JobPost `json:"-"          gorm:"foreignKey:JobPostID;constraint:OnDelete:CASCADE;"`
}

func (LshIndex) TableName() string { return "lsh_index" }

// API payloads

type LevelChildJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	OpenJobCount int64  `json:"open_job_count"`
}

type TotalVacancyCount struct {
	FromDate       string `json:"from_date"`
	ToDate         string `json:"to_date"`
	TotalVacancies int64  `json:"total_vacancies"`
}

type VacancyTrendPoint struct {
	Label        string `json:"label"`
	OpenJobCount int64  `json:"open_job_count"`
}

type BatchSavePayload struct {
	NewJobs    []JobPost  `json:"new_jobs"    binding:"required"`
	LshIndexes []LshIndex `json:"lsh_indexes" binding:"required"`
}

type BatchUpdatePayload struct {
	Duplicates []JobMetaData `json:"duplicates" binding:"required"`
}

type ReconciliationPayload struct {
	CrawlerRunID uint `json:"crawler_run_id" binding:"required"`
}

type SkillDemand struct {
	ID           uint   `json:"id"`
	Skill        string `json:"skill"`
	OpenJobCount int64  `json:"open_job_count"`
}

type EmployerDemand struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type OccupationJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type IndustryJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type ExperienceJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type EducationLevelJobCount struct {
	ID           uint   `json:"id"`
	Level        string `json:"level"`
	OpenJobCount int64  `json:"open_job_count"`
}

type RemoteOnSiteHybridCount struct {
	RemoteCount int64 `json:"remote_count"`
	OnSiteCount int64 `json:"on_site_count"`
	HybridCount int64 `json:"hybrid_count"`
}

type JobTypeJobCount struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"`
	OpenJobCount int64  `json:"open_job_count"`
}

type TopJobRole struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type ProvinceJobCount struct {
	ID           uint    `json:"id"`
	Province     string  `json:"province"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
	OpenJobCount int64   `json:"open_job_count"`
}

type FormalityJobCount struct {
	ID             uint   `json:"id"`
	FormalityType  string `json:"formality_type"`
	OpenJobCount   int64  `json:"open_job_count"`
}

type EmploymentSectorJobCount struct {
	ID           uint   `json:"id"`
	Sector       string `json:"sector"`
	OpenJobCount int64  `json:"open_job_count"`
}

type GenderJobCount struct {
	ID           uint   `json:"id"`
	GenderType   string `json:"gender_type"`
	OpenJobCount int64  `json:"open_job_count"`
}

type VocationalEducationJobCount struct {
	ID           uint   `json:"id"`
	Level        string `json:"level"`
	OpenJobCount int64  `json:"open_job_count"`
}