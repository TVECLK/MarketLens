package models

import "time"

// Employers model for jobs
type Employer struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Name      string    `json:"name"       gorm:"size:255;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Employer) TableName() string { return "employer" }
