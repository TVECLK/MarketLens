package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type CRUDService[T any] struct {
	create      func(ctx context.Context, item *T) error
	getAll      func(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, int64, error)
	getByID     func(ctx context.Context, id uint) (T, error)
	update      func(ctx context.Context, id uint, updates map[string]interface{}) (T, error)
	delete      func(ctx context.Context, id uint) error
	getByParent func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]T, error)

	validate func(ctx context.Context, item *T) error

	// paginated is set by WithPagination; pageDefault/pageMax are then the
	// limit used when none is given and the largest limit allowed.
	paginated   bool
	pageDefault int
	pageMax     int
}

type CRUDOption[T any] func(*CRUDService[T])

// WithPagination makes GetAll apply a default and a maximum page size, for
// resources too large to ever return whole (occupation groups, industry
// subclasses). Without it GetAll leaves limit alone, and limit=0 means "all
// rows", which the small lookup tables rely on.
func WithPagination[T any](defaultLimit, maxLimit int) CRUDOption[T] {
	return func(s *CRUDService[T]) {
		s.paginated = true
		s.pageDefault = defaultLimit
		s.pageMax = maxLimit
	}
}

// PageLimit returns the limit GetAll will actually use for a requested one,
// so callers can echo it back in their response.
func (s *CRUDService[T]) PageLimit(requested int) int {
	if !s.paginated {
		return requested
	}
	return pageLimit(requested, s.pageDefault, s.pageMax)
}

// WithParentLookups gives a hierarchy child-level entity its
// "children of a parent id" reads.
func WithParentLookups[T any](getByParent func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]T, error)) CRUDOption[T] {
	return func(s *CRUDService[T]) { s.getByParent = getByParent }
}

func NewCRUDService[T any](
	create func(ctx context.Context, item *T) error,
	getAll func(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, int64, error),
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
	GetAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, int64, error)
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
 
func (s *CRUDService[T]) GetAll(ctx context.Context, limit, offset int, fromDate, toDate *time.Time) ([]T, int64, error) {
	if fromDate != nil && toDate != nil {
		if err := validateDateRange(*fromDate, *toDate); err != nil {
			return nil, 0, err
		}
	}
	if s.paginated {
		limit = s.PageLimit(limit)
		offset = pageOffset(offset)
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
