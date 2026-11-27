package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"marketlens-go-backend/internal/models"
)

// retentionLockKey is the Postgres advisory lock key that keeps two retention
// runs from overlapping ("MLRE" in ASCII).
const retentionLockKey int64 = 0x4d4c5245

// ErrYearAlreadyPurged is returned when a roll-up is asked for a year whose
// raw rows are already deleted. Rolling it up again would overwrite the
// summary with zeros.
var ErrYearAlreadyPurged = errors.New("year already purged")

// rollUpDimension is one breakdown of the yearly summary: its name and the
// SQL expression, over meta_data (m) and job_post (jp), that gives a job's
// value_id.
type rollUpDimension struct {
	name string
	expr string
}

// rollUpDimensions lists every dimension written for a year. A missing value
// is stored as value_id 0 so each dimension adds up to the total.
var rollUpDimensions = []rollUpDimension{
	{models.DimensionTotal, "0"},
	{models.DimensionGender, "COALESCE(m.gender_id, 0)"},
	{models.DimensionEducationLevel, "COALESCE(m.education_level_id, 0)"},
	{models.DimensionExperience, "COALESCE(m.experience_id, 0)"},
	{models.DimensionFormality, "COALESCE(m.formality_id, 0)"},
	{models.DimensionVocationalEducation, "COALESCE(m.vocational_education_id, 0)"},
	{models.DimensionEmploymentSector, "COALESCE(m.employment_sector_id, 0)"},
	{models.DimensionProvince, "COALESCE(m.geo_data_id, 0)"},
	{models.DimensionJobType, "COALESCE(jp.job_type_id, 0)"},
	{models.DimensionWorkMode, workModeCaseExpr()},
}

// workModeCaseExpr builds the CASE that turns job_post.work_mode into its
// code from models.WorkModeCodes. It has no ELSE on purpose: a work mode
// that has no code gives NULL, the NOT NULL value_id column rejects it, and
// the roll-up fails loudly instead of filing those jobs under a wrong code.
func workModeCaseExpr() string {
	var b strings.Builder
	b.WriteString("CASE jp.work_mode")
	for _, c := range models.WorkModeCodes {
		fmt.Fprintf(&b, " WHEN '%s' THEN %d", c.Mode, c.Code)
	}
	b.WriteString(" END")
	return b.String()
}

// rollUpStandard describes one hierarchy: the summary table it writes, the
// meta_data column that holds the job's leaf node, and a query mapping every
// leaf node to its root.
type rollUpStandard struct {
	table      string
	leafColumn string
	leafToRoot string
}

var rollUpStandards = []rollUpStandard{
	{
		table:      "stats_year_occupation",
		leafColumn: "occupation_group_id",
		leafToRoot: `SELECT og.id AS leaf_id, smg.major_group_id AS root_id
			FROM occupation_group og
			JOIN unit_group ug ON ug.id = og.unit_group_id
			JOIN minor_group mg ON mg.id = ug.minor_group_id
			JOIN sub_major_group smg ON smg.id = mg.sub_major_group_id`,
	},
	{
		table:      "stats_year_industry",
		leafColumn: "industry_subclass_id",
		leafToRoot: `SELECT isc.id AS leaf_id, idv.industry_sector_id AS root_id
			FROM industry_subclass isc
			JOIN industry_class ic ON ic.id = isc.industry_class_id
			JOIN industry_group ig ON ig.id = ic.industry_group_id
			JOIN industry_division idv ON idv.id = ig.industry_division_id`,
	},
}

