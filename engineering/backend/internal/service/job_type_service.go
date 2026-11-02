package service

import (
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

func NewJobTypeService(store *repositories.Store) *CRUDService[models.JobType] {
	return NewCRUDServiceFromRepository(
		store.JobTypes(),
		requireNonEmptyField("type", func(j *models.JobType) string { return j.Type }),
	)
}
