package service

import (
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
)

func NewGeoDataService(store *repositories.Store) *CRUDService[models.GeoData] {
	return NewCRUDServiceFromRepository(
		store.GeoData(),
		requireNonEmptyField("province", func(g *models.GeoData) string { return g.Province }),
	)
}
