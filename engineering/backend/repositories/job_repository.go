package repositories

import (
	"errors"
	"fmt"
	"marketlens-go-backend/models"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// ErrPermanentSaveFailure marks a job save failure as non-retryable: a
// missing/invalid foreign key (e.g. a province not registered in
// geo_data) or a Postgres integrity-constraint violation (SQLSTATE class
// 23: foreign key, unique, not-null, check). Retrying the identical save
// will fail the same way every time.
var ErrPermanentSaveFailure = errors.New("permanent save failure")

// isPermanentDBError reports whether err is a Postgres integrity-
// constraint violation, as surfaced by gorm's postgres driver
// (jackc/pgx) via *pgconn.PgError.Code (SQLSTATE).
func isPermanentDBError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "23")
	}
	return false
}

// wrapSaveErr classifies err as permanent (constraint violation) or
// transient (anything else — connection issues, deadlocks, etc.).
func wrapSaveErr(err error, msg string) error {
	if isPermanentDBError(err) {
		return fmt.Errorf("%w: %s: %v", ErrPermanentSaveFailure, msg, err)
	}
	return fmt.Errorf("%s: %w", msg, err)
}



type JobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

//This function pings the database connection to confirm it's reachable - used for Kubernetes readiness probes
func (r *JobRepository) Ping() error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

