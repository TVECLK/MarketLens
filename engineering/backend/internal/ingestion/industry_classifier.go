package ingestion

import (
	"context"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/service"
)

// IndustryClassifier walks Industry Sector -> Division -> Group -> Class
// -> Subclass, using the same services the HTTP/MCP layers use.
type IndustryClassifier struct {
	industrySectors    *service.CRUDService[models.IndustrySector]
	industryDivisions  *service.CRUDService[models.IndustryDivision]
	industryGroups     *service.CRUDService[models.IndustryGroup]
	industryClasses    *service.CRUDService[models.IndustryClass]
	industrySubclasses *service.CRUDService[models.IndustrySubclass]
	llm                *DeepSeekClient
}

func NewIndustryClassifier(
	industrySectors *service.CRUDService[models.IndustrySector],
	industryDivisions *service.CRUDService[models.IndustryDivision],
	industryGroups *service.CRUDService[models.IndustryGroup],
	industryClasses *service.CRUDService[models.IndustryClass],
	industrySubclasses *service.CRUDService[models.IndustrySubclass],
	llm *DeepSeekClient,
) *IndustryClassifier {
	return &IndustryClassifier{
		industrySectors:    industrySectors,
		industryDivisions:  industryDivisions,
		industryGroups:     industryGroups,
		industryClasses:    industryClasses,
		industrySubclasses: industrySubclasses,
		llm:                llm,
	}
}

// Classify returns the final industry_subclass_id
func (c *IndustryClassifier) Classify(ctx context.Context, jobText string) (uint, error) {
	levels := []Level{
		{
			Name: "Industry Sector",
			Fetch: func(uint) ([]Option, error) {
				items, err := c.industrySectors.GetAll(ctx, 0, 0, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustrySector) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Division",
			Fetch: func(sectorID uint) ([]Option, error) {
				items, err := c.industryDivisions.GetByParent(ctx, sectorID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustryDivision) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Group",
			Fetch: func(divisionID uint) ([]Option, error) {
				items, err := c.industryGroups.GetByParent(ctx, divisionID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustryGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Class",
			Fetch: func(groupID uint) ([]Option, error) {
				items, err := c.industryClasses.GetByParent(ctx, groupID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustryClass) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Subclass",
			Fetch: func(classID uint) ([]Option, error) {
				items, err := c.industrySubclasses.GetByParent(ctx, classID, nil, nil)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustrySubclass) (uint, string) { return i.ID, i.Name }), nil
			},
		},
	}

	return Walk(ctx, c.llm, jobText, levels)
}
