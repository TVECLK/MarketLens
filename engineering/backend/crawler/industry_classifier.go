package crawler

import (
	"context"

	"marketlens-go-backend/models"
)

// IndustryClassifier walks Industry Sector -> Division -> Group -> Class
// -> Subclass.
type IndustryClassifier struct {
	repo IndustryRepository
	llm  *DeepSeekClient
}

func NewIndustryClassifier(repo IndustryRepository, llm *DeepSeekClient) *IndustryClassifier {
	return &IndustryClassifier{repo: repo, llm: llm}
}

// Classify returns the final industry_subclass_id
func (c *IndustryClassifier) Classify(ctx context.Context, jobText string) (uint, error) {
	levels := []Level{
		{
			Name: "Industry Sector",
			Fetch: func(uint) ([]Option, error) {
				items, err := c.repo.GetAllIndustrySectors()
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustrySector) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Division",
			Fetch: func(sectorID uint) ([]Option, error) {
				items, err := c.repo.GetIndustryDivisionsByIndustrySector(sectorID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustryDivision) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Group",
			Fetch: func(divisionID uint) ([]Option, error) {
				items, err := c.repo.GetIndustryGroupsByIndustryDivision(divisionID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustryGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Class",
			Fetch: func(groupID uint) ([]Option, error) {
				items, err := c.repo.GetIndustryClassesByIndustryGroup(groupID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustryClass) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Industry Subclass",
			Fetch: func(classID uint) ([]Option, error) {
				items, err := c.repo.GetIndustrySubclassesByIndustryClass(classID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.IndustrySubclass) (uint, string) { return i.ID, i.Name }), nil
			},
		},
	}

	return Walk(ctx, c.llm, jobText, levels)
}