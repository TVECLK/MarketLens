package models

import (
	"time"

	"gorm.io/gorm"
)

// Industry classification tables

type IndustrySector struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	Name      string         `json:"name"       gorm:"size:255;not null"`
	Code      string         `json:"code"       gorm:"size:50;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (IndustrySector) TableName() string { return "industry_sector" }

type IndustryDivision struct {
	ID               uint            `json:"id"                 gorm:"primaryKey"`
	IndustrySectorID uint            `json:"industry_sector_id" gorm:"not null"`
	IndustrySector   *IndustrySector `json:"industry_sector"    gorm:"foreignKey:IndustrySectorID;constraint:OnDelete:CASCADE;"`
	Name             string          `json:"name"               gorm:"size:255;not null"`
	Code             string          `json:"code"               gorm:"size:50;not null"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `json:"-" gorm:"index"`
}

func (IndustryDivision) TableName() string { return "industry_division" }

type IndustryGroup struct {
	ID                 uint              `json:"id"                   gorm:"primaryKey"`
	IndustryDivisionID uint              `json:"industry_division_id" gorm:"not null"`
	IndustryDivision   *IndustryDivision `json:"industry_division"    gorm:"foreignKey:IndustryDivisionID;constraint:OnDelete:CASCADE;"`
	Name               string            `json:"name"                 gorm:"size:255;not null"`
	Code               string            `json:"code"                 gorm:"size:50;not null"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
	DeletedAt          gorm.DeletedAt    `json:"-" gorm:"index"`
}

func (IndustryGroup) TableName() string { return "industry_group" }

type IndustryClass struct {
	ID              uint           `json:"id"                gorm:"primaryKey"`
	IndustryGroupID uint           `json:"industry_group_id" gorm:"not null"`
	IndustryGroup   *IndustryGroup `json:"industry_group"    gorm:"foreignKey:IndustryGroupID;constraint:OnDelete:CASCADE;"`
	Name            string         `json:"name"              gorm:"size:255;not null"`
	Code            string         `json:"code"              gorm:"size:50;not null"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (IndustryClass) TableName() string { return "industry_class" }

type IndustrySubclass struct {
	ID              uint           `json:"id"                gorm:"primaryKey"`
	IndustryClassID uint           `json:"industry_class_id" gorm:"not null"`
	IndustryClass   *IndustryClass `json:"industry_class"    gorm:"foreignKey:IndustryClassID;constraint:OnDelete:CASCADE;"`
	Name            string         `json:"name"              gorm:"size:255;not null"`
	Code            string         `json:"code"              gorm:"size:50;not null"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (IndustrySubclass) TableName() string { return "industry_subclass" }
