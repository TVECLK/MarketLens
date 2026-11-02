package models

import (
	"time"

	"gorm.io/gorm"
)

// Education level model
type EducationLevel struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	Level     string         `json:"level"      gorm:"size:100;not null;unique"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (EducationLevel) TableName() string { return "education_level" }
