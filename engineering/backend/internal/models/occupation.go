package models

import (
	"time"

	"gorm.io/gorm"
)

// Occupation classification tables

type MajorGroup struct {
	ID        uint           `json:"id"         gorm:"primaryKey"`
	Name      string         `json:"name"       gorm:"size:255;not null"`
	Code      string         `json:"code"       gorm:"size:50;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (MajorGroup) TableName() string { return "major_group" }

type SubMajorGroup struct {
	ID           uint           `json:"id"             gorm:"primaryKey"`
	MajorGroupID uint           `json:"major_group_id" gorm:"not null"`
	MajorGroup   *MajorGroup    `json:"major_group"    gorm:"foreignKey:MajorGroupID;constraint:OnDelete:CASCADE;"`
	Name         string         `json:"name"           gorm:"size:255;not null"`
	Code         string         `json:"code"           gorm:"size:50;not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SubMajorGroup) TableName() string { return "sub_major_group" }

type MinorGroup struct {
	ID              uint           `json:"id"                 gorm:"primaryKey"`
	SubMajorGroupID uint           `json:"sub_major_group_id" gorm:"not null"`
	SubMajorGroup   *SubMajorGroup `json:"sub_major_group"    gorm:"foreignKey:SubMajorGroupID;constraint:OnDelete:CASCADE;"`
	Name            string         `json:"name"               gorm:"size:255;not null"`
	Code            string         `json:"code"               gorm:"size:50;not null"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

func (MinorGroup) TableName() string { return "minor_group" }

type UnitGroup struct {
	ID           uint           `json:"id"             gorm:"primaryKey"`
	MinorGroupID uint           `json:"minor_group_id" gorm:"not null"`
	MinorGroup   *MinorGroup    `json:"minor_group"    gorm:"foreignKey:MinorGroupID;constraint:OnDelete:CASCADE;"`
	Name         string         `json:"name"           gorm:"size:255;not null"`
	Code         string         `json:"code"           gorm:"size:50;not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (UnitGroup) TableName() string { return "unit_group" }

type OccupationGroup struct {
	ID          uint           `json:"id"            gorm:"primaryKey"`
	UnitGroupID uint           `json:"unit_group_id" gorm:"not null"`
	UnitGroup   *UnitGroup     `json:"unit_group"    gorm:"foreignKey:UnitGroupID;constraint:OnDelete:CASCADE;"`
	Name        string         `json:"name"          gorm:"size:255;not null"`
	Code        string         `json:"code"          gorm:"size:50;not null"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (OccupationGroup) TableName() string { return "occupation_group" }
