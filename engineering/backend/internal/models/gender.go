package models

import (
	"time"

	"gorm.io/gorm"
)

// Gender model
type Gender struct {
	ID         uint           `json:"id"				gorm:"primaryKey"`
	GenderType string         `json:"gender_type"		gorm:"size:50;not null;unique"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Gender) TableName() string { return "gender" }
