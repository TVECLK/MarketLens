package service

import (
	"context"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"time"

	"gorm.io/gorm"
)

func NewMajorGroupService(store *repositories.Store) *CRUDService[models.MajorGroup] {
	return NewCRUDServiceFromRepository(
		store.MajorGroups(),
		func(_ context.Context, item *models.MajorGroup) error {
			return validateNameAndCode(item.Name, item.Code)
		},
	)
}

func NewSubMajorGroupService(store *repositories.Store) *CRUDService[models.SubMajorGroup] {
	return NewCRUDServiceFromRepository(
		store.SubMajorGroups(),
		func(ctx context.Context, item *models.SubMajorGroup) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.MajorGroups().GetByID(ctx, item.MajorGroupID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "major_group_id", item.MajorGroupID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.SubMajorGroup, error) {
			if fromDate != nil && toDate != nil {
				return store.GetSubMajorGroupsByMajorGroupForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetSubMajorGroupsByMajorGroup(ctx, parentID)
		}),
	)
}

func NewMinorGroupService(store *repositories.Store) *CRUDService[models.MinorGroup] {
	return NewCRUDServiceFromRepository(
		store.MinorGroups(),
		func(ctx context.Context, item *models.MinorGroup) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.SubMajorGroups().GetByID(ctx, item.SubMajorGroupID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "sub_major_group_id", item.SubMajorGroupID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.MinorGroup, error) {
			if fromDate != nil && toDate != nil {
				return store.GetMinorGroupsBySubMajorGroupForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetMinorGroupsBySubMajorGroup(ctx, parentID)
		}),
	)
}

func NewUnitGroupService(store *repositories.Store) *CRUDService[models.UnitGroup] {
	return NewCRUDServiceFromRepository(
		store.UnitGroups(),
		func(ctx context.Context, item *models.UnitGroup) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.MinorGroups().GetByID(ctx, item.MinorGroupID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "minor_group_id", item.MinorGroupID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.UnitGroup, error) {
			if fromDate != nil && toDate != nil {
				return store.GetUnitGroupsByMinorGroupForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetUnitGroupsByMinorGroup(ctx, parentID)
		}),
	)
}

func NewOccupationGroupService(store *repositories.Store) *CRUDService[models.OccupationGroup] {
	return NewCRUDServiceFromRepository(
		store.OccupationGroups(),
		func(ctx context.Context, item *models.OccupationGroup) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.UnitGroups().GetByID(ctx, item.UnitGroupID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "unit_group_id", item.UnitGroupID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.OccupationGroup, error) {
			if fromDate != nil && toDate != nil {
				return store.GetOccupationGroupsByUnitGroupForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetOccupationGroupsByUnitGroup(ctx, parentID)
		}),
	)
}
