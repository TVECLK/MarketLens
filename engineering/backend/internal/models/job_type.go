package models

import "time"

// Type of the job like full time, part time etc
type JobType struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Type      string    `json:"type"       gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (JobType) TableName() string { return "job_type" }
