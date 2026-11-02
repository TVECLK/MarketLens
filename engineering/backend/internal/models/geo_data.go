package models

import "time"

// Geo data model (provinces)
type GeoData struct {
	ID        uint      `json:"id"         gorm:"primaryKey"`
	Longitude float64   `json:"lng"        gorm:"type:decimal(9,6)"`
	Latitude  float64   `json:"lat"        gorm:"type:decimal(9,6)"`
	Province  string    `json:"province"   gorm:"size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (GeoData) TableName() string { return "geo_data" }
