package models

import "time"

// AI version model
type AiVersion struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Version   string    `json:"version"    gorm:"size:50;not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AiVersion) TableName() string { return "ai_version" }
