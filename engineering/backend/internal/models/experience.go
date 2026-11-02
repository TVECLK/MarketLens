package models

import (
	"time"

	"gorm.io/gorm"
)

// Experience level model
type Experience struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	Name      string         `json:"name"       gorm:"size:100;not null;unique"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Experience) TableName() string { return "experience" }
