package api

import (
	"marketlens-go-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type topHiringEmployersResponse struct {
	Level     string                  `json:"level"`
	ID        uint                    `json:"id"`
	FromDate  string                  `json:"from_date"`
	ToDate    string                  `json:"to_date"`
	Count     int                     `json:"count"`
	Employers []models.EmployerDemand `json:"employers"`
}

// This function returns the top hiring employers for a given occupation hierarchy
// level and id, filtered by date range.
func (h *Handler) GetTopHiringEmployersByOccupationLevelHandler(c *gin.Context) {
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetTopHiringEmployersByOccupationLevel(c.Request.Context(), level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve top hiring employers", err)
		return
	}

	c.JSON(http.StatusOK, topHiringEmployersResponse{Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Employers: results})
}

type skillsPaginatedResponse struct {
	Level    string               `json:"level"`
	ID       uint                 `json:"id"`
	FromDate string               `json:"from_date"`
	ToDate   string               `json:"to_date"`
	Count    int                  `json:"count"`
	Total    int64                `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
	Skills   []models.SkillDemand `json:"skills"`
}

// This function returns all skills (paginated) for a given occupation hierarchy
// level and id, filtered by date range.
func (h *Handler) GetAllSkillsByOccupationLevelHandler(c *gin.Context) {
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			if v > 100 {
				v = 100
			}
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	results, total, err := h.Analytics.GetAllSkillsByOccupationLevel(c.Request.Context(), level, id, dr.From, dr.To, limit, offset)
	if err != nil {
		respondRepoError(c, "Failed to retrieve skills", err)
		return
	}

	c.JSON(http.StatusOK, skillsPaginatedResponse{Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Total: total, Limit: limit, Offset: offset, Skills: results})
}

type top15SkillsResponse struct {
	Level    string               `json:"level"`
	ID       uint                 `json:"id"`
	FromDate string               `json:"from_date"`
	ToDate   string               `json:"to_date"`
	Count    int                  `json:"count"`
	Skills   []models.SkillDemand `json:"skills"`
}

// This function returns the top 15 skills for a given occupation hierarchy
// level and id, filtered by date range.
func (h *Handler) GetTop15SkillsByOccupationLevelHandler(c *gin.Context) {
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetTop15SkillsByOccupationLevel(c.Request.Context(), level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve top 15 skills", err)
		return
	}

	c.JSON(http.StatusOK, top15SkillsResponse{Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Skills: results})
}

type jobTypeByLevelResponse struct {
	Standard string                     `json:"standard"`
	Level    string                     `json:"level"`
	ID       uint                       `json:"id"`
	FromDate string                     `json:"from_date"`
	ToDate   string                     `json:"to_date"`
	Count    int                        `json:"count"`
	JobTypes []models.JobTypeJobCount `json:"job_types"`
}

// This function returns job type breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetJobTypeByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetJobTypeByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve job type breakdown", err)
		return
	}

	c.JSON(http.StatusOK, jobTypeByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), JobTypes: results})
}

type remoteOnSiteHybridResponse struct {
	Standard                string                          `json:"standard"`
	Level                   string                          `json:"level"`
	ID                      uint                            `json:"id"`
	FromDate                string                          `json:"from_date"`
	ToDate                  string                          `json:"to_date"`
	RemoteVsOnsiteVsHybrid  models.RemoteOnSiteHybridCount `json:"remote_vs_onsite_vs_hybrid"`
}

// This function returns remote vs on-site job counts for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetRemoteOnSiteHybridByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	result, err := h.Analytics.GetRemoteOnSiteHybridByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve remote/on-site breakdown", err)
		return
	}

	c.JSON(http.StatusOK, remoteOnSiteHybridResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, RemoteVsOnsiteVsHybrid: result})
}

type vocationalEducationByLevelResponse struct {
	Standard              string                                `json:"standard"`
	Level                 string                                `json:"level"`
	ID                    uint                                  `json:"id"`
	FromDate              string                                `json:"from_date"`
	ToDate                string                                `json:"to_date"`
	Count                 int                                   `json:"count"`
	VocationalEducations []models.VocationalEducationJobCount `json:"vocational_educations"`
}

// This function returns vocational education breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetVocationalEducationByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetVocationalEducationByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve vocational education breakdown", err)
		return
	}

	c.JSON(http.StatusOK, vocationalEducationByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), VocationalEducations: results})
}

type genderByLevelResponse struct {
	Standard string                   `json:"standard"`
	Level    string                   `json:"level"`
	ID       uint                     `json:"id"`
	FromDate string                   `json:"from_date"`
	ToDate   string                   `json:"to_date"`
	Count    int                      `json:"count"`
	Genders  []models.GenderJobCount `json:"genders"`
}

// This function returns gender breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetGenderByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetGenderByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve gender breakdown", err)
		return
	}

	c.JSON(http.StatusOK, genderByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Genders: results})
}

type formalityByLevelResponse struct {
	Standard    string                      `json:"standard"`
	Level       string                      `json:"level"`
	ID          uint                        `json:"id"`
	FromDate    string                      `json:"from_date"`
	ToDate      string                      `json:"to_date"`
	Count       int                         `json:"count"`
	Formalities []models.FormalityJobCount `json:"formalities"`
}

// This function returns formality breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetFormalityByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetFormalityByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve formality breakdown", err)
		return
	}

	c.JSON(http.StatusOK, formalityByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Formalities: results})
}

type educationLevelByLevelResponse struct {
	Standard         string                           `json:"standard"`
	Level            string                           `json:"level"`
	ID               uint                             `json:"id"`
	FromDate         string                           `json:"from_date"`
	ToDate           string                           `json:"to_date"`
	Count            int                              `json:"count"`
	EducationLevels []models.EducationLevelJobCount `json:"education_levels"`
}

// This function returns education level breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetEducationLevelByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetEducationLevelByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve education level breakdown", err)
		return
	}

	c.JSON(http.StatusOK, educationLevelByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), EducationLevels: results})
}

type provinceByLevelResponse struct {
	Standard  string                     `json:"standard"`
	Level     string                     `json:"level"`
	ID        uint                       `json:"id"`
	FromDate  string                     `json:"from_date"`
	ToDate    string                     `json:"to_date"`
	Count     int                        `json:"count"`
	Provinces []models.ProvinceJobCount `json:"provinces"`
}

// This function returns province breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetProvinceByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetProvinceByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve province breakdown", err)
		return
	}

	c.JSON(http.StatusOK, provinceByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Provinces: results})
}

type experienceByLevelResponse struct {
	Standard    string                       `json:"standard"`
	Level       string                       `json:"level"`
	ID          uint                         `json:"id"`
	FromDate    string                       `json:"from_date"`
	ToDate      string                       `json:"to_date"`
	Count       int                          `json:"count"`
	Experiences []models.ExperienceJobCount `json:"experiences"`
}

// This function returns experience-level breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetExperienceByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetExperienceByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve experience breakdown", err)
		return
	}

	c.JSON(http.StatusOK, experienceByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Experiences: results})
}

type employmentSectorByLevelResponse struct {
	Standard           string                               `json:"standard"`
	Level              string                               `json:"level"`
	ID                 uint                                 `json:"id"`
	FromDate           string                               `json:"from_date"`
	ToDate             string                               `json:"to_date"`
	Count              int                                  `json:"count"`
	EmploymentSectors []models.EmploymentSectorJobCount `json:"employment_sectors"`
}

// This function returns employment sector breakdown for a given occupation/industry
// hierarchy level and id, filtered by date range.
func (h *Handler) GetEmploymentSectorByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetEmploymentSectorByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve employment sector breakdown", err)
		return
	}

	c.JSON(http.StatusOK, employmentSectorByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), EmploymentSectors: results})
}

type levelChildrenResponse struct {
	Standard   string                        `json:"standard"`
	Level      string                        `json:"level"`
	ID         uint                          `json:"id"`
	ChildLevel string                        `json:"child_level"`
	FromDate   string                        `json:"from_date"`
	ToDate     string                        `json:"to_date"`
	Count      int                           `json:"count"`
	Children   []models.LevelChildJobCount `json:"children"`
}

// This function returns the immediate children of a given occupation/industry hierarchy
// level and id, each with its aggregated job count for a given date range.
func (h *Handler) GetLevelChildrenHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, childLevel, err := h.Analytics.GetLevelChildren(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve child level data", err)
		return
	}

	c.JSON(http.StatusOK, levelChildrenResponse{Standard: standard, Level: level, ID: id, ChildLevel: childLevel, FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Children: results})
}

type totalJobCountByLevelResponse struct {
	Standard      string `json:"standard"`
	Level         string `json:"level"`
	ID            uint   `json:"id"`
	FromDate      string `json:"from_date"`
	ToDate        string `json:"to_date"`
	TotalJobCount int64  `json:"total_job_count"`
}

// This function returns the total job count for a given occupation/industry hierarchy level and id,
// filtered by an optional date range.
func (h *Handler) GetTotalJobCountByLevelHandler(c *gin.Context) {
	standard := c.Param("standard")
	level := c.Param("level")
	id, ok := parseIDParamMsg(c, "invalid id parameter, must be a positive integer")
	if !ok {
		return
	}
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	total, err := h.Analytics.GetTotalJobCountByLevel(c.Request.Context(), standard, level, id, dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve total job count", err)
		return
	}

	c.JSON(http.StatusOK, totalJobCountByLevelResponse{Standard: standard, Level: level, ID: id, FromDate: dr.FromStr, ToDate: dr.ToStr, TotalJobCount: total})
}

type industryJobCountResponse struct {
	FromDate   string                     `json:"from_date"`
	ToDate     string                     `json:"to_date"`
	Count      int                        `json:"count"`
	Industries []models.IndustryJobCount `json:"industries"`
}

// This function returns vacancy counts grouped by industry sector, for a given date range
func (h *Handler) GetIndustryJobCountByDateRangeHandler(c *gin.Context) {
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetIndustryJobCountByDateRange(c.Request.Context(), dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve industry job counts", err)
		return
	}

	c.JSON(http.StatusOK, industryJobCountResponse{FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Industries: results})
}

type occupationJobCountResponse struct {
	FromDate    string                       `json:"from_date"`
	ToDate      string                       `json:"to_date"`
	Count       int                          `json:"count"`
	Occupations []models.OccupationJobCount `json:"occupations"`
}

// This function returns vacancy counts grouped by major group (occupation), for a given date range
func (h *Handler) GetOccupationJobCountByDateRangeHandler(c *gin.Context) {
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, err := h.Analytics.GetOccupationJobCountByDateRange(c.Request.Context(), dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve occupation job counts", err)
		return
	}

	c.JSON(http.StatusOK, occupationJobCountResponse{FromDate: dr.FromStr, ToDate: dr.ToStr, Count: len(results), Occupations: results})
}

// This function returns the total number of vacancies posted within a given date range
func (h *Handler) GetTotalVacancyCountHandler(c *gin.Context) {
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	total, err := h.Analytics.GetTotalVacancyCount(c.Request.Context(), dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve total vacancy count", err)
		return
	}

	c.JSON(http.StatusOK, models.TotalVacancyCount{
		FromDate:       dr.FromStr,
		ToDate:         dr.ToStr,
		TotalVacancies: total,
	})
}

type vacancyTrendResponse struct {
	FromDate    string                        `json:"from_date"`
	ToDate      string                        `json:"to_date"`
	Granularity string                        `json:"granularity"`
	Count       int                           `json:"count"`
	Data        []models.VacancyTrendPoint `json:"data"`
}

// This function returns vacancy counts over a date range with granularity that adapts
// to the range's length: daily under 60 days, monthly otherwise
func (h *Handler) GetVacancyTrendHandler(c *gin.Context) {
	dr, ok := parseDateRangeQuery(c)
	if !ok {
		return
	}

	results, granularity, err := h.Analytics.GetVacancyTrend(c.Request.Context(), dr.From, dr.To)
	if err != nil {
		respondRepoError(c, "Failed to retrieve vacancy trend data", err)
		return
	}

	c.JSON(http.StatusOK, vacancyTrendResponse{FromDate: dr.FromStr, ToDate: dr.ToStr, Granularity: granularity, Count: len(results), Data: results})
}
