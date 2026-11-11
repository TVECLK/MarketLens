package repositories

import (
	"context"
	"errors"
	"fmt"
	"marketlens-go-backend/internal/models"

	"gorm.io/gorm"
)

// Sources returns the generic CRUD repository for Source (get by id, list,
// update, soft delete). Create goes through CreateSource instead.
func (r *Store) Sources() *CRUDRepository[models.Source] {
	return NewCRUDRepository[models.Source](r.db)
}

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

// GetSourceByName looks a source up by its exact (case-sensitive) name, the
// value crawlers and manual uploads send in RawJobInput.Source.
func (r *Store) GetSourceByName(ctx context.Context, name string) (models.Source, error) {
	var source models.Source
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&source).Error
	return source, err
}

// CreateSource adds a source, or restores the soft-deleted source with the
// same name so it keeps its id and its jobs stay linked to it. A name that
// belongs to a live source fails on the UNIQUE(name) constraint, which the
// API reports as "already exists".
func (r *Store) CreateSource(ctx context.Context, source *models.Source) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.Source
		err := tx.Unscoped().Where("name = ? AND deleted_at IS NOT NULL", source.Name).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isActive := source.IsActive
			if err := tx.Create(source).Error; err != nil {
				return err
			}
			// Create skips is_active = false because of its default:true tag.
			if !isActive {
				return tx.Model(source).Update("is_active", false).Error
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("look up deleted source %q: %w", source.Name, err)
		}

		// A map so that is_active = false and deleted_at = NULL are written.
		if err := tx.Unscoped().Model(&existing).Updates(map[string]interface{}{
			"type":       source.Type,
			"is_active":  source.IsActive,
			"deleted_at": nil,
		}).Error; err != nil {
			return fmt.Errorf("restore source %d: %w", existing.ID, err)
		}
		return tx.First(source, existing.ID).Error
	})
}
