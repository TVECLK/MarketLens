package models

import (
	"time"

	"gorm.io/gorm"
)

// Employment sector model like private, goverment
type EmploymentSector struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	Sector    string         `json:"sector"     gorm:"size:100;not null;unique"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (EmploymentSector) TableName() string { return "employment_sector" }
