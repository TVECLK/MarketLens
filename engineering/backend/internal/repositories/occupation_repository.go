package repositories

import (
	"context"
	"marketlens-go-backend/internal/models"
	"time"
)

// Methods needed for multi-level filtering in the crawler
// Occupation levels by parent ids
func (r *Store) GetSubMajorGroupsByMajorGroup(ctx context.Context, majorGroupID uint) ([]models.SubMajorGroup, error) {
	return newCRUDRepository[models.SubMajorGroup](r.db).getByParent(ctx, "major_group_id", majorGroupID, nil, nil)
}

func (r *Store) GetSubMajorGroupsByMajorGroupForDateRange(ctx context.Context, majorGroupID uint, fromDate, toDate time.Time) ([]models.SubMajorGroup, error) {
	return newCRUDRepository[models.SubMajorGroup](r.db).getByParent(ctx, "major_group_id", majorGroupID, &fromDate, &toDate)
}

func (r *Store) GetMinorGroupsBySubMajorGroup(ctx context.Context, subMajorGroupID uint) ([]models.MinorGroup, error) {
	return newCRUDRepository[models.MinorGroup](r.db).getByParent(ctx, "sub_major_group_id", subMajorGroupID, nil, nil)
}

func (r *Store) GetMinorGroupsBySubMajorGroupForDateRange(ctx context.Context, subMajorGroupID uint, fromDate, toDate time.Time) ([]models.MinorGroup, error) {
	return newCRUDRepository[models.MinorGroup](r.db).getByParent(ctx, "sub_major_group_id", subMajorGroupID, &fromDate, &toDate)
}

func (r *Store) GetUnitGroupsByMinorGroup(ctx context.Context, minorGroupID uint) ([]models.UnitGroup, error) {
	return newCRUDRepository[models.UnitGroup](r.db).getByParent(ctx, "minor_group_id", minorGroupID, nil, nil)
}

func (r *Store) GetUnitGroupsByMinorGroupForDateRange(ctx context.Context, minorGroupID uint, fromDate, toDate time.Time) ([]models.UnitGroup, error) {
	return newCRUDRepository[models.UnitGroup](r.db).getByParent(ctx, "minor_group_id", minorGroupID, &fromDate, &toDate)
}

func (r *Store) GetOccupationGroupsByUnitGroup(ctx context.Context, unitGroupID uint) ([]models.OccupationGroup, error) {
	return newCRUDRepository[models.OccupationGroup](r.db).getByParent(ctx, "unit_group_id", unitGroupID, nil, nil)
}

func (r *Store) GetOccupationGroupsByUnitGroupForDateRange(ctx context.Context, unitGroupID uint, fromDate, toDate time.Time) ([]models.OccupationGroup, error) {
	return newCRUDRepository[models.OccupationGroup](r.db).getByParent(ctx, "unit_group_id", unitGroupID, &fromDate, &toDate)
}

// MajorGroups returns the generic CRUD repository for MajorGroup. Its
// GetAll already supports the date-range listing the old
// GetAllMajorGroupsForDateRange method used to provide separately.
func (r *Store) MajorGroups() *CRUDRepository[models.MajorGroup] {
	return NewCRUDRepository[models.MajorGroup](r.db)
}

// SubMajorGroups returns the generic CRUD repository for SubMajorGroup.
func (r *Store) SubMajorGroups() *CRUDRepository[models.SubMajorGroup] {
	return NewCRUDRepository[models.SubMajorGroup](r.db, "MajorGroup")
}

// MinorGroups returns the generic CRUD repository for MinorGroup.
func (r *Store) MinorGroups() *CRUDRepository[models.MinorGroup] {
	return NewCRUDRepository[models.MinorGroup](r.db, "SubMajorGroup")
}

// UnitGroups returns the generic CRUD repository for UnitGroup.
func (r *Store) UnitGroups() *CRUDRepository[models.UnitGroup] {
	return NewCRUDRepository[models.UnitGroup](r.db, "MinorGroup")
}

// OccupationGroups returns the generic CRUD repository for
// OccupationGroup. Its GetAll already supports the limit/offset the old
// GetAllOccupationGroups method used to provide separately.
func (r *Store) OccupationGroups() *CRUDRepository[models.OccupationGroup] {
	return NewCRUDRepository[models.OccupationGroup](r.db, "UnitGroup")
}
