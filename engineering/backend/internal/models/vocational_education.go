package models

import (
	"time"

	"gorm.io/gorm"
)

// Vocational education model
type VocationalEducation struct {
	ID        uint           `json:"id"				gorm:"primaryKey"`
	Level     string         `json:"level"			gorm:"size:50;not null;unique"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (VocationalEducation) TableName() string { return "vocational_education" }
