package repositories

import (
	"marketlens-go-backend/internal/models"
)

// Genders returns the generic CRUD repository for Gender.
func (r *Store) Genders() *CRUDRepository[models.Gender] {
	return NewCRUDRepository[models.Gender](r.db)
}
