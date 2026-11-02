package models

import "time"

// Skill model for each job
type Skill struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Skill     string    `json:"skill"      gorm:"size:100;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Skill) TableName() string { return "skills" }