// rollUpInsertSQL builds the INSERT ... SELECT that writes one dimension of
// one year for one standard. The year is formatted in as an integer, not
// bound, because Postgres cannot infer the type of a bound parameter in the
// select list of an INSERT. The date range is [from, to).
func rollUpInsertSQL(std rollUpStandard, dim rollUpDimension, year int) string {
	groupBy := "lr.root_id"
	if dim.name != models.DimensionTotal {
		// "GROUP BY 0" would be read as a column position, so the constant
		// 'total' dimension is not grouped on.
		groupBy += ", " + dim.expr
	}
	return fmt.Sprintf(`INSERT INTO %s (root_id, dimension, year, value_id, vacancy_count)
		SELECT lr.root_id, '%s', %d, %s, SUM(jp.no_of_vacancies)
		FROM meta_data m
		JOIN job_post jp ON jp.id = m.job_post_id
		JOIN (%s) lr ON lr.leaf_id = m.%s
		WHERE m.posted_at >= ? AND m.posted_at < ?
		GROUP BY %s`,
		std.table, dim.name, year, dim.expr, std.leafToRoot, std.leafColumn, groupBy)
}

// GetYearTotal returns the summary row of a year, or nil if the year has not
// been rolled up.
func (r *Store) GetYearTotal(ctx context.Context, year int) (*models.StatsYearTotal, error) {
	var rows []models.StatsYearTotal
	if err := r.db.WithContext(ctx).Where("year = ?", year).Limit(1).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to load the summary of year %d: %w", year, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}

// EarliestPostedAt returns the oldest posted_at among the raw jobs, or nil if
// there are none.
func (r *Store) EarliestPostedAt(ctx context.Context) (*time.Time, error) {
	var posted []time.Time
	err := r.db.WithContext(ctx).Model(&models.JobMetaData{}).
		Order("posted_at ASC").Limit(1).
		Pluck("posted_at", &posted).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find the earliest posted_at: %w", err)
	}
	if len(posted) == 0 {
		return nil, nil
	}
	return &posted[0], nil
}

// RawVacancyTotal sums the vacancies of the raw jobs posted in [from, to),
// classified or not.
func (r *Store) RawVacancyTotal(ctx context.Context, from, to time.Time) (int64, error) {
	return rawVacancyTotal(r.db.WithContext(ctx), from, to)
}

func rawVacancyTotal(db *gorm.DB, from, to time.Time) (int64, error) {
	var total int64
	err := db.Table("job_post").
		Select("COALESCE(SUM(job_post.no_of_vacancies), 0)").
		Joins("JOIN meta_data ON meta_data.job_post_id = job_post.id").
		Where("meta_data.posted_at >= ? AND meta_data.posted_at < ?", from, to).
		Scan(&total).Error
	if err != nil {
		return 0, fmt.Errorf("failed to sum the raw vacancies: %w", err)
	}
	return total, nil
}

// RollUpYear writes the yearly summary of the raw jobs posted in [from, to).
// It runs in one transaction and replaces whatever was stored for the year,
// so running it again is safe and also picks up jobs that arrived late. It
// refuses a purged year, whose raw rows are gone.
func (r *Store) RollUpYear(ctx context.Context, year int, from, to time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing []models.StatsYearTotal
		if err := tx.Where("year = ?", year).Limit(1).Find(&existing).Error; err != nil {
			return fmt.Errorf("failed to load the summary of year %d: %w", year, err)
		}
		if len(existing) > 0 && existing[0].Status == models.YearStatusPurged {
			return fmt.Errorf("roll up year %d: %w", year, ErrYearAlreadyPurged)
		}

		total, err := rawVacancyTotal(tx, from, to)
		if err != nil {
			return err
		}

		// The total row goes first: the other tables reference it.
		row := models.StatsYearTotal{
			Year:         year,
			VacancyCount: total,
			Status:       models.YearStatusRolledUp,
			RolledUpAt:   time.Now(),
		}
		err = tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "year"}},
			DoUpdates: clause.AssignmentColumns([]string{"vacancy_count", "status", "rolled_up_at"}),
		}).Create(&row).Error
		if err != nil {
			return fmt.Errorf("failed to write the total of year %d: %w", year, err)
		}

		for _, std := range rollUpStandards {
			if err := tx.Exec("DELETE FROM "+std.table+" WHERE year = ?", year).Error; err != nil {
				return fmt.Errorf("failed to clear %s for year %d: %w", std.table, year, err)
			}
			for _, dim := range rollUpDimensions {
				if err := tx.Exec(rollUpInsertSQL(std, dim, year), from, to).Error; err != nil {
					return fmt.Errorf("failed to roll up %s/%s for year %d: %w", std.table, dim.name, year, err)
				}
			}
		}
		return nil
	})
}

