package repositories

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// crudRepository is a generic, GORM-backed repository for the flat-lookup
// and hierarchy-level entities
type crudRepository[T any] struct {
	db *gorm.DB
}

func newCRUDRepository[T any](db *gorm.DB) crudRepository[T] {
	return crudRepository[T]{db: db}
}

func (c crudRepository[T]) create(ctx context.Context, item *T) error {
	return c.db.WithContext(ctx).Create(item).Error
}

// getAll lists rows
func (c crudRepository[T]) getAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time, preloads ...string) ([]T, error) {
	q := c.db.WithContext(ctx)
	for _, p := range preloads {
		q = q.Preload(p)
	}
	if fromDate != nil && toDate != nil {
		endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		q = q.Unscoped().
			Where("created_at <= ?", endOfToDate).
			Where("deleted_at IS NULL OR deleted_at >= ?", *fromDate)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	var items []T
	err := q.Order("id ASC").Find(&items).Error
	return items, err
}

func (c crudRepository[T]) getByID(ctx context.Context, id uint, preloads ...string) (T, error) {
	var item T
	q := c.db.WithContext(ctx)
	for _, p := range preloads {
		q = q.Preload(p)
	}
	err := q.First(&item, id).Error
	return item, err
}

// update applies a partial update atomically
func (c crudRepository[T]) update(ctx context.Context, id uint, updates map[string]interface{}) (T, error) {
	var zero T
	res := c.db.WithContext(ctx).Model(&zero).Where("id = ?", id).Updates(updates)
	if res.Error != nil {
		return zero, fmt.Errorf("update %T %d: %w", zero, id, res.Error)
	}
	if res.RowsAffected == 0 {
		return zero, gorm.ErrRecordNotFound
	}

	var item T
	if err := c.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return item, fmt.Errorf("reload %T %d: %w", zero, id, err)
	}
	return item, nil
}

// delete 
func (c crudRepository[T]) delete(ctx context.Context, id uint) error {
	var zero T
	res := c.db.WithContext(ctx).Delete(&zero, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// getByParent lists the rows
func (c crudRepository[T]) getByParent(ctx context.Context, parentColumn string, parentID uint, fromDate, toDate *time.Time) ([]T, error) {
	q := c.db.WithContext(ctx).Where(parentColumn+" = ?", parentID)
	if fromDate != nil && toDate != nil {
		endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		q = q.Unscoped().
			Where("created_at <= ?", endOfToDate).
			Where("deleted_at IS NULL OR deleted_at >= ?", *fromDate)
	}
	var items []T
	err := q.Find(&items).Error
	return items, err
}

// CRUDRepository is an exported wrapper around 
type CRUDRepository[T any] struct {
	repo     crudRepository[T]
	preloads []string
}

func NewCRUDRepository[T any](db *gorm.DB, preloads ...string) *CRUDRepository[T] {
	return &CRUDRepository[T]{repo: newCRUDRepository[T](db), preloads: preloads}
}

func (r *CRUDRepository[T]) Create(ctx context.Context, item *T) error {
	return r.repo.create(ctx, item)
}

func (r *CRUDRepository[T]) GetAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, error) {
	return r.repo.getAll(ctx, limit, offset, fromDate, toDate, r.preloads...)
}

func (r *CRUDRepository[T]) GetByID(ctx context.Context, id uint) (T, error) {
	return r.repo.getByID(ctx, id, r.preloads...)
}

func (r *CRUDRepository[T]) Update(ctx context.Context, id uint, updates map[string]interface{}) (T, error) {
	return r.repo.update(ctx, id, updates)
}

func (r *CRUDRepository[T]) Delete(ctx context.Context, id uint) error {
	return r.repo.delete(ctx, id)
}

