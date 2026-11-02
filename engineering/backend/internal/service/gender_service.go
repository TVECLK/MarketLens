package service

import (
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

func NewGenderService(store *repositories.Store) *CRUDService[models.Gender] {
	return NewCRUDServiceFromRepository(
		store.Genders(),
		requireNonEmptyField("gender_type", func(g *models.Gender) string { return g.GenderType }),
	)
}
