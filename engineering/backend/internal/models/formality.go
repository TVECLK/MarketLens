package models

import (
	"time"

	"gorm.io/gorm"
)

// Formal / Informal model
type Formality struct {
	ID            uint           `json:"id"					gorm:"primaryKey"`
	FormalityType string         `json:"formality_type"		gorm:"size:50;not null;unique"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Formality) TableName() string { return "formality" }
