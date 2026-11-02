package repositories

import (
	"marketlens-go-backend/internal/models"
)

// EducationLevels returns the generic CRUD repository for EducationLevel.
func (r *Store) EducationLevels() *CRUDRepository[models.EducationLevel] {
	return NewCRUDRepository[models.EducationLevel](r.db)
}
