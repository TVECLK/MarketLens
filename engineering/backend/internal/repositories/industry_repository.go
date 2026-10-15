package repositories

import (
	"context"
	"marketlens-go-backend/internal/models"
	"time"
)

// Industry levels by parent ids
func (r *Store) GetIndustryDivisionsByIndustrySector(ctx context.Context, industrySectorID uint) ([]models.IndustryDivision, error) {
	return newCRUDRepository[models.IndustryDivision](r.db).getByParent(ctx, "industry_sector_id", industrySectorID, nil, nil)
}

func (r *Store) GetIndustryDivisionsByIndustrySectorForDateRange(ctx context.Context, industrySectorID uint, fromDate, toDate time.Time) ([]models.IndustryDivision, error) {
	return newCRUDRepository[models.IndustryDivision](r.db).getByParent(ctx, "industry_sector_id", industrySectorID, &fromDate, &toDate)
}

func (r *Store) GetIndustryGroupsByIndustryDivision(ctx context.Context, industryDivisionID uint) ([]models.IndustryGroup, error) {
	return newCRUDRepository[models.IndustryGroup](r.db).getByParent(ctx, "industry_division_id", industryDivisionID, nil, nil)
}

func (r *Store) GetIndustryGroupsByIndustryDivisionForDateRange(ctx context.Context, industryDivisionID uint, fromDate, toDate time.Time) ([]models.IndustryGroup, error) {
	return newCRUDRepository[models.IndustryGroup](r.db).getByParent(ctx, "industry_division_id", industryDivisionID, &fromDate, &toDate)
}

func (r *Store) GetIndustryClassesByIndustryGroup(ctx context.Context, industryGroupID uint) ([]models.IndustryClass, error) {
	return newCRUDRepository[models.IndustryClass](r.db).getByParent(ctx, "industry_group_id", industryGroupID, nil, nil)
}

func (r *Store) GetIndustryClassesByIndustryGroupForDateRange(ctx context.Context, industryGroupID uint, fromDate, toDate time.Time) ([]models.IndustryClass, error) {
	return newCRUDRepository[models.IndustryClass](r.db).getByParent(ctx, "industry_group_id", industryGroupID, &fromDate, &toDate)
}

func (r *Store) GetIndustrySubclassesByIndustryClass(ctx context.Context, industryClassID uint) ([]models.IndustrySubclass, error) {
	return newCRUDRepository[models.IndustrySubclass](r.db).getByParent(ctx, "industry_class_id", industryClassID, nil, nil)
}

func (r *Store) GetIndustrySubclassesByIndustryClassForDateRange(ctx context.Context, industryClassID uint, fromDate, toDate time.Time) ([]models.IndustrySubclass, error) {
	return newCRUDRepository[models.IndustrySubclass](r.db).getByParent(ctx, "industry_class_id", industryClassID, &fromDate, &toDate)
}

// IndustrySectors returns the generic CRUD repository for IndustrySector.
// Its GetAll already supports the date-range listing the old
// GetAllIndustrySectorsForDateRange method used to provide separately.
func (r *Store) IndustrySectors() *CRUDRepository[models.IndustrySector] {
	return NewCRUDRepository[models.IndustrySector](r.db)
}

// IndustryDivisions returns the generic CRUD repository for IndustryDivision.
func (r *Store) IndustryDivisions() *CRUDRepository[models.IndustryDivision] {
	return NewCRUDRepository[models.IndustryDivision](r.db, "IndustrySector")
}

// IndustryGroups returns the generic CRUD repository for IndustryGroup.
func (r *Store) IndustryGroups() *CRUDRepository[models.IndustryGroup] {
	return NewCRUDRepository[models.IndustryGroup](r.db, "IndustryDivision")
}

// IndustryClasses returns the generic CRUD repository for IndustryClass.
func (r *Store) IndustryClasses() *CRUDRepository[models.IndustryClass] {
	return NewCRUDRepository[models.IndustryClass](r.db, "IndustryGroup")
}

// IndustrySubclasses returns the generic CRUD repository for
// IndustrySubclass. Its GetAll already supports the limit/offset the old
// GetAllIndustrySubclasses method used to provide separately.
func (r *Store) IndustrySubclasses() *CRUDRepository[models.IndustrySubclass] {
	return NewCRUDRepository[models.IndustrySubclass](r.db, "IndustryClass")
}
