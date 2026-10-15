package models

// API response DTOs for stats/count endpoints

type LevelChildJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	OpenJobCount int64  `json:"open_job_count"`
}

type TotalVacancyCount struct {
	FromDate       string `json:"from_date"`
	ToDate         string `json:"to_date"`
	TotalVacancies int64  `json:"total_vacancies"`
}

type VacancyTrendPoint struct {
	Label        string `json:"label"`
	OpenJobCount int64  `json:"open_job_count"`
}

type SkillDemand struct {
	ID           uint   `json:"id"`
	Skill        string `json:"skill"`
	OpenJobCount int64  `json:"open_job_count"`
}

type EmployerDemand struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type OccupationJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type IndustryJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type ExperienceJobCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type EducationLevelJobCount struct {
	ID           uint   `json:"id"`
	Level        string `json:"level"`
	OpenJobCount int64  `json:"open_job_count"`
}

type RemoteOnSiteHybridCount struct {
	RemoteCount int64 `json:"remote_count"`
	OnSiteCount int64 `json:"on_site_count"`
	HybridCount int64 `json:"hybrid_count"`
}

type JobTypeJobCount struct {
	ID           uint   `json:"id"`
	Type         string `json:"type"`
	OpenJobCount int64  `json:"open_job_count"`
}

type TopJobRole struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	OpenJobCount int64  `json:"open_job_count"`
}

type ProvinceJobCount struct {
	ID           uint    `json:"id"`
	Province     string  `json:"province"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
	OpenJobCount int64   `json:"open_job_count"`
}

type FormalityJobCount struct {
	ID            uint   `json:"id"`
	FormalityType string `json:"formality_type"`
	OpenJobCount  int64  `json:"open_job_count"`
}

type EmploymentSectorJobCount struct {
	ID           uint   `json:"id"`
	Sector       string `json:"sector"`
	OpenJobCount int64  `json:"open_job_count"`
}

type GenderJobCount struct {
	ID           uint   `json:"id"`
	GenderType   string `json:"gender_type"`
	OpenJobCount int64  `json:"open_job_count"`
}

type VocationalEducationJobCount struct {
	ID           uint   `json:"id"`
	Level        string `json:"level"`
	OpenJobCount int64  `json:"open_job_count"`
}
