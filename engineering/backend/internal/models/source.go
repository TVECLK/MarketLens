package models

import "time"

// Source model like ikman jobs, rooster
type Source struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Name      string    `json:"name"     gorm:"size:255;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Source) TableName() string { return "source" }
