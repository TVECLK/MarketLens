package service

import (
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

func NewExperienceService(store *repositories.Store) *CRUDService[models.Experience] {
	return NewCRUDServiceFromRepository(
		store.Experiences(),
		requireNonEmptyField("name", func(e *models.Experience) string { return e.Name }),
	)
}
