package service

import (
	"context"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"strings"
)

// SourceService fronts the Store's source methods. Sources are seeded, not
// created or deleted through the API; the only change allowed is turning one
// on or off.
type SourceService struct {
	store *repositories.Store
}

func NewSourceService(store *repositories.Store) *SourceService {
	return &SourceService{store: store}
}

// validSourceTypes are the values accepted for the type filter, matching
// source_type_enum.
var validSourceTypes = map[models.SourceType]bool{
	models.SourceTypeAuto:   true,
	models.SourceTypeManual: true,
}

// ListSources returns sources filtered by type ("" for any) and active state
// (nil for any).
func (s *SourceService) ListSources(ctx context.Context, sourceType string, isActive *bool) ([]models.Source, error) {
	filter := repositories.SourceFilter{IsActive: isActive}
	if sourceType != "" {
		t := models.SourceType(sourceType)
		if !validSourceTypes[t] {
			return nil, ValidationErrorf("invalid type %q, must be 'auto' or 'manual'", sourceType)
		}
		filter.Type = &t
	}
	return s.store.ListSources(ctx, filter)
}

// ListActiveAutoSources returns the sources the crawler should run.
func (s *SourceService) ListActiveAutoSources(ctx context.Context) ([]models.Source, error) {
	sourceType := models.SourceTypeAuto
	isActive := true
	return s.store.ListSources(ctx, repositories.SourceFilter{Type: &sourceType, IsActive: &isActive})
}

func (s *SourceService) GetSourceByID(ctx context.Context, id uint) (models.Source, error) {
	return s.store.GetSourceByID(ctx, id)
}

// UpdateSourceStatus enables or disables a source.
func (s *SourceService) UpdateSourceStatus(ctx context.Context, id uint, isActive bool, updatedBy string) (models.Source, error) {
	updatedBy = strings.TrimSpace(updatedBy)
	if updatedBy == "" {
		return models.Source{}, ValidationErrorf("updated_by is required")
	}
	return s.store.UpdateSourceStatus(ctx, id, isActive, updatedBy)
}
