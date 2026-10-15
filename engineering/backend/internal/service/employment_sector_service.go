package service

import (
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

func NewEmploymentSectorService(store *repositories.Store) *CRUDService[models.EmploymentSector] {
	return NewCRUDServiceFromRepository(
		store.EmploymentSectors(),
		requireNonEmptyField("sector", func(e *models.EmploymentSector) string { return e.Sector }),
	)
}
