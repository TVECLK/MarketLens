package service

import (
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

func NewVocationalEducationService(store *repositories.Store) *CRUDService[models.VocationalEducation] {
	return NewCRUDServiceFromRepository(
		store.VocationalEducations(),
		requireNonEmptyField("level", func(v *models.VocationalEducation) string { return v.Level }),
	)
}
