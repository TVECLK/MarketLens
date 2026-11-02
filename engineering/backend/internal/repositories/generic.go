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

// count reports how many rows match the same date-range filter getAll
func (c crudRepository[T]) count(ctx context.Context, fromDate, toDate *time.Time) (int64, error) {
	var zero T
	q := c.db.WithContext(ctx).Model(&zero)
	if fromDate != nil && toDate != nil {
		q = q.Unscoped().
			Where("created_at < ?", exclusiveUpperBound(*toDate)).
			Where("deleted_at IS NULL OR deleted_at >= ?", *fromDate)
	}
	var total int64
	err := q.Count(&total).Error
	return total, err
}

// getAll lists rows
func (c crudRepository[T]) getAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time, preloads ...string) ([]T, int64, error) {
	total, err := c.count(ctx, fromDate, toDate)
	if err != nil {
		return nil, 0, err
	}

	q := c.db.WithContext(ctx)
	for _, p := range preloads {
		q = q.Preload(p)
	}
	if fromDate != nil && toDate != nil {
		q = q.Unscoped().
			Where("created_at < ?", exclusiveUpperBound(*toDate)).
			Where("deleted_at IS NULL OR deleted_at >= ?", *fromDate)
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}

	var items []T
	err = q.Order("id ASC").Find(&items).Error
	return items, total, err
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

type cascadeTarget struct {
	table    string
	fkColumn string
}

func (c crudRepository[T]) delete(ctx context.Context, id uint, cascade []cascadeTarget) error {
	var zero T

	if len(cascade) == 0 {
		res := c.db.WithContext(ctx).Delete(&zero, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}

	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		parentIDs := []uint{id}
		for _, step := range cascade {
			var ids []uint
			if err := tx.Table(step.table).Where(step.fkColumn+" IN (?)", parentIDs).Pluck("id", &ids).Error; err != nil {
				return fmt.Errorf("finding %s for cascade delete: %w", step.table, err)
			}
			if len(ids) == 0 {
				break
			}
			if err := tx.Table(step.table).Where("id IN (?) AND deleted_at IS NULL", ids).Update("deleted_at", time.Now()).Error; err != nil {
				return fmt.Errorf("cascade deleting %s: %w", step.table, err)
			}
			parentIDs = ids
		}

		res := tx.Model(&zero).Where("id = ?", id).Update("deleted_at", time.Now())
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

// getByParent lists the rows
func (c crudRepository[T]) getByParent(ctx context.Context, parentColumn string, parentID uint, fromDate, toDate *time.Time) ([]T, error) {
	q := c.db.WithContext(ctx).Where(parentColumn+" = ?", parentID)
	if fromDate != nil && toDate != nil {
		q = q.Unscoped().
			Where("created_at < ?", exclusiveUpperBound(*toDate)).
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
	cascade  []cascadeTarget
}

func NewCRUDRepository[T any](db *gorm.DB, preloads ...string) *CRUDRepository[T] {
	return &CRUDRepository[T]{repo: newCRUDRepository[T](db), preloads: preloads}
}

func (r *CRUDRepository[T]) withCascade(table, fkColumn string) *CRUDRepository[T] {
	r.cascade = append(r.cascade, cascadeTarget{table: table, fkColumn: fkColumn})
	return r
}

func (r *CRUDRepository[T]) Create(ctx context.Context, item *T) error {
	return r.repo.create(ctx, item)
}

func (r *CRUDRepository[T]) GetAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, int64, error) {
	return r.repo.getAll(ctx, limit, offset, fromDate, toDate, r.preloads...)
}

func (r *CRUDRepository[T]) GetByID(ctx context.Context, id uint) (T, error) {
	return r.repo.getByID(ctx, id, r.preloads...)
}

func (r *CRUDRepository[T]) Update(ctx context.Context, id uint, updates map[string]interface{}) (T, error) {
	return r.repo.update(ctx, id, updates)
}

func (r *CRUDRepository[T]) Delete(ctx context.Context, id uint) error {
	return r.repo.delete(ctx, id, r.cascade)
}
