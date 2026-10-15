package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CRUDService fronts 
type CRUDService[T any] struct {
	create      func(ctx context.Context, item *T) error
	getAll      func(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, error)
	getByID     func(ctx context.Context, id uint) (T, error)
	update      func(ctx context.Context, id uint, updates map[string]interface{}) (T, error)
	delete      func(ctx context.Context, id uint) error
	getByParent func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]T, error)

	validate func(ctx context.Context, item *T) error
}

type CRUDOption[T any] func(*CRUDService[T])

// WithParentLookups gives a hierarchy child-level entity its
// "children of a parent id" reads.
func WithParentLookups[T any](getByParent func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]T, error)) CRUDOption[T] {
	return func(s *CRUDService[T]) { s.getByParent = getByParent }
}

func NewCRUDService[T any](
	create func(ctx context.Context, item *T) error,
	getAll func(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, error),
	getByID func(ctx context.Context, id uint) (T, error),
	update func(ctx context.Context, id uint, updates map[string]interface{}) (T, error),
	del func(ctx context.Context, id uint) error,
	validate func(ctx context.Context, item *T) error,
	opts ...CRUDOption[T],
) *CRUDService[T] {
	s := &CRUDService[T]{create: create, getAll: getAll, getByID: getByID, update: update, delete: del, validate: validate}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Repository is the minimal shape *repositories.CRUDRepository[T]
// satisfies — defined here (the consumer), not in internal/repositories,
// so that package never needs to import this one.
type Repository[T any] interface {
	Create(ctx context.Context, item *T) error
	GetAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, error)
	GetByID(ctx context.Context, id uint) (T, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) (T, error)
	Delete(ctx context.Context, id uint) error
}

// NewCRUDServiceFromRepository is NewCRUDService, but taking one
// Repository[T] instead of five separate positional function arguments.
func NewCRUDServiceFromRepository[T any](repo Repository[T], validate func(context.Context, *T) error, opts ...CRUDOption[T]) *CRUDService[T] {
	return NewCRUDService(repo.Create, repo.GetAll, repo.GetByID, repo.Update, repo.Delete, validate, opts...)
}

func (s *CRUDService[T]) Create(ctx context.Context, item *T) error {
	if s.validate != nil {
		if err := s.validate(ctx, item); err != nil {
			return err
		}
	}
	return s.create(ctx, item)
}

// GetAll lists rows 
func (s *CRUDService[T]) GetAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, error) {
	if fromDate != nil && toDate != nil {
		if err := validateDateRange(*fromDate, *toDate); err != nil {
			return nil, err
		}
	}
	return s.getAll(ctx, limit, offset, fromDate, toDate)
}

func (s *CRUDService[T]) GetByID(ctx context.Context, id uint) (T, error) {
	return s.getByID(ctx, id)
}

func (s *CRUDService[T]) Update(ctx context.Context, id uint, updates map[string]interface{}) (T, error) {
	if s.validate != nil {
		current, err := s.getByID(ctx, id)
		if err != nil {
			var zero T
			return zero, err
		}
		b, err := json.Marshal(updates)
		if err != nil {
			var zero T
			return zero, err
		}
		if err := json.Unmarshal(b, &current); err != nil {
			var zero T
			return zero, err
		}
		if err := s.validate(ctx, &current); err != nil {
			var zero T
			return zero, err
		}
	}
	return s.update(ctx, id, updates)
}

func (s *CRUDService[T]) Delete(ctx context.Context, id uint) error {
	return s.delete(ctx, id)
}

// GetByParent lists the children of a parent id, optionally scoped to a
// date range (fromDate/toDate nil means no filter). Only valid for a
// hierarchy child-level entity constructed with WithParentLookups.
func (s *CRUDService[T]) GetByParent(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]T, error) {
	if s.getByParent == nil {
		return nil, fmt.Errorf("GetByParent is not supported for this entity")
	}
	if fromDate != nil && toDate != nil {
		if err := validateDateRange(*fromDate, *toDate); err != nil {
			return nil, err
		}
	}
	return s.getByParent(ctx, parentID, fromDate, toDate)
}
