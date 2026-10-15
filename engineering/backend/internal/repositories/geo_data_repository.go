package repositories

import (
	"marketlens-go-backend/internal/models"
)

// GeoData returns the generic CRUD repository for GeoData.
func (r *Store) GeoData() *CRUDRepository[models.GeoData] {
	return NewCRUDRepository[models.GeoData](r.db)
}
