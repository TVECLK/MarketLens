package repositories

import (
	"marketlens-go-backend/internal/models"
)

// JobTypes returns the generic CRUD repository for JobType.
func (r *Store) JobTypes() *CRUDRepository[models.JobType] {
	return NewCRUDRepository[models.JobType](r.db)
}
