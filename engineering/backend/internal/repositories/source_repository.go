package repositories

import (
	"context"
	"fmt"
	"marketlens-go-backend/internal/models"

	"gorm.io/gorm"
)

// SourceFilter narrows ListSources. A nil field means "don't filter on it".
type SourceFilter struct {
	Type     *models.SourceType
	IsActive *bool
}

// ListSources returns the sources matching filter, ordered by id.
func (r *Store) ListSources(ctx context.Context, filter SourceFilter) ([]models.Source, error) {
	q := r.db.WithContext(ctx)
	if filter.Type != nil {
		q = q.Where("type = ?", *filter.Type)
	}
	if filter.IsActive != nil {
		q = q.Where("is_active = ?", *filter.IsActive)
	}

	var sources []models.Source
	err := q.Order("id ASC").Find(&sources).Error
	return sources, err
}

func (r *Store) GetSourceByID(ctx context.Context, id uint) (models.Source, error) {
	var source models.Source
	err := r.db.WithContext(ctx).First(&source, id).Error
	return source, err
}

// GetSourceByName looks a source up by its exact (case-sensitive) name, the
// value crawlers and manual uploads send in RawJobInput.Source.
func (r *Store) GetSourceByName(ctx context.Context, name string) (models.Source, error) {
	var source models.Source
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&source).Error
	return source, err
}

// UpdateSourceStatus enables or disables a source and records who did it.
// A map is used so that is_active = false is written (a struct update would
// skip the zero value).
func (r *Store) UpdateSourceStatus(ctx context.Context, id uint, isActive bool, updatedBy string) (models.Source, error) {
	res := r.db.WithContext(ctx).Model(&models.Source{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active":  isActive,
		"updated_by": updatedBy,
	})
	if res.Error != nil {
		return models.Source{}, fmt.Errorf("update source %d status: %w", id, res.Error)
	}
	if res.RowsAffected == 0 {
		return models.Source{}, gorm.ErrRecordNotFound
	}

	source, err := r.GetSourceByID(ctx, id)
	if err != nil {
		return source, fmt.Errorf("reload source %d: %w", id, err)
	}
	return source, nil
}
