package crawler

import (
	"context"

	"marketlens-go-backend/models"
)

// OccupationClassifier walks Major Group -> Sub Major Group -> Minor
// Group -> Unit Group -> Occupation Group.
type OccupationClassifier struct {
	repo OccupationRepository
	llm  *DeepSeekClient
}

func NewOccupationClassifier(repo OccupationRepository, llm *DeepSeekClient) *OccupationClassifier {
	return &OccupationClassifier{repo: repo, llm: llm}
}

// Classify returns the final occupation_group_id, or 0 if no match was
// found at any level.
func (c *OccupationClassifier) Classify(ctx context.Context, jobText string) (uint, error) {
	levels := []Level{
		{
			Name: "Major Group",
			Fetch: func(uint) ([]Option, error) {
				items, err := c.repo.GetAllMajorGroups()
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.MajorGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Sub Major Group",
			Fetch: func(majorID uint) ([]Option, error) {
				items, err := c.repo.GetSubMajorGroupsByMajorGroup(majorID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.SubMajorGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Minor Group",
			Fetch: func(subMajorID uint) ([]Option, error) {
				items, err := c.repo.GetMinorGroupsBySubMajorGroup(subMajorID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.MinorGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Unit Group",
			Fetch: func(minorID uint) ([]Option, error) {
				items, err := c.repo.GetUnitGroupsByMinorGroup(minorID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.UnitGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
		{
			Name: "Occupation Group",
			Fetch: func(unitID uint) ([]Option, error) {
				items, err := c.repo.GetOccupationGroupsByUnitGroup(unitID)
				if err != nil {
					return nil, err
				}
				return toOptions(items, func(i models.OccupationGroup) (uint, string) { return i.ID, i.Name }), nil
			},
		},
	}

	return Walk(ctx, c.llm, jobText, levels)
}