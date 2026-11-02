package service

import (
	"context"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"time"

	"gorm.io/gorm"
)

func NewIndustrySectorService(store *repositories.Store) *CRUDService[models.IndustrySector] {
	return NewCRUDServiceFromRepository(
		store.IndustrySectors(),
		func(_ context.Context, item *models.IndustrySector) error {
			return validateNameAndCode(item.Name, item.Code)
		},
	)
}

func NewIndustryDivisionService(store *repositories.Store) *CRUDService[models.IndustryDivision] {
	return NewCRUDServiceFromRepository(
		store.IndustryDivisions(),
		func(ctx context.Context, item *models.IndustryDivision) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.IndustrySectors().GetByID(ctx, item.IndustrySectorID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "industry_sector_id", item.IndustrySectorID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.IndustryDivision, error) {
			if fromDate != nil && toDate != nil {
				return store.GetIndustryDivisionsByIndustrySectorForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetIndustryDivisionsByIndustrySector(ctx, parentID)
		}),
	)
}

func NewIndustryGroupService(store *repositories.Store) *CRUDService[models.IndustryGroup] {
	return NewCRUDServiceFromRepository(
		store.IndustryGroups(),
		func(ctx context.Context, item *models.IndustryGroup) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.IndustryDivisions().GetByID(ctx, item.IndustryDivisionID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "industry_division_id", item.IndustryDivisionID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.IndustryGroup, error) {
			if fromDate != nil && toDate != nil {
				return store.GetIndustryGroupsByIndustryDivisionForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetIndustryGroupsByIndustryDivision(ctx, parentID)
		}),
	)
}

func NewIndustryClassService(store *repositories.Store) *CRUDService[models.IndustryClass] {
	return NewCRUDServiceFromRepository(
		store.IndustryClasses(),
		func(ctx context.Context, item *models.IndustryClass) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.IndustryGroups().GetByID(ctx, item.IndustryGroupID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "industry_group_id", item.IndustryGroupID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.IndustryClass, error) {
			if fromDate != nil && toDate != nil {
				return store.GetIndustryClassesByIndustryGroupForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetIndustryClassesByIndustryGroup(ctx, parentID)
		}),
	)
}

func NewIndustrySubclassService(store *repositories.Store) *CRUDService[models.IndustrySubclass] {
	return NewCRUDServiceFromRepository(
		store.IndustrySubclasses(),
		func(ctx context.Context, item *models.IndustrySubclass) error {
			if err := validateNameAndCode(item.Name, item.Code); err != nil {
				return err
			}
			_, err := store.IndustryClasses().GetByID(ctx, item.IndustryClassID)
			return validateParentExists(err, gorm.ErrRecordNotFound, "industry_class_id", item.IndustryClassID)
		},
		WithParentLookups(func(ctx context.Context, parentID uint, fromDate, toDate *time.Time) ([]models.IndustrySubclass, error) {
			if fromDate != nil && toDate != nil {
				return store.GetIndustrySubclassesByIndustryClassForDateRange(ctx, parentID, *fromDate, *toDate)
			}
			return store.GetIndustrySubclassesByIndustryClass(ctx, parentID)
		}),
	)
}
