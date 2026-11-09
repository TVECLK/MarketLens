package models

import "time"

type SourceType string

const (
	SourceTypeAuto   SourceType = "auto"   // collected by a crawler
	SourceTypeManual SourceType = "manual" // entered by hand, e.g. newspaper ads
)

// Source model like ikman jobs, rooster
type Source struct {
	ID        uint       `json:"id"         gorm:"primaryKey"`
	Name      string     `json:"name"       gorm:"size:255;not null;unique"`
	Type      SourceType `json:"type"       gorm:"type:source_type_enum;not null"`
	IsActive  bool       `json:"is_active"  gorm:"not null;default:true"`
	UpdatedBy *string    `json:"updated_by" gorm:"size:255"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (Source) TableName() string { return "source" }
