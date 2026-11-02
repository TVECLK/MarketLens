package repositories

import (
	"marketlens-go-backend/internal/models"
)

// Experiences returns the generic CRUD repository for Experience.
func (r *Store) Experiences() *CRUDRepository[models.Experience] {
	return NewCRUDRepository[models.Experience](r.db)
}
