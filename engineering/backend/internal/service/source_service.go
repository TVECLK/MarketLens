package service

import (
	"context"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"strings"
)

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
	return s.store.Sources().GetByID(ctx, id)
}

// GetSourceByName looks a source up by its exact name, as sent by crawlers
// and manual uploads.
func (s *SourceService) GetSourceByName(ctx context.Context, name string) (models.Source, error) {
	return s.store.GetSourceByName(ctx, name)
}

// CreateSource adds a source, or restores a deleted source with the same
// name. isActive nil means active. A name already used by a live source is
// reported as "already exists" by the UNIQUE(name) constraint.
func (s *SourceService) CreateSource(ctx context.Context, name, sourceType string, isActive *bool) (models.Source, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Source{}, ValidationErrorf("name is required")
	}
	t := models.SourceType(sourceType)
	if !validSourceTypes[t] {
		return models.Source{}, ValidationErrorf("invalid type %q, must be 'auto' or 'manual'", sourceType)
	}
	active := true
	if isActive != nil {
		active = *isActive
	}

	source := models.Source{Name: name, Type: t, IsActive: active}
	if err := s.store.CreateSource(ctx, &source); err != nil {
		return models.Source{}, err
	}
	return source, nil
}

// DeleteSource soft-deletes a source. Its jobs keep their link to it, and
// creating a source with the same name later restores it. A missing or
// already-deleted source is gorm.ErrRecordNotFound.
func (s *SourceService) DeleteSource(ctx context.Context, id uint) error {
	return s.store.Sources().Delete(ctx, id)
}

func (s *SourceService) RenameSource(ctx context.Context, id uint, name string) (models.Source, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Source{}, ValidationErrorf("name is required")
	}

	current, err := s.store.Sources().GetByID(ctx, id)
	if err != nil {
		return models.Source{}, err
	}
	if current.Type != models.SourceTypeManual {
		return models.Source{}, ValidationErrorf("only manual sources can be renamed; %q is %s and must match its crawler's SOURCE_NAME", current.Name, current.Type)
	}
	if current.Name == name {
		return current, nil
	}

	return s.store.Sources().Update(ctx, id, map[string]interface{}{"name": name})
}

// UpdateSourceStatus enables or disables a source. A map is used so that
// is_active = false is written (a struct update would skip the zero value).
func (s *SourceService) UpdateSourceStatus(ctx context.Context, id uint, isActive bool) (models.Source, error) {
	return s.store.Sources().Update(ctx, id, map[string]interface{}{"is_active": isActive})
}
