package models

import "time"

// Yearly summary tables, written by the retention job when a calendar year
// leaves the raw-data window. See the comment above stats_year_total in
// sql/schema.sql for what each dimension's value_id means.

// Year statuses, in the order a year goes through them:
//   - rolled_up: the summary is written and checked; no raw row is deleted yet,
//     so the roll-up may be repeated.
//   - purging: the deletes have started and some raw rows may already be gone.
//     The summary is final: rolling up again would replace it with a total
//     counted from the rows that remain.
//   - purged: every raw row of the year is deleted.
const (
	YearStatusRolledUp = "rolled_up"
	YearStatusPurging  = "purging"
	YearStatusPurged   = "purged"
)

// Dimensions of the yearly summaries.
const (
	DimensionTotal               = "total"
	DimensionGender              = "gender"
	DimensionEducationLevel      = "education_level"
	DimensionExperience          = "experience"
	DimensionFormality           = "formality"
	DimensionVocationalEducation = "vocational_education"
	DimensionEmploymentSector    = "employment_sector"
	DimensionProvince            = "province"
	DimensionJobType             = "job_type"
	DimensionWorkMode            = "work_mode"
)

// UnknownValueID is the value_id of jobs that have no value for a dimension,
// and of the 'total' dimension.
const UnknownValueID = 0

// WorkModeCode pairs a work mode with the code stored in value_id.
type WorkModeCode struct {
	Mode WorkMode
	Code int
}

// WorkModeCodes is the one place the work mode codes are defined; the
// roll-up SQL is built from it. work_mode_enum is a Postgres enum, so it has
// no id of its own. Stored history depends on these numbers: never renumber
// or reuse one, only append.
var WorkModeCodes = []WorkModeCode{
	{WorkModeRemote, 1},
	{WorkModeOnsite, 2},
	{WorkModeHybrid, 3},
}

// Code returns the stored code of a work mode, or false for an unknown one.
func (m WorkMode) Code() (int, bool) {
	for _, c := range WorkModeCodes {
		if c.Mode == m {
			return c.Code, true
		}
	}
	return 0, false
}

type StatsYearTotal struct {
	Year         int       `gorm:"primaryKey;autoIncrement:false;not null"`
	VacancyCount int64     `gorm:"not null"`
	Status       string    `gorm:"type:year_status_enum;not null"`
	RolledUpAt   time.Time `gorm:"not null"`
	PurgedAt     *time.Time
}

func (StatsYearTotal) TableName() string { return "stats_year_total" }

// StatsYearOccupation is one summary row per major group. The field order is
// the primary key order.
type StatsYearOccupation struct {
	RootID       uint   `gorm:"primaryKey;autoIncrement:false;not null"`
	Dimension    string `gorm:"primaryKey;size:30;not null"`
	Year         int    `gorm:"primaryKey;autoIncrement:false;not null"`
	ValueID      int    `gorm:"primaryKey;autoIncrement:false;not null"`
	VacancyCount int64  `gorm:"not null"`
}

func (StatsYearOccupation) TableName() string { return "stats_year_occupation" }

// StatsYearIndustry is one summary row per industry sector.
type StatsYearIndustry struct {
	RootID       uint   `gorm:"primaryKey;autoIncrement:false;not null"`
	Dimension    string `gorm:"primaryKey;size:30;not null"`
	Year         int    `gorm:"primaryKey;autoIncrement:false;not null"`
	ValueID      int    `gorm:"primaryKey;autoIncrement:false;not null"`
	VacancyCount int64  `gorm:"not null"`
}

func (StatsYearIndustry) TableName() string { return "stats_year_industry" }