// VerifyYearSummary checks that every dimension of every root adds up to the
// same figure as the 'total' dimension, in both summary tables. It catches a
// dimension the roll-up missed or double counted, before any raw row is
// deleted.
func (r *Store) VerifyYearSummary(ctx context.Context, year int) error {
	for _, std := range rollUpStandards {
		var rows []struct {
			Dimension string
			Total     int64
		}
		err := r.db.WithContext(ctx).Raw(
			"SELECT dimension, SUM(vacancy_count) AS total FROM "+std.table+" WHERE year = ? GROUP BY dimension",
			year,
		).Scan(&rows).Error
		if err != nil {
			return fmt.Errorf("failed to read %s for year %d: %w", std.table, year, err)
		}

		sums := make(map[string]int64, len(rows))
		for _, row := range rows {
			sums[row.Dimension] = row.Total
		}
		want := sums[models.DimensionTotal]
		for _, dim := range rollUpDimensions {
			if want == 0 && sums[dim.name] == 0 {
				continue // an empty year has no rows at all
			}
			if got, ok := sums[dim.name]; !ok || got != want {
				return fmt.Errorf("year %d: %s/%s adds up to %d, want %d (the total)", year, std.table, dim.name, got, want)
			}
		}
	}
	return nil
}

// PurgeBatch deletes up to limit raw jobs posted in [from, to) and returns how
// many it deleted; 0 means none are left. The foreign keys cascade to
// meta_data, lsh_index and job_post_skills. Each call is its own transaction,
// so a stopped run loses at most one batch and the next run repeats the same
// delete.
func (r *Store) PurgeBatch(ctx context.Context, from, to time.Time, limit int) (int64, error) {
	res := r.db.WithContext(ctx).Exec(
		`DELETE FROM job_post WHERE id IN (
			SELECT job_post_id FROM meta_data WHERE posted_at >= ? AND posted_at < ? LIMIT ?)`,
		from, to, limit,
	)
	if res.Error != nil {
		return 0, fmt.Errorf("failed to purge a batch of jobs: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// MarkYearPurged records that the raw rows of a rolled-up year are gone.
func (r *Store) MarkYearPurged(ctx context.Context, year int) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&models.StatsYearTotal{}).
		Where("year = ? AND status = ?", year, models.YearStatusRolledUp).
		Updates(map[string]any{"status": models.YearStatusPurged, "purged_at": &now})
	if res.Error != nil {
		return fmt.Errorf("failed to mark year %d purged: %w", year, res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("year %d has no rolled-up summary to mark purged", year)
	}
	return nil
}

// TryRetentionLock takes the retention advisory lock without waiting. If
// another run holds it, acquired is false. Otherwise the caller must call
// release when it is done. The lock belongs to one database connection that is
// held until release, and Postgres drops it if the process dies, so a killed
// pod never leaves it stuck.
func (r *Store) TryRetentionLock(ctx context.Context) (release func(), acquired bool, err error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get the database handle: %w", err)
	}
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to reserve a connection for the lock: %w", err)
	}
	var got bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", retentionLockKey).Scan(&got); err != nil {
		conn.Close()
		return nil, false, fmt.Errorf("failed to take the retention lock: %w", err)
	}
	if !got {
		conn.Close()
		return nil, false, nil
	}
	return func() { releaseLock(conn) }, true, nil
}

func releaseLock(conn *sql.Conn) {
	// The caller's context may be cancelled already (SIGTERM), so unlock with
	// a fresh one. Closing the connection would drop the lock anyway.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", retentionLockKey); err != nil {
		// Closing would hand the connection back to the pool with the lock
		// still on it. Report it as broken so the pool throws it away.
		_ = conn.Raw(func(any) error { return driver.ErrBadConn })
	}
	conn.Close()
}