//This function builds a subquery of job_post.ids that fall under the given
//occupation/industry hierarchy level and id — used to scope other aggregations
//(like employment sector breakdown) to that level.
func (r *JobRepository) buildJobPostIDsForLevel(standard, level string, id uint, fromDate, toDate time.Time) (*gorm.DB, error) {
	toDateExclusive := toDate.AddDate(0, 0, 1)

	if standard == "occupation" {
		switch level {
		case "major-group":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Joins("JOIN unit_group ON unit_group.id = occupation_group.unit_group_id").
				Joins("JOIN minor_group ON minor_group.id = unit_group.minor_group_id").
				Joins("JOIN sub_major_group ON sub_major_group.id = minor_group.sub_major_group_id").
				Where("sub_major_group.major_group_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		case "sub-major-group":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Joins("JOIN unit_group ON unit_group.id = occupation_group.unit_group_id").
				Joins("JOIN minor_group ON minor_group.id = unit_group.minor_group_id").
				Where("minor_group.sub_major_group_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil


		case "minor-group":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Joins("JOIN unit_group ON unit_group.id = occupation_group.unit_group_id").
				Where("unit_group.minor_group_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		case "unit-group":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Where("occupation_group.unit_group_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		case "occupation-group":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Where("meta_data.occupation_group_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		default:
			return nil, fmt.Errorf("invalid level '%s' for standard 'occupation'", level)
		}
	} else if standard == "industry" {
		switch level {
		case "industry-sector":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Joins("JOIN industry_class ON industry_class.id = industry_subclass.industry_class_id").
				Joins("JOIN industry_group ON industry_group.id = industry_class.industry_group_id").
				Joins("JOIN industry_division ON industry_division.id = industry_group.industry_division_id").
				Where("industry_division.industry_sector_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		case "industry-division":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Joins("JOIN industry_class ON industry_class.id = industry_subclass.industry_class_id").
				Joins("JOIN industry_group ON industry_group.id = industry_class.industry_group_id").
				Where("industry_group.industry_division_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		case "industry-group":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Joins("JOIN industry_class ON industry_class.id = industry_subclass.industry_class_id").
				Where("industry_class.industry_group_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		case "industry-class":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Where("industry_subclass.industry_class_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		case "industry-subclass":
			return r.db.Table("job_post").
				Select("job_post.id").
				Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
				Where("meta_data.industry_subclass_id = ?", id).
				Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive), nil

		default:
			return nil, fmt.Errorf("invalid level '%s' for standard 'industry'", level)
		}
	}

	return nil, fmt.Errorf("invalid standard '%s', must be 'occupation' or 'industry'", standard)
}

//This function returns the top hiring employers (by summed vacancy count) for jobs
//under the given occupation hierarchy level and id, filtered by date range.
func (r *JobRepository) GetTopHiringEmployersByOccupationLevel(level string, id uint, fromDate, toDate time.Time) ([]models.EmployerDemand, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel("occupation", level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	var results []models.EmployerDemand
	err = r.db.Table("employer").
		Select("employer.id, employer.name, SUM(job_post.no_of_vacancies) AS open_job_count").
		Joins("JOIN job_post ON job_post.employer_id = employer.id").
		Where("job_post.id IN (?)", jobPostIDs).
		Group("employer.id, employer.name").
		Order("open_job_count DESC").
		Limit(5).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query top hiring employers for occupation/%s/%d: %w", level, id, err)
	}

	return results, nil
}

//This function returns all skills (paginated, by summed vacancy count) for jobs
//under the given occupation hierarchy level and id, filtered by date range.
func (r *JobRepository) GetAllSkillsByOccupationLevel(level string, id uint, fromDate, toDate time.Time, limit, offset int) ([]models.SkillDemand, int64, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel("occupation", level, id, fromDate, toDate)
	if err != nil {
		return nil, 0, err
	}

	baseQuery := r.db.Table("skills").
		Joins("JOIN job_post_skills ON job_post_skills.skill_id = skills.id").
		Joins("JOIN job_post ON job_post.id = job_post_skills.job_post_id").
		Where("job_post.id IN (?)", jobPostIDs)

	var total int64
	if err := baseQuery.Session(&gorm.Session{}).
		Distinct("skills.id").
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count skills for occupation/%s/%d: %w", level, id, err)
	}

	var results []models.SkillDemand
	query := baseQuery.Session(&gorm.Session{}).
		Select("skills.id, skills.skill, SUM(job_post.no_of_vacancies) AS open_job_count").
		Group("skills.id, skills.skill").
		Order("open_job_count DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Scan(&results).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query all skills for occupation/%s/%d: %w", level, id, err)
	}

	return results, total, nil
}

//This function returns the top 15 skills (by summed vacancy count) for jobs
//under the given occupation hierarchy level and id, filtered by date range.
func (r *JobRepository) GetTop15SkillsByOccupationLevel(level string, id uint, fromDate, toDate time.Time) ([]models.SkillDemand, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel("occupation", level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	var results []models.SkillDemand
	err = r.db.Table("skills").
		Select("skills.id, skills.skill, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Joins("JOIN job_post_skills ON job_post_skills.skill_id = skills.id").
		Joins("JOIN job_post ON job_post.id = job_post_skills.job_post_id").
		Where("job_post.id IN (?)", jobPostIDs).
		Group("skills.id, skills.skill").
		Order("open_job_count DESC").
		Limit(15).
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query top 15 skills for occupation/%s/%d: %w", level, id, err)
	}

	return results, nil
}

//This function returns job type breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetJobTypeByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.JobTypeJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	var results []models.JobTypeJobCount
	err = r.db.Table("job_type").
		Select("job_type.id, job_type.type, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Joins(
			"LEFT JOIN job_post ON job_post.job_type_id = job_type.id AND job_post.id IN (?)",
			jobPostIDs,
		).
		Group("job_type.id, job_type.type").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query job type breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns remote vs on-site job counts for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetRemoteOnSiteHybridByLevel(standard, level string, id uint, fromDate, toDate time.Time) (models.RemoteOnSiteHybridCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return models.RemoteOnSiteHybridCount{}, err
	}

	type workModeRow struct {
		WorkMode models.WorkMode
		Count    int64
	}
	var rows []workModeRow

	err = r.db.Table("job_post").
		Select("job_post.work_mode, COALESCE(SUM(job_post.no_of_vacancies), 0) AS count").
		Where("job_post.id IN (?)", jobPostIDs).
		Group("job_post.work_mode").
		Scan(&rows).Error

	if err != nil {
		return models.RemoteOnSiteHybridCount{}, fmt.Errorf("failed to query remote/on-site breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	var result models.RemoteOnSiteHybridCount
	for _, row := range rows {
		switch row.WorkMode {
		case models.WorkModeRemote:
			result.RemoteCount = row.Count
		case models.WorkModeOnsite:
			result.OnSiteCount = row.Count
		case models.WorkModeHybrid:
			result.HybridCount = row.Count
		}
	}

	return result, nil
}

//This function returns vocational education breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetVocationalEducationByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.VocationalEducationJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var results []models.VocationalEducationJobCount
	err = r.db.Table("vocational_education").
		Select("vocational_education.id, vocational_education.level, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("vocational_education.created_at <= ?", endOfToDate).
		Where("vocational_education.deleted_at IS NULL OR vocational_education.deleted_at >= ?", fromDate).
		Joins(
			"LEFT JOIN meta_data ON meta_data.vocational_education_id = vocational_education.id "+
				"AND meta_data.job_post_id IN (?)",
			jobPostIDs,
		).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("vocational_education.id, vocational_education.level").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query vocational education breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns gender breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetGenderByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.GenderJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var results []models.GenderJobCount
	err = r.db.Table("gender").
		Select("gender.id, gender.gender_type, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("gender.created_at <= ?", endOfToDate).
 		Where("gender.deleted_at IS NULL OR gender.deleted_at >= ?", fromDate).
		Joins(
			"LEFT JOIN meta_data ON meta_data.gender_id = gender.id "+
				"AND meta_data.job_post_id IN (?)",
			jobPostIDs,
		).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("gender.id, gender.gender_type").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query gender breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns formality breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetFormalityByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.FormalityJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var results []models.FormalityJobCount
	err = r.db.Table("formality").
		Select("formality.id, formality.formality_type, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("formality.created_at <= ?", endOfToDate).
 		Where("formality.deleted_at IS NULL OR formality.deleted_at >= ?", fromDate).
		Joins(
			"LEFT JOIN meta_data ON meta_data.formality_id = formality.id "+
				"AND meta_data.job_post_id IN (?)",
			jobPostIDs,
		).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("formality.id, formality.formality_type").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query formality breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns education level breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetEducationLevelByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.EducationLevelJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var results []models.EducationLevelJobCount
	err = r.db.Table("education_level").
		Select("education_level.id, education_level.level, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("education_level.created_at <= ?", endOfToDate).
 		Where("education_level.deleted_at IS NULL OR education_level.deleted_at >= ?", fromDate).
		Joins(
			"LEFT JOIN meta_data ON meta_data.education_level_id = education_level.id "+
				"AND meta_data.job_post_id IN (?)",
			jobPostIDs,
		).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("education_level.id, education_level.level").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query education level breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns province breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetProvinceByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.ProvinceJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	var results []models.ProvinceJobCount
	err = r.db.Table("geo_data").
		Select("geo_data.id, geo_data.province, geo_data.latitude, geo_data.longitude, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Joins(
			"LEFT JOIN meta_data ON meta_data.geo_data_id = geo_data.id "+
				"AND meta_data.job_post_id IN (?)",
			jobPostIDs,
		).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("geo_data.id, geo_data.province, geo_data.latitude, geo_data.longitude").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query province breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns experience-level breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetExperienceByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.ExperienceJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var results []models.ExperienceJobCount
	err = r.db.Table("experience").
		Select("experience.id, experience.name, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("experience.created_at <= ?", endOfToDate).
 		Where("experience.deleted_at IS NULL OR experience.deleted_at >= ?", fromDate).
		Joins(
			"LEFT JOIN meta_data ON meta_data.experience_id = experience.id "+
				"AND meta_data.job_post_id IN (?)",
			jobPostIDs,
		).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("experience.id, experience.name").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query experience breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns employment sector breakdown (with job counts) for the given
//occupation/industry hierarchy level and id, filtered by date range.
func (r *JobRepository) GetEmploymentSectorByLevel(standard, level string, id uint, fromDate, toDate time.Time) ([]models.EmploymentSectorJobCount, error) {
	jobPostIDs, err := r.buildJobPostIDsForLevel(standard, level, id, fromDate, toDate)
	if err != nil {
		return nil, err
	}

	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var results []models.EmploymentSectorJobCount
	err = r.db.Table("employment_sector").
		Select("employment_sector.id, employment_sector.sector, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("employment_sector.created_at <= ?", endOfToDate).
 		Where("employment_sector.deleted_at IS NULL OR employment_sector.deleted_at >= ?", fromDate).
		Joins(
			"LEFT JOIN meta_data ON meta_data.employment_sector_id = employment_sector.id "+
				"AND meta_data.job_post_id IN (?)",
			jobPostIDs,
		).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("employment_sector.id, employment_sector.sector").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query employment sector breakdown for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, nil
}

//This function returns the immediate child-level entities under a given occupation/industry
//hierarchy level and id, each with its aggregated job count for the given date range.
func (r *JobRepository) GetLevelChildren(standard, level string, id uint, fromDate, toDate time.Time) ([]models.LevelChildJobCount, string, error) {
	var results []models.LevelChildJobCount
	var childLevel string
	var query *gorm.DB

	toDateExclusive := toDate.AddDate(0, 0, 1)

	if standard == "occupation" {
		switch level {
		case "major-group":
			childLevel = "sub-major-group"
			query = r.db.Table("sub_major_group").
				Select("sub_major_group.id, sub_major_group.name, sub_major_group.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("sub_major_group.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
                Where("sub_major_group.deleted_at IS NULL OR sub_major_group.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN minor_group ON minor_group.sub_major_group_id = sub_major_group.id").
				Joins("LEFT JOIN unit_group ON unit_group.minor_group_id = minor_group.id").
				Joins("LEFT JOIN occupation_group ON occupation_group.unit_group_id = unit_group.id").
				Joins("LEFT JOIN meta_data ON meta_data.occupation_group_id = occupation_group.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("sub_major_group.major_group_id = ?", id).
				Group("sub_major_group.id, sub_major_group.name, sub_major_group.code")

		case "sub-major-group":
			childLevel = "minor-group"
			query = r.db.Table("minor_group").
				Select("minor_group.id, minor_group.name, minor_group.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("minor_group.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
                Where("minor_group.deleted_at IS NULL OR minor_group.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN unit_group ON unit_group.minor_group_id = minor_group.id").
				Joins("LEFT JOIN occupation_group ON occupation_group.unit_group_id = unit_group.id").
				Joins("LEFT JOIN meta_data ON meta_data.occupation_group_id = occupation_group.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("minor_group.sub_major_group_id = ?", id).
				Group("minor_group.id, minor_group.name, minor_group.code")

		case "minor-group":
			childLevel = "unit-group"
			query = r.db.Table("unit_group").
				Select("unit_group.id, unit_group.name, unit_group.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("unit_group.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
	            Where("unit_group.deleted_at IS NULL OR unit_group.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN occupation_group ON occupation_group.unit_group_id = unit_group.id").
				Joins("LEFT JOIN meta_data ON meta_data.occupation_group_id = occupation_group.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("unit_group.minor_group_id = ?", id).
				Group("unit_group.id, unit_group.name, unit_group.code")

		case "unit-group":
			childLevel = "occupation-group"
			query = r.db.Table("occupation_group").
				Select("occupation_group.id, occupation_group.name, occupation_group.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("occupation_group.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
	            Where("occupation_group.deleted_at IS NULL OR occupation_group.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN meta_data ON meta_data.occupation_group_id = occupation_group.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("occupation_group.unit_group_id = ?", id).
				Group("occupation_group.id, occupation_group.name, occupation_group.code")

		case "occupation-group":
			return nil, "", fmt.Errorf("'occupation-group' is a leaf level and has no children")

		default:
			return nil, "", fmt.Errorf("invalid level '%s' for standard 'occupation'", level)
		}
	} else if standard == "industry" {
		switch level {
		case "industry-sector":
			childLevel = "industry-division"
			query = r.db.Table("industry_division").
				Select("industry_division.id, industry_division.name, industry_division.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("industry_division.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
 				Where("industry_division.deleted_at IS NULL OR industry_division.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN industry_group ON industry_group.industry_division_id = industry_division.id").
				Joins("LEFT JOIN industry_class ON industry_class.industry_group_id = industry_group.id").
				Joins("LEFT JOIN industry_subclass ON industry_subclass.industry_class_id = industry_class.id").
				Joins("LEFT JOIN meta_data ON meta_data.industry_subclass_id = industry_subclass.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("industry_division.industry_sector_id = ?", id).
				Group("industry_division.id, industry_division.name, industry_division.code")

		case "industry-division":
			childLevel = "industry-group"
			query = r.db.Table("industry_group").
				Select("industry_group.id, industry_group.name, industry_group.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("industry_group.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
 				Where("industry_group.deleted_at IS NULL OR industry_group.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN industry_class ON industry_class.industry_group_id = industry_group.id").
				Joins("LEFT JOIN industry_subclass ON industry_subclass.industry_class_id = industry_class.id").
				Joins("LEFT JOIN meta_data ON meta_data.industry_subclass_id = industry_subclass.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("industry_group.industry_division_id = ?", id).
				Group("industry_group.id, industry_group.name, industry_group.code")

		case "industry-group":
			childLevel = "industry-class"
			query = r.db.Table("industry_class").
				Select("industry_class.id, industry_class.name, industry_class.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("industry_class.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
  				Where("industry_class.deleted_at IS NULL OR industry_class.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN industry_subclass ON industry_subclass.industry_class_id = industry_class.id").
				Joins("LEFT JOIN meta_data ON meta_data.industry_subclass_id = industry_subclass.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("industry_class.industry_group_id = ?", id).
				Group("industry_class.id, industry_class.name, industry_class.code")

		case "industry-class":
			childLevel = "industry-subclass"
			query = r.db.Table("industry_subclass").
				Select("industry_subclass.id, industry_subclass.name, industry_subclass.code, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
				Where("industry_subclass.created_at <= ?", toDate.Add(23*time.Hour+59*time.Minute+59*time.Second)).
 				Where("industry_subclass.deleted_at IS NULL OR industry_subclass.deleted_at >= ?", fromDate).
				Joins("LEFT JOIN meta_data ON meta_data.industry_subclass_id = industry_subclass.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
				Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
				Where("industry_subclass.industry_class_id = ?", id).
				Group("industry_subclass.id, industry_subclass.name, industry_subclass.code")

		case "industry-subclass":
			return nil, "", fmt.Errorf("'industry-subclass' is a leaf level and has no children")

		default:
			return nil, "", fmt.Errorf("invalid level '%s' for standard 'industry'", level)
		}
	} else {
		return nil, "", fmt.Errorf("invalid standard '%s', must be 'occupation' or 'industry'", standard)
	}

	err := query.Order("open_job_count DESC").Scan(&results).Error
	if err != nil {
		return nil, "", fmt.Errorf("failed to query children for %s/%s/%d: %w", standard, level, id, err)
	}

	return results, childLevel, nil
}

//This function returns the total job count for a given occupation/industry hierarchy level and id,
//filtered by date range. standard is "occupation" or "industry"; level depends on the standard.
func (r *JobRepository) GetTotalJobCountByLevel(standard, level string, id uint, fromDate, toDate time.Time) (int64, error) {
	toDateExclusive := toDate.AddDate(0, 0, 1)

	query := r.db.Table("job_post").
		Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
		Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive)

	if standard == "occupation" {
		switch level {
		case "major-group":
			query = query.
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Joins("JOIN unit_group ON unit_group.id = occupation_group.unit_group_id").
				Joins("JOIN minor_group ON minor_group.id = unit_group.minor_group_id").
				Joins("JOIN sub_major_group ON sub_major_group.id = minor_group.sub_major_group_id").
				Where("sub_major_group.major_group_id = ?", id)

		case "sub-major-group":
			query = query.
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Joins("JOIN unit_group ON unit_group.id = occupation_group.unit_group_id").
				Joins("JOIN minor_group ON minor_group.id = unit_group.minor_group_id").
				Where("minor_group.sub_major_group_id = ?", id)

		case "minor-group":
			query = query.
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Joins("JOIN unit_group ON unit_group.id = occupation_group.unit_group_id").
				Where("unit_group.minor_group_id = ?", id)

		case "unit-group":
			query = query.
				Joins("JOIN occupation_group ON occupation_group.id = meta_data.occupation_group_id").
				Where("occupation_group.unit_group_id = ?", id)

		case "occupation-group":
			query = query.Where("meta_data.occupation_group_id = ?", id)

		default:
			return 0, fmt.Errorf("invalid level '%s' for standard 'occupation'", level)
		}
	} else if standard == "industry" {
		switch level {
		case "industry-sector":
			query = query.
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Joins("JOIN industry_class ON industry_class.id = industry_subclass.industry_class_id").
				Joins("JOIN industry_group ON industry_group.id = industry_class.industry_group_id").
				Joins("JOIN industry_division ON industry_division.id = industry_group.industry_division_id").
				Where("industry_division.industry_sector_id = ?", id)

		case "industry-division":
			query = query.
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Joins("JOIN industry_class ON industry_class.id = industry_subclass.industry_class_id").
				Joins("JOIN industry_group ON industry_group.id = industry_class.industry_group_id").
				Where("industry_group.industry_division_id = ?", id)

		case "industry-group":
			query = query.
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Joins("JOIN industry_class ON industry_class.id = industry_subclass.industry_class_id").
				Where("industry_class.industry_group_id = ?", id)

		case "industry-class":
			query = query.
				Joins("JOIN industry_subclass ON industry_subclass.id = meta_data.industry_subclass_id").
				Where("industry_subclass.industry_class_id = ?", id)

		case "industry-subclass":
			query = query.Where("meta_data.industry_subclass_id = ?", id)

		default:
			return 0, fmt.Errorf("invalid level '%s' for standard 'industry'", level)
		}
	} else {
		return 0, fmt.Errorf("invalid standard '%s', must be 'occupation' or 'industry'", standard)
	}

	var total int64
	err := query.Select("COALESCE(SUM(job_post.no_of_vacancies), 0)").Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query total job count for %s/%s/%d: %w", standard, level, id, err)
	}

	return total, nil
}

//This function returns vacancy counts grouped by industry sector, for jobs posted within the given date range
func (r *JobRepository) GetIndustryJobCountByDateRange(fromDate, toDate time.Time) ([]models.IndustryJobCount, error) {
	var results []models.IndustryJobCount
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	toDateExclusive := toDate.AddDate(0, 0, 1)

	err := r.db.Table("industry_sector").
		Select("industry_sector.id, industry_sector.name, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("industry_sector.created_at <= ?", endOfToDate).
 		Where("industry_sector.deleted_at IS NULL OR industry_sector.deleted_at >= ?", fromDate).
		Joins("LEFT JOIN industry_division ON industry_division.industry_sector_id = industry_sector.id").
		Joins("LEFT JOIN industry_group ON industry_group.industry_division_id = industry_division.id").
		Joins("LEFT JOIN industry_class ON industry_class.industry_group_id = industry_group.id").
		Joins("LEFT JOIN industry_subclass ON industry_subclass.industry_class_id = industry_class.id").
		Joins("LEFT JOIN meta_data ON meta_data.industry_subclass_id = industry_subclass.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("industry_sector.id, industry_sector.name").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query industry job count by date range: %w", err)
	}

	return results, nil
}

//This function returns vacancy counts grouped by major group (occupation), for jobs posted within the given date range
func (r *JobRepository) GetOccupationJobCountByDateRange(fromDate, toDate time.Time) ([]models.OccupationJobCount, error) {
	var results []models.OccupationJobCount
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	toDateExclusive := toDate.AddDate(0, 0, 1)

	err := r.db.Table("major_group").
		Select("major_group.id, major_group.name, COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count").
		Where("major_group.created_at <= ?", endOfToDate).
 		Where("major_group.deleted_at IS NULL OR major_group.deleted_at >= ?", fromDate).
		Joins("LEFT JOIN sub_major_group ON sub_major_group.major_group_id = major_group.id").
		Joins("LEFT JOIN minor_group ON minor_group.sub_major_group_id = sub_major_group.id").
		Joins("LEFT JOIN unit_group ON unit_group.minor_group_id = minor_group.id").
		Joins("LEFT JOIN occupation_group ON occupation_group.unit_group_id = unit_group.id").
		Joins("LEFT JOIN meta_data ON meta_data.occupation_group_id = occupation_group.id AND meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
		Joins("LEFT JOIN job_post ON job_post.id = meta_data.job_post_id").
		Group("major_group.id, major_group.name").
		Order("open_job_count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query occupation job count by date range: %w", err)
	}

	return results, nil
}

//This function returns the total sum of no_of_vacancies for jobs posted within the given date range
func (r *JobRepository) GetTotalVacancyCount(fromDate, toDate time.Time) (int64, error) {
	var total int64
	toDateExclusive := toDate.AddDate(0, 0, 1)

	err := r.db.Table("job_post").
		Select("COALESCE(SUM(job_post.no_of_vacancies), 0)").
		Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
		Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", fromDate, toDateExclusive).
		Scan(&total).Error

	if err != nil {
		return 0, fmt.Errorf("failed to query total vacancy count: %w", err)
	}

	return total, nil
}

// Get vacancy trend for a selected date range
//This function returns job counts grouped day by day within the given date range
func (r *JobRepository) GetVacancyTrendDaily(fromDate, toDate time.Time) ([]models.VacancyTrendPoint, error) {
	var rows []struct {
		PeriodStart  time.Time
		OpenJobCount int64
	}

	err := r.db.Raw(`
		SELECT day_series.day AS period_start,
		       COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count
		FROM generate_series(?::date, ?::date, '1 day'::interval) AS day_series(day)
		LEFT JOIN meta_data ON meta_data.posted_at >= day_series.day 
		                    AND meta_data.posted_at < day_series.day + interval '1 day'
		LEFT JOIN job_post ON job_post.id = meta_data.job_post_id
		GROUP BY day_series.day
		ORDER BY day_series.day ASC
	`, fromDate, toDate).Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query daily vacancy trend: %w", err)
	}

	results := make([]models.VacancyTrendPoint, 0, len(rows))
	for _, row := range rows {
		results = append(results, models.VacancyTrendPoint{
			Label:        formatWeekLabel(row.PeriodStart),
			OpenJobCount: row.OpenJobCount,
		})
	}

	return results, nil
}

//This function returns job counts grouped month by month within the given date range
func (r *JobRepository) GetVacancyTrendMonthly(fromDate, toDate time.Time) ([]models.VacancyTrendPoint, error) {
	var rows []struct {
		PeriodStart  time.Time
		OpenJobCount int64
	}

	err := r.db.Raw(`
		SELECT month_series.month AS period_start,
		       COALESCE(SUM(job_post.no_of_vacancies), 0) AS open_job_count
		FROM generate_series(
		         date_trunc('month', ?::date),
		         date_trunc('month', ?::date),
		         '1 month'::interval
		     ) AS month_series(month)
		LEFT JOIN meta_data ON meta_data.posted_at >= month_series.month
		                    AND meta_data.posted_at < month_series.month + interval '1 month'
		                    AND meta_data.posted_at >= ?
		                    AND meta_data.posted_at <= ?
		LEFT JOIN job_post ON job_post.id = meta_data.job_post_id
		GROUP BY month_series.month
		ORDER BY month_series.month ASC
	`, fromDate, toDate, fromDate, toDate).Scan(&rows).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query monthly vacancy trend: %w", err)
	}

	results := make([]models.VacancyTrendPoint, 0, len(rows))
	for _, row := range rows {
		results = append(results, models.VacancyTrendPoint{
			Label:        row.PeriodStart.Format("Jan 2006"),
			OpenJobCount: row.OpenJobCount,
		})
	}

	n := len(rows)
	if n == 0 {
		return results, nil
	}

	firstBucketStart := rows[0].PeriodStart
	firstBucketNaturalEnd := firstBucketStart.AddDate(0, 1, 0).AddDate(0, 0, -1)
	lastBucketStart := rows[n-1].PeriodStart
	lastBucketNaturalEnd := lastBucketStart.AddDate(0, 1, 0).AddDate(0, 0, -1)

	firstIsPartial := fromDate.After(firstBucketStart)   
	lastIsPartial := toDate.Before(lastBucketNaturalEnd) 

	switch {
	case n == 1 && firstIsPartial && lastIsPartial:
		results[0].Label = fmt.Sprintf("%s - %s", fromDate.Format("Jan 2"), toDate.Format("Jan 2, 2006"))

	default:
		if firstIsPartial {
			results[0].Label = fmt.Sprintf("%s - %s", fromDate.Format("Jan 2"), firstBucketNaturalEnd.Format("Jan 2, 2006"))
		}
		if lastIsPartial {
			results[n-1].Label = fmt.Sprintf("%s - %s", lastBucketStart.Format("Jan 2"), toDate.Format("Jan 2, 2006"))
		}
	}

	return results, nil
}

func formatWeekLabel(t time.Time) string {
	return t.Format("Jan 2")
}

func (r *JobRepository) CreateCrawlerRun() (models.CrawlerRun, error) {
	now := time.Now()
	run := models.CrawlerRun{
		StartedAt: &now,
		Status:    "RUNNING",
	}

	err := r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		err := tx.Model(&models.CrawlerRun{}).
			Where("status = ?", "RUNNING").
			Updates(map[string]interface{}{
				"status":      "FAILED",
				"finished_at": &now,
			}).Error
		if err != nil {
			return err
		}

		if err := tx.Create(&run).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.CrawlerRun{}, err
	}

	return run, nil
}

func (r *JobRepository) CompleteCrawlerRun(id uint, status string) error {
	now := time.Now()
	return r.db.Model(&models.CrawlerRun{}).Where("id = ?", id).Updates(map[string]interface{}{
		"finished_at": &now,
		"status":      status, // 'COMPLETED' or 'FAILED'
	}).Error
}

func (r *JobRepository) GetJobsByBucketKeys(bucketKeys []string) ([]models.JobPost, error) {
	var jobs []models.JobPost

	err := r.db.Distinct("job_post.*").
		Joins("JOIN lsh_index ON lsh_index.job_post_id = job_post.id").
		Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
		Where("lsh_index.bucket_key IN ? AND meta_data.end_date IS NULL", bucketKeys).
		Preload("MetaData"). 
		Find(&jobs).Error

	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// SaveOneJob persists a single new job — employer/job_type/skills
// FirstOrCreate lookups, a geo lookup, source/ai_version FirstOrCreate,
// the job itself, then its LSH records keyed off the job's real DB id —
// in its own transaction, isolated from every other job in the batch.
func (r *JobRepository) SaveOneJob(job *models.JobPost, lshIndexRecords []models.LshIndex) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Employer (FirstOrCreate)
		if job.Employer != nil && job.Employer.Name != "" {
			var employer models.Employer
			if err := tx.Where(models.Employer{Name: job.Employer.Name}).
				FirstOrCreate(&employer).Error; err != nil {
				return wrapSaveErr(err, "employer lookup failed")
			}
			job.EmployerID = &employer.ID
			job.Employer = nil
		}

		// JobType (FirstOrCreate)
		if job.JobType != nil && job.JobType.Type != "" {
			var jt models.JobType
			if err := tx.Where(models.JobType{Type: job.JobType.Type}).
				FirstOrCreate(&jt).Error; err != nil {
				return wrapSaveErr(err, "job_type lookup failed")
			}
			job.JobTypeID = &jt.ID
			job.JobType = nil
		}

		// Skills (FirstOrCreate per skill)
		var linkedSkills []models.Skill
		for _, s := range job.Skills {
			if s.Skill == "" {
				continue
			}
			var skill models.Skill
			if err := tx.Where(models.Skill{Skill: s.Skill}).
				FirstOrCreate(&skill).Error; err != nil {
				return wrapSaveErr(err, fmt.Sprintf("skill '%s' lookup failed", s.Skill))
			}
			linkedSkills = append(linkedSkills, skill)
		}
		job.Skills = linkedSkills

		// Geo (lookup only, no create)
		if job.MetaData.GeoData != nil && job.MetaData.GeoData.Province != "" {
			var geo models.GeoData
			err := tx.Where("province = ?", job.MetaData.GeoData.Province).
				First(&geo).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: province '%s' not registered in geo_data",
					ErrPermanentSaveFailure, job.MetaData.GeoData.Province)
			} else if err != nil {
				return wrapSaveErr(err, "geo lookup failed")
			}
			job.MetaData.GeoDataID = &geo.ID
			job.MetaData.GeoData = nil
		}

		// Source (FirstOrCreate)
		if job.MetaData.Source != nil && job.MetaData.Source.Source != "" {
			var source models.Source
			if err := tx.Where(models.Source{Source: job.MetaData.Source.Source}).
				FirstOrCreate(&source).Error; err != nil {
				return wrapSaveErr(err, "source lookup failed")
			}
			job.MetaData.SourceID = &source.ID
			job.MetaData.Source = nil
		}

		// AiVersion (FirstOrCreate)
		if job.MetaData.AiVersion != nil && job.MetaData.AiVersion.Version != "" {
			var av models.AiVersion
			if err := tx.Where(models.AiVersion{Version: job.MetaData.AiVersion.Version}).
				FirstOrCreate(&av).Error; err != nil {
				return wrapSaveErr(err, "ai_version lookup failed")
			}
			job.MetaData.AiVersionID = &av.ID
			job.MetaData.AiVersion = nil
		}

		// Save JobPost (cascades MetaData + join table)
		if err := tx.Create(job).Error; err != nil {
			return wrapSaveErr(err, "job insert failed")
		}

		for i := range lshIndexRecords {
			lshIndexRecords[i].JobPostID = job.ID
		}
		if len(lshIndexRecords) > 0 {
			if err := tx.Omit("JobPost").Create(&lshIndexRecords).Error; err != nil {
				return wrapSaveErr(err, "lsh_index insert failed")
			}
		}

		return nil
	})
}

// UpdateDuplicateJob records that an already-saved job was seen again in
// crawlerRunID, in its own transaction.
func (r *JobRepository) UpdateDuplicateJob(jobPostID uint, crawlerRunID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("meta_data").
			Where("job_post_id = ?", jobPostID).
			Update("crawler_run_id", crawlerRunID).Error
	})
}

func (r *JobRepository) ReconcileStaleVacancies(currentRunID uint) (int64, error) {
	var affectedRows int64
	now := time.Now()

	err := r.db.Transaction(func(tx *gorm.DB) error {

		var staleJobIDs []uint
		err := tx.Model(&models.JobMetaData{}).
			Where("crawler_run_id <> ? AND end_date IS NULL", currentRunID).
			Pluck("job_post_id", &staleJobIDs).Error
		if err != nil {
			return err
		}

		if len(staleJobIDs) == 0 {
			return nil
		}

		if err := tx.Where("job_post_id IN ?", staleJobIDs).Delete(&models.LshIndex{}).Error; err != nil {
			return err
		}

		result := tx.Model(&models.JobMetaData{}).
			Where("job_post_id IN ?", staleJobIDs).
			Updates(map[string]interface{}{
				"end_date": &now,
			})
		if result.Error != nil {
			return result.Error
		}

		affectedRows = result.RowsAffected
		return nil
	})

	if err != nil {
		return 0, err
	}

	return affectedRows, nil
}

// Methods need to multi level filtering in the Crawler
// Occupation levels by parent ids
func (r *JobRepository) GetAllMajorGroupsForDateRange(fromDate, toDate time.Time) ([]models.MajorGroup, error) {
	var items []models.MajorGroup
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	err := r.db.
		Unscoped().
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at >= ?", fromDate).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetSubMajorGroupsByMajorGroup(majorGroupID uint) ([]models.SubMajorGroup, error) {
	var items []models.SubMajorGroup
	err := r.db.Where("major_group_id = ?", majorGroupID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetSubMajorGroupsByMajorGroupForDateRange(majorGroupID uint, fromDate, toDate time.Time) ([]models.SubMajorGroup, error) {
	var items []models.SubMajorGroup
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	err := r.db.
		Unscoped().
		Where("major_group_id = ?", majorGroupID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at >= ?", fromDate).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetMinorGroupsBySubMajorGroup(subMajorGroupID uint) ([]models.MinorGroup, error) {
	var items []models.MinorGroup
	err := r.db.Where("sub_major_group_id = ?", subMajorGroupID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetMinorGroupsBySubMajorGroupForDateRange(subMajorGroupID uint, fromDate, toDate time.Time) ([]models.MinorGroup, error) {
	var items []models.MinorGroup
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	err := r.db.
		Unscoped().
		Where("sub_major_group_id = ?", subMajorGroupID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at >= ?", fromDate).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetUnitGroupsByMinorGroup(minorGroupID uint) ([]models.UnitGroup, error) {
	var items []models.UnitGroup
	err := r.db.Where("minor_group_id = ?", minorGroupID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetUnitGroupsByMinorGroupForDateRange(minorGroupID uint, fromDate, toDate time.Time) ([]models.UnitGroup, error) {
	var items []models.UnitGroup
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	err := r.db.
		Unscoped().
		Where("minor_group_id = ?", minorGroupID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at >= ?", fromDate).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetOccupationGroupsByUnitGroup(unitGroupID uint) ([]models.OccupationGroup, error) {
	var items []models.OccupationGroup
	err := r.db.Where("unit_group_id = ?", unitGroupID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetOccupationGroupsByUnitGroupForDateRange(unitGroupID uint, fromDate, toDate time.Time) ([]models.OccupationGroup, error) {
	var items []models.OccupationGroup
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	err := r.db.
		Unscoped().
		Where("unit_group_id = ?", unitGroupID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at >= ?", fromDate).
		Find(&items).Error
	return items, err
}

// Industry levels by parent ids
func (r *JobRepository) GetAllIndustrySectorsForDateRange(fromDate, toDate time.Time) ([]models.IndustrySector, error) {
	var items []models.IndustrySector
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.db.
		Unscoped().
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at::date >= ?", fromDate.Format("2006-01-02")).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustryDivisionsByIndustrySector(industrySectorID uint) ([]models.IndustryDivision, error) {
	var items []models.IndustryDivision
	err := r.db.Where("industry_sector_id = ?", industrySectorID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustryDivisionsByIndustrySectorForDateRange(industrySectorID uint, fromDate, toDate time.Time) ([]models.IndustryDivision, error) {
	var items []models.IndustryDivision
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.db.
		Unscoped().
		Where("industry_sector_id = ?", industrySectorID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at::date >= ?", fromDate.Format("2006-01-02")).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustryGroupsByIndustryDivision(industryDivisionID uint) ([]models.IndustryGroup, error) {
	var items []models.IndustryGroup
	err := r.db.Where("industry_division_id = ?", industryDivisionID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustryGroupsByIndustryDivisionForDateRange(industryDivisionID uint, fromDate, toDate time.Time) ([]models.IndustryGroup, error) {
	var items []models.IndustryGroup
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.db.
		Unscoped().
		Where("industry_division_id = ?", industryDivisionID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at::date >= ?", fromDate.Format("2006-01-02")).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustryClassesByIndustryGroup(industryGroupID uint) ([]models.IndustryClass, error) {
	var items []models.IndustryClass
	err := r.db.Where("industry_group_id = ?", industryGroupID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustryClassesByIndustryGroupForDateRange(industryGroupID uint, fromDate, toDate time.Time) ([]models.IndustryClass, error) {
	var items []models.IndustryClass
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.db.
		Unscoped().
		Where("industry_group_id = ?", industryGroupID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at::date >= ?", fromDate.Format("2006-01-02")).
		Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustrySubclassesByIndustryClass(industryClassID uint) ([]models.IndustrySubclass, error) {
	var items []models.IndustrySubclass
	err := r.db.Where("industry_class_id = ?", industryClassID).Find(&items).Error
	return items, err
}

func (r *JobRepository) GetIndustrySubclassesByIndustryClassForDateRange(industryClassID uint, fromDate, toDate time.Time) ([]models.IndustrySubclass, error) {
	var items []models.IndustrySubclass
	endOfToDate := toDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.db.
		Unscoped().
		Where("industry_class_id = ?", industryClassID).
		Where("created_at <= ?", endOfToDate).
		Where("deleted_at IS NULL OR deleted_at::date >= ?", fromDate.Format("2006-01-02")).
		Find(&items).Error
	return items, err
}

// CRUD for database entities
// Geo data CRUD
func (r *JobRepository) CreateGeoData(item *models.GeoData) error {
	return r.db.Create(item).Error
}

func (r *JobRepository) GetAllGeoData() ([]models.GeoData, error) {
	var items []models.GeoData
	err := r.db.Find(&items).Error
	return items, err
}

func (r *JobRepository) GetGeoDataByID(id uint) (models.GeoData, error) {
	var item models.GeoData
	err := r.db.First(&item, id).Error
	return item, err
}

func (r *JobRepository) UpdateGeoData(id uint, updates map[string]interface{}) (models.GeoData, error) {
	var item models.GeoData
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}

func (r *JobRepository) DeleteGeoData(id uint) error {
	return r.db.Delete(&models.GeoData{}, id).Error
}

// Education level CRUD
func (r *JobRepository) CreateEducationLevel(item *models.EducationLevel) error {
	return r.db.Create(item).Error
}

func (r *JobRepository) GetAllEducationLevels() ([]models.EducationLevel, error) {
	var items []models.EducationLevel
	err := r.db.Find(&items).Error
	return items, err
}

func (r *JobRepository) GetEducationLevelByID(id uint) (models.EducationLevel, error) {
	var item models.EducationLevel
	err := r.db.First(&item, id).Error
	return item, err
}

func (r *JobRepository) UpdateEducationLevel(id uint, updates map[string]interface{}) (models.EducationLevel, error) {
	var item models.EducationLevel
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}

func (r *JobRepository) DeleteEducationLevel(id uint) error {
	return r.db.Delete(&models.EducationLevel{}, id).Error
}

// Formality CRUD
func (r *JobRepository) CreateFormality(item *models.Formality) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllFormalities() ([]models.Formality, error) {
	var items []models.Formality
	err := r.db.Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetFormalityByID(id uint) (models.Formality, error) {
	var item models.Formality
	err := r.db.First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateFormality(id uint, updates map[string]interface{}) (models.Formality, error) {
	var item models.Formality
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteFormality(id uint) error {
	return r.db.Delete(&models.Formality{}, id).Error
}

// Gender CRUD
func (r *JobRepository) CreateGender(item *models.Gender) error {
	return r.db.Create(item).Error
}

func (r *JobRepository) GetAllGenders() ([]models.Gender, error) {
	var items []models.Gender
	err := r.db.Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetGenderByID(id uint) (models.Gender, error) {
	var item models.Gender
	err := r.db.First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateGender(id uint, updates map[string]interface{}) (models.Gender, error) {
	var item models.Gender
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteGender(id uint) error {
	return r.db.Delete(&models.Gender{}, id).Error
}

// Employment Sector CRUD
func (r *JobRepository) CreateEmploymentSector(item *models.EmploymentSector) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllEmploymentSectors() ([]models.EmploymentSector, error) {
	var items []models.EmploymentSector
	err := r.db.Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetEmploymentSectorByID(id uint) (models.EmploymentSector, error) {
	var item models.EmploymentSector
	err := r.db.First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateEmploymentSector(id uint, updates map[string]interface{}) (models.EmploymentSector, error) {
	var item models.EmploymentSector
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteEmploymentSector(id uint) error {
	return r.db.Delete(&models.EmploymentSector{}, id).Error
}

// Vocational Education CRUD
func (r *JobRepository) CreateVocationalEducation(item *models.VocationalEducation) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllVocationalEducations() ([]models.VocationalEducation, error) {
	var items []models.VocationalEducation
	err := r.db.Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetVocationalEducationByID(id uint) (models.VocationalEducation, error) {
	var item models.VocationalEducation
	err := r.db.First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateVocationalEducation(id uint, updates map[string]interface{}) (models.VocationalEducation, error) {
	var item models.VocationalEducation
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteVocationalEducation(id uint) error {
	return r.db.Delete(&models.VocationalEducation{}, id).Error
}

// Experience CRUD
func (r *JobRepository) GetAllExperiences() ([]models.Experience, error) {
	var experiences []models.Experience
	if err := r.db.Find(&experiences).Error; err != nil {
		return nil, err
	}
	return experiences, nil
}

func (r *JobRepository) CreateExperience(item *models.Experience) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetExperienceByID(id uint) (models.Experience, error) {
	var item models.Experience
	err := r.db.First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateExperience(id uint, updates map[string]interface{}) (models.Experience, error) {
	var item models.Experience
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteExperience(id uint) error {
	return r.db.Delete(&models.Experience{}, id).Error
}

// Job Type CRUD
func (r *JobRepository) CreateJobType(item *models.JobType) error {
	return r.db.Create(item).Error
}

func (r *JobRepository) GetAllJobTypes() ([]models.JobType, error) {
	var jobTypes []models.JobType
	if err := r.db.Find(&jobTypes).Error; err != nil {
		return nil, err
	}
	return jobTypes, nil
}

func (r *JobRepository) GetJobTypeByID(id uint) (models.JobType, error) {
	var item models.JobType
	err := r.db.First(&item, id).Error
	return item, err
}

func (r *JobRepository) UpdateJobType(id uint, updates map[string]interface{}) (models.JobType, error) {
	var item models.JobType
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}

func (r *JobRepository) DeleteJobType(id uint) error {
	return r.db.Delete(&models.JobType{}, id).Error
}

// Major Group CRUD
func (r *JobRepository) CreateMajorGroup(item *models.MajorGroup) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllMajorGroups() ([]models.MajorGroup, error) {
	var items []models.MajorGroup
	err := r.db.Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetMajorGroupByID(id uint) (models.MajorGroup, error) {
	var item models.MajorGroup
	err := r.db.First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateMajorGroup(id uint, updates map[string]interface{}) (models.MajorGroup, error) {
	var item models.MajorGroup
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteMajorGroup(id uint) error {
	return r.db.Delete(&models.MajorGroup{}, id).Error
}

// Sub Major Group CRUD
func (r *JobRepository) CreateSubMajorGroup(item *models.SubMajorGroup) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllSubMajorGroups() ([]models.SubMajorGroup, error) {
	var items []models.SubMajorGroup
	err := r.db.Preload("MajorGroup").Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetSubMajorGroupByID(id uint) (models.SubMajorGroup, error) {
	var item models.SubMajorGroup
	err := r.db.Preload("MajorGroup").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateSubMajorGroup(id uint, updates map[string]interface{}) (models.SubMajorGroup, error) {
	var item models.SubMajorGroup
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteSubMajorGroup(id uint) error {
	return r.db.Delete(&models.SubMajorGroup{}, id).Error
}

// Minor Group CRUD
func (r *JobRepository) CreateMinorGroup(item *models.MinorGroup) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllMinorGroups() ([]models.MinorGroup, error) {
	var items []models.MinorGroup
	err := r.db.Preload("SubMajorGroup").Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetMinorGroupByID(id uint) (models.MinorGroup, error) {
	var item models.MinorGroup
	err := r.db.Preload("SubMajorGroup").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateMinorGroup(id uint, updates map[string]interface{}) (models.MinorGroup, error) {
	var item models.MinorGroup
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteMinorGroup(id uint) error {
	return r.db.Delete(&models.MinorGroup{}, id).Error
}

// Unit Group CRUD
func (r *JobRepository) CreateUnitGroup(item *models.UnitGroup) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllUnitGroups() ([]models.UnitGroup, error) {
	var items []models.UnitGroup
	err := r.db.Preload("MinorGroup").Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetUnitGroupByID(id uint) (models.UnitGroup, error) {
	var item models.UnitGroup
	err := r.db.Preload("MinorGroup").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateUnitGroup(id uint, updates map[string]interface{}) (models.UnitGroup, error) {
	var item models.UnitGroup
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteUnitGroup(id uint) error {
	return r.db.Delete(&models.UnitGroup{}, id).Error
}

// Ocation Group CRUD
func (r *JobRepository) CreateOccupationGroup(item *models.OccupationGroup) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllOccupationGroups(limit, offset int) ([]models.OccupationGroup, int64, error) {
	var items []models.OccupationGroup
	var total int64

	if err := r.db.Model(&models.OccupationGroup{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Preload("UnitGroup")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("id ASC").Find(&items).Error

	return items, total, err
}
 
func (r *JobRepository) GetOccupationGroupByID(id uint) (models.OccupationGroup, error) {
	var item models.OccupationGroup
	err := r.db.Preload("UnitGroup").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateOccupationGroup(id uint, updates map[string]interface{}) (models.OccupationGroup, error) {
	var item models.OccupationGroup
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteOccupationGroup(id uint) error {
	return r.db.Delete(&models.OccupationGroup{}, id).Error
}

// Industry sector CRUD
func (r *JobRepository) CreateIndustrySector(item *models.IndustrySector) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllIndustrySectors() ([]models.IndustrySector, error) {
	var items []models.IndustrySector
	err := r.db.Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetIndustrySectorByID(id uint) (models.IndustrySector, error) {
	var item models.IndustrySector
	err := r.db.First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateIndustrySector(id uint, updates map[string]interface{}) (models.IndustrySector, error) {
	var item models.IndustrySector
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteIndustrySector(id uint) error {
	return r.db.Delete(&models.IndustrySector{}, id).Error
}

// Industry division CRUD
func (r *JobRepository) CreateIndustryDivision(item *models.IndustryDivision) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllIndustryDivisions() ([]models.IndustryDivision, error) {
	var items []models.IndustryDivision
	err := r.db.Preload("IndustrySector").Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetIndustryDivisionByID(id uint) (models.IndustryDivision, error) {
	var item models.IndustryDivision
	err := r.db.Preload("IndustrySector").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateIndustryDivision(id uint, updates map[string]interface{}) (models.IndustryDivision, error) {
	var item models.IndustryDivision
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteIndustryDivision(id uint) error {
	return r.db.Delete(&models.IndustryDivision{}, id).Error
}

// Industry group CRUD
func (r *JobRepository) CreateIndustryGroup(item *models.IndustryGroup) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllIndustryGroups() ([]models.IndustryGroup, error) {
	var items []models.IndustryGroup
	err := r.db.Preload("IndustryDivision").Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetIndustryGroupByID(id uint) (models.IndustryGroup, error) {
	var item models.IndustryGroup
	err := r.db.Preload("IndustryDivision").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateIndustryGroup(id uint, updates map[string]interface{}) (models.IndustryGroup, error) {
	var item models.IndustryGroup
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteIndustryGroup(id uint) error {
	return r.db.Delete(&models.IndustryGroup{}, id).Error
}

// Industry class CRUD
func (r *JobRepository) CreateIndustryClass(item *models.IndustryClass) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllIndustryClasses() ([]models.IndustryClass, error) {
	var items []models.IndustryClass
	err := r.db.Preload("IndustryGroup").Find(&items).Error
	return items, err
}
 
func (r *JobRepository) GetIndustryClassByID(id uint) (models.IndustryClass, error) {
	var item models.IndustryClass
	err := r.db.Preload("IndustryGroup").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateIndustryClass(id uint, updates map[string]interface{}) (models.IndustryClass, error) {
	var item models.IndustryClass
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteIndustryClass(id uint) error {
	return r.db.Delete(&models.IndustryClass{}, id).Error
}

// Industry sub class CRUD
func (r *JobRepository) CreateIndustrySubclass(item *models.IndustrySubclass) error {
	return r.db.Create(item).Error
}
 
func (r *JobRepository) GetAllIndustrySubclasses(limit, offset int) ([]models.IndustrySubclass, int64, error) {
	var items []models.IndustrySubclass
	var total int64

	if err := r.db.Model(&models.IndustrySubclass{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.db.Preload("IndustryClass")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("id ASC").Find(&items).Error
	return items, total, err
}
 
func (r *JobRepository) GetIndustrySubclassByID(id uint) (models.IndustrySubclass, error) {
	var item models.IndustrySubclass
	err := r.db.Preload("IndustryClass").First(&item, id).Error
	return item, err
}
 
func (r *JobRepository) UpdateIndustrySubclass(id uint, updates map[string]interface{}) (models.IndustrySubclass, error) {
	var item models.IndustrySubclass
	if err := r.db.First(&item, id).Error; err != nil {
		return item, err
	}
	if err := r.db.Model(&item).Updates(updates).Error; err != nil {
		return item, err
	}
	return item, nil
}
 
func (r *JobRepository) DeleteIndustrySubclass(id uint) error {
	return r.db.Delete(&models.IndustrySubclass{}, id).Error
}