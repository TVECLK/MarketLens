package models

import "time"

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
