package ingestion

import (
	"context"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/service"
)

// OccupationClassifier walks Major Group -> Sub Major Group -> Minor
// Group -> Unit Group -> Occupation Group, using the same services the
// HTTP/MCP layers use.
type OccupationClassifier struct {
	majorGroups      *service.CRUDService[models.MajorGroup]
	subMajorGroups   *service.CRUDService[models.SubMajorGroup]
	minorGroups      *service.CRUDService[models.MinorGroup]
	unitGroups       *service.CRUDService[models.UnitGroup]
	occupationGroups *service.CRUDService[models.OccupationGroup]
	llm              *DeepSeekClient
}

func NewOccupationClassifier(
	majorGroups *service.CRUDService[models.MajorGroup],
	subMajorGroups *service.CRUDService[models.SubMajorGroup],
	minorGroups *service.CRUDService[models.MinorGroup],
	unitGroups *service.CRUDService[models.UnitGroup],
	occupationGroups *service.CRUDService[models.OccupationGroup],
	llm *DeepSeekClient,
) *OccupationClassifier {
	return &OccupationClassifier{
		majorGroups:      majorGroups,
		subMajorGroups:   subMajorGroups,
		minorGroups:      minorGroups,
		unitGroups:       unitGroups,
		occupationGroups: occupationGroups,
		llm:              llm,
	}
}

// Classify returns the final occupation_group_id, or 0 if no match was
// found at any level.
func (c *OccupationClassifier) Classify(ctx context.Context, jobText string) (uint, error) {
	levels := []Level{
		{
			Name: "Major Group",
			Fetch: func(uint) ([]Option, error) {
				items, err := c.majorGroups.GetAll(ctx, 0, 0, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.MajorGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Sub Major Group",
			Fetch: func(majorID uint) ([]Option, error) {
				items, err := c.subMajorGroups.GetByParent(ctx, majorID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.SubMajorGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Minor Group",
			Fetch: func(subMajorID uint) ([]Option, error) {
				items, err := c.minorGroups.GetByParent(ctx, subMajorID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.MinorGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Unit Group",
			Fetch: func(minorID uint) ([]Option, error) {
				items, err := c.unitGroups.GetByParent(ctx, minorID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.UnitGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Occupation Group",
			Fetch: func(unitID uint) ([]Option, error) {
				items, err := c.occupationGroups.GetByParent(ctx, unitID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.OccupationGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
	}

	return Walk(ctx, c.llm, jobText, levels)
}
