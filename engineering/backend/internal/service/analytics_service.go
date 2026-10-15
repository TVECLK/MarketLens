package service

import (
	"context"
	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"time"
)

// validOccupationLevels / validIndustryLevels / leafLevels are the single
// canonical copy of these rules — the HTTP handlers and the MCP tools both
// call through AnalyticsService, so neither transport needs (or should
// keep) its own copy of "what's a valid level".
var validOccupationLevels = map[string]bool{
	"major-group":      true,
	"sub-major-group":  true,
	"minor-group":      true,
	"unit-group":       true,
	"occupation-group": true,
}

var validIndustryLevels = map[string]bool{
	"industry-sector":   true,
	"industry-division": true,
	"industry-group":    true,
	"industry-class":    true,
	"industry-subclass": true,
}

// leafLevels marks the bottom of each hierarchy, which has no children to return.
var leafLevels = map[string]bool{
	"occupation-group":  true,
	"industry-subclass": true,
}

func validateStandardAndLevel(standard, level string) error {
	switch standard {
	case "occupation":
		if !validOccupationLevels[level] {
			return ValidationErrorf("invalid level %q for standard 'occupation'", level)
		}
	case "industry":
		if !validIndustryLevels[level] {
			return ValidationErrorf("invalid level %q for standard 'industry'", level)
		}
	default:
		return ValidationErrorf("invalid standard %q, must be 'occupation' or 'industry'", standard)
	}
	return nil
}

// AnalyticsService fronts the Store's read-only aggregation queries,
// validating the standard/level/date-range inputs they all share — the
// one place this is checked, instead of the HTTP handlers and MCP tools
// each keeping their own copy.
type AnalyticsService struct {
	store *repositories.Store
}

func NewAnalyticsService(store *repositories.Store) *AnalyticsService {
	return &AnalyticsService{store: store}
}

func (s *AnalyticsService) GetTopHiringEmployersByOccupationLevel(ctx context.Context, level string, id uint, fromDate, toDate time.Time) ([]models.EmployerDemand, error) {
	if err := validateStandardAndLevel("occupation", level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetTopHiringEmployersByOccupationLevel(ctx, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetAllSkillsByOccupationLevel(ctx context.Context, level string, id uint, fromDate, toDate time.Time, limit, offset int) ([]models.SkillDemand, int64, error) {
	if err := validateStandardAndLevel("occupation", level); err != nil {
		return nil, 0, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, 0, err
	}
	return s.store.GetAllSkillsByOccupationLevel(ctx, level, id, fromDate, toDate, limit, offset)
}

func (s *AnalyticsService) GetTop15SkillsByOccupationLevel(ctx context.Context, level string, id uint, fromDate, toDate time.Time) ([]models.SkillDemand, error) {
	if err := validateStandardAndLevel("occupation", level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetTop15SkillsByOccupationLevel(ctx, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetJobTypeByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.JobTypeJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetJobTypeByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetRemoteOnSiteHybridByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) (models.RemoteOnSiteHybridCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return models.RemoteOnSiteHybridCount{}, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return models.RemoteOnSiteHybridCount{}, err
	}
	return s.store.GetRemoteOnSiteHybridByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetVocationalEducationByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.VocationalEducationJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetVocationalEducationByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetGenderByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.GenderJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetGenderByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetFormalityByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.FormalityJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetFormalityByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetEducationLevelByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.EducationLevelJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetEducationLevelByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetProvinceByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.ProvinceJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetProvinceByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetExperienceByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.ExperienceJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetExperienceByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetEmploymentSectorByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.EmploymentSectorJobCount, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetEmploymentSectorByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetLevelChildren(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) ([]models.LevelChildJobCount, string, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return nil, "", err
	}
	if leafLevels[level] {
		return nil, "", ValidationErrorf("'%s' is a leaf level and has no children", level)
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, "", err
	}
	return s.store.GetLevelChildren(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetTotalJobCountByLevel(ctx context.Context, standard, level string, id uint, fromDate, toDate time.Time) (int64, error) {
	if err := validateStandardAndLevel(standard, level); err != nil {
		return 0, err
	}
	if err := validateDateRange(fromDate, toDate); err != nil {
		return 0, err
	}
	return s.store.GetTotalJobCountByLevel(ctx, standard, level, id, fromDate, toDate)
}

func (s *AnalyticsService) GetIndustryJobCountByDateRange(ctx context.Context, fromDate, toDate time.Time) ([]models.IndustryJobCount, error) {
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetIndustryJobCountByDateRange(ctx, fromDate, toDate)
}

func (s *AnalyticsService) GetOccupationJobCountByDateRange(ctx context.Context, fromDate, toDate time.Time) ([]models.OccupationJobCount, error) {
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, err
	}
	return s.store.GetOccupationJobCountByDateRange(ctx, fromDate, toDate)
}

func (s *AnalyticsService) GetTotalVacancyCount(ctx context.Context, fromDate, toDate time.Time) (int64, error) {
	if err := validateDateRange(fromDate, toDate); err != nil {
		return 0, err
	}
	return s.store.GetTotalVacancyCount(ctx, fromDate, toDate)
}

const vacancyTrendDailyThresholdDays = 60

func (s *AnalyticsService) GetVacancyTrend(ctx context.Context, fromDate, toDate time.Time) ([]models.VacancyTrendPoint, string, error) {
	if err := validateDateRange(fromDate, toDate); err != nil {
		return nil, "", err
	}
	diffDays := int(toDate.Sub(fromDate).Hours() / 24)
	if diffDays < vacancyTrendDailyThresholdDays {
		points, err := s.store.GetVacancyTrendDaily(ctx, fromDate, toDate)
		return points, "daily", err
	}
	points, err := s.store.GetVacancyTrendMonthly(ctx, fromDate, toDate)
	return points, "monthly", err
}
