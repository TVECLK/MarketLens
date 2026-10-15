package repositories

import (
	"marketlens-go-backend/internal/models"
)

// Formalities returns the generic CRUD repository for Formality.
func (r *Store) Formalities() *CRUDRepository[models.Formality] {
	return NewCRUDRepository[models.Formality](r.db)
}
