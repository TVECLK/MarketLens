package repositories

import (
	"marketlens-go-backend/internal/models"
)

// VocationalEducations returns the generic CRUD repository for VocationalEducation.
func (r *Store) VocationalEducations() *CRUDRepository[models.VocationalEducation] {
	return NewCRUDRepository[models.VocationalEducation](r.db)
}
