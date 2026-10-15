package repositories

import (
	"marketlens-go-backend/internal/models"
)

// EmploymentSectors returns the generic CRUD repository for EmploymentSector.
func (r *Store) EmploymentSectors() *CRUDRepository[models.EmploymentSector] {
	return NewCRUDRepository[models.EmploymentSector](r.db)
}
