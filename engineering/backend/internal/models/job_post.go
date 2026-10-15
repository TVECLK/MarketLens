package models

import "time"

type WorkMode string

const (
	WorkModeRemote WorkMode = "remote"
	WorkModeOnsite WorkMode = "onsite"
	WorkModeHybrid WorkMode = "hybrid"
)

// Core job tables

type JobPost struct {
	ID             uint        `json:"id"              gorm:"primaryKey"`
	EmployerID     *uint       `json:"employer_id"`
	Employer       *Employer   `json:"employer"        gorm:"foreignKey:EmployerID"`
	JobTypeID      *uint       `json:"job_type_id"`
	JobType        *JobType    `json:"job_type"        gorm:"foreignKey:JobTypeID"`
	JobRole        string      `json:"job_role"        gorm:"size:255;not null"`
	WorkMode       WorkMode    `json:"work_mode"       gorm:"type:work_mode_enum;not null;default:onsite"`
	JobDescription string      `json:"job_description" gorm:"type:text"`
	Location       string      `json:"location"        gorm:"size:255"`
	NoOfVacancies  int         `json:"no_of_vacancies" gorm:"not null;default:1"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	MetaData       JobMetaData `json:"meta_data" gorm:"foreignKey:JobPostID;constraint:OnDelete:CASCADE;"`
	Skills         []Skill     `json:"skills"    gorm:"many2many:job_post_skills;constraint:OnDelete:CASCADE;"`
}

func (JobPost) TableName() string { return "job_post" }
