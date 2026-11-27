package repositories_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/testutil"
)

// ---------- Seed and query helpers ----------

func utcDate(year int, month time.Month, day, hour int) time.Time {
	return time.Date(year, month, day, hour, 0, 0, 0, time.UTC)
}

func year(y int) (time.Time, time.Time) {
	return utcDate(y, 1, 1, 0), utcDate(y+1, 1, 1, 0)
}

func ptr[T any](v T) *T { return &v }

// seedOccupationLeaf creates sub-major, minor, unit and occupation groups
// under majorID and returns the occupation group (the leaf).
func seedOccupationLeaf(t *testing.T, db *gorm.DB, majorID uint) uint {
	t.Helper()
	sub := models.SubMajorGroup{MajorGroupID: majorID, Name: "sub", Code: "1"}
	minor := models.MinorGroup{Name: "minor", Code: "1"}
	unit := models.UnitGroup{Name: "unit", Code: "1"}
	leaf := models.OccupationGroup{Name: "leaf", Code: "1"}
	if err := db.Create(&sub).Error; err != nil {
		t.Fatalf("failed to create sub-major group: %v", err)
	}
	minor.SubMajorGroupID = sub.ID
	if err := db.Create(&minor).Error; err != nil {
		t.Fatalf("failed to create minor group: %v", err)
	}
	unit.MinorGroupID = minor.ID
	if err := db.Create(&unit).Error; err != nil {
		t.Fatalf("failed to create unit group: %v", err)
	}
	leaf.UnitGroupID = unit.ID
	if err := db.Create(&leaf).Error; err != nil {
		t.Fatalf("failed to create occupation group: %v", err)
	}
	return leaf.ID
}

// seedIndustryLeaf creates division, group, class and subclass under sectorID
// and returns the subclass (the leaf).
func seedIndustryLeaf(t *testing.T, db *gorm.DB, sectorID uint) uint {
	t.Helper()
	division := models.IndustryDivision{IndustrySectorID: sectorID, Name: "division", Code: "1"}
	group := models.IndustryGroup{Name: "group", Code: "1"}
	class := models.IndustryClass{Name: "class", Code: "1"}
	sub := models.IndustrySubclass{Name: "subclass", Code: "1"}
	if err := db.Create(&division).Error; err != nil {
		t.Fatalf("failed to create industry division: %v", err)
	}
	group.IndustryDivisionID = division.ID
	if err := db.Create(&group).Error; err != nil {
		t.Fatalf("failed to create industry group: %v", err)
	}
	class.IndustryGroupID = group.ID
	if err := db.Create(&class).Error; err != nil {
		t.Fatalf("failed to create industry class: %v", err)
	}
	sub.IndustryClassID = class.ID
	if err := db.Create(&sub).Error; err != nil {
		t.Fatalf("failed to create industry subclass: %v", err)
	}
	return sub.ID
}

type jobSpec struct {
	postedAt   time.Time
	vacancies  int
	workMode   models.WorkMode
	jobTypeID  *uint
	genderID   *uint
	provinceID *uint
	occupation *uint // occupation_group id
	industry   *uint // industry_subclass id
	bucketKeys []string
	skillNames []string
}

func seedJobSpec(t *testing.T, db *gorm.DB, s jobSpec) uint {
	t.Helper()
	if s.workMode == "" {
		s.workMode = models.WorkModeOnsite
	}
	job := models.JobPost{
		JobRole:       "role",
		WorkMode:      s.workMode,
		NoOfVacancies: s.vacancies,
		JobTypeID:     s.jobTypeID,
		MetaData: models.JobMetaData{
			PostedAt:           s.postedAt,
			GenderID:           s.genderID,
			GeoDataID:          s.provinceID,
			OccupationGroupID:  s.occupation,
			IndustrySubclassID: s.industry,
		},
	}
	for _, name := range s.skillNames {
		job.Skills = append(job.Skills, models.Skill{Skill: name})
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("failed to create job: %v", err)
	}
	for i, key := range s.bucketKeys {
		if err := db.Create(&models.LshIndex{BucketKey: key, BandNo: i, JobPostID: job.ID}).Error; err != nil {
			t.Fatalf("failed to create lsh row: %v", err)
		}
	}
	return job.ID
}

// statValue reads one summary row; a missing row reads as 0.
func statValue(t *testing.T, db *gorm.DB, table string, root uint, dimension string, yr, valueID int) int64 {
	t.Helper()
	var rows []int64
	err := db.Table(table).
		Where("root_id = ? AND dimension = ? AND year = ? AND value_id = ?", root, dimension, yr, valueID).
		Pluck("vacancy_count", &rows).Error
	if err != nil {
		t.Fatalf("failed to read %s: %v", table, err)
	}
	if len(rows) == 0 {
		return 0
	}
	return rows[0]
}

// retentionFixture is a small data set with every case the roll-up handles.
//
//	j1 2022-03-01  3 vac  male, type A, province P, remote, occupation leaf L1 (major M1), industry leaf I1 (sector S1)
//	j2 2022-06-01  2 vac  no gender, no type, onsite, occupation leaf L2 (major M1), no industry
//	j3 2022-12-31 23:00  4 vac  hybrid, occupation leaf L3 (major M2), industry leaf I1 (sector S1)
//	j4 2022-05-01  5 vac  unclassified: no occupation, no industry
//	j5 2023-01-01 00:00  7 vac  next year: must be left out
type retentionFixture struct {
	db                 *gorm.DB
	store              *repositories.Store
	m1, m2             uint
	s1                 uint
	male, typeA        uint
	province           uint
	j1, j2, j3, j4, j5 uint
}

func newRetentionFixture(t *testing.T) retentionFixture {
	t.Helper()
	db := testutil.SetupTestDB(t)
	f := retentionFixture{db: db, store: repositories.NewStore(db)}

	m1, m2 := models.MajorGroup{Name: "M1", Code: "1"}, models.MajorGroup{Name: "M2", Code: "2"}
	s1 := models.IndustrySector{Name: "S1", Code: "A"}
	for _, row := range []any{&m1, &m2, &s1} {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("failed to create root group: %v", err)
		}
	}
	f.m1, f.m2, f.s1 = m1.ID, m2.ID, s1.ID

	male := models.Gender{GenderType: "male"}
	typeA := models.JobType{Type: "full time"}
	province := models.GeoData{Province: "Western"}
	for _, row := range []any{&male, &typeA, &province} {
		if err := db.Create(row).Error; err != nil {
			t.Fatalf("failed to create lookup row: %v", err)
		}
	}
	f.male, f.typeA, f.province = male.ID, typeA.ID, province.ID

	l1, l2 := seedOccupationLeaf(t, db, f.m1), seedOccupationLeaf(t, db, f.m1)
	l3 := seedOccupationLeaf(t, db, f.m2)
	i1 := seedIndustryLeaf(t, db, f.s1)

	f.j1 = seedJobSpec(t, db, jobSpec{postedAt: utcDate(2022, 3, 1, 0), vacancies: 3, workMode: models.WorkModeRemote,
		genderID: &f.male, jobTypeID: &f.typeA, provinceID: &f.province, occupation: &l1, industry: &i1})
	f.j2 = seedJobSpec(t, db, jobSpec{postedAt: utcDate(2022, 6, 1, 0), vacancies: 2, workMode: models.WorkModeOnsite,
		occupation: &l2})
	f.j3 = seedJobSpec(t, db, jobSpec{postedAt: utcDate(2022, 12, 31, 23), vacancies: 4, workMode: models.WorkModeHybrid,
		occupation: &l3, industry: &i1})
	f.j4 = seedJobSpec(t, db, jobSpec{postedAt: utcDate(2022, 5, 1, 0), vacancies: 5})
	f.j5 = seedJobSpec(t, db, jobSpec{postedAt: utcDate(2023, 1, 1, 0), vacancies: 7, occupation: &l1})
	return f
}

// ---------- RollUpYear ----------

func TestStore_RollUpYear_WritesTheSummary(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)

	if err := f.store.RollUpYear(context.Background(), 2022, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}

	total, err := f.store.GetYearTotal(context.Background(), 2022)
	if err != nil || total == nil {
		t.Fatalf("GetYearTotal: %v, %v", total, err)
	}
	if total.VacancyCount != 14 || total.Status != models.YearStatusRolledUp {
		t.Errorf("year total = %d (%s), want 14 (rolled_up): unclassified jobs count, next year's job does not",
			total.VacancyCount, total.Status)
	}

	occ, ind := "stats_year_occupation", "stats_year_industry"
	checks := []struct {
		name  string
		table string
		root  uint
		dim   string
		value int
		want  int64
	}{
		{"M1 total adds its two leaves", occ, f.m1, models.DimensionTotal, 0, 5},
		{"M2 total", occ, f.m2, models.DimensionTotal, 0, 4},
		{"sector total ignores jobs with no industry", ind, f.s1, models.DimensionTotal, 0, 7},
		{"gender male", occ, f.m1, models.DimensionGender, int(f.male), 3},
		{"gender unknown is value 0", occ, f.m1, models.DimensionGender, models.UnknownValueID, 2},
		{"job type A", occ, f.m1, models.DimensionJobType, int(f.typeA), 3},
		{"job type unknown", occ, f.m1, models.DimensionJobType, models.UnknownValueID, 2},
		{"province", occ, f.m1, models.DimensionProvince, int(f.province), 3},
		{"work mode remote is code 1", occ, f.m1, models.DimensionWorkMode, 1, 3},
		{"work mode onsite is code 2", occ, f.m1, models.DimensionWorkMode, 2, 2},
		{"work mode hybrid is code 3", occ, f.m2, models.DimensionWorkMode, 3, 4},
		{"industry work mode", ind, f.s1, models.DimensionWorkMode, 1, 3},
		{"job at 23:00 on 31 Dec is in the year", occ, f.m2, models.DimensionGender, models.UnknownValueID, 4},
	}
	for _, c := range checks {
		if got := statValue(t, f.db, c.table, c.root, c.dim, 2022, c.value); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}

	if err := f.store.VerifyYearSummary(context.Background(), 2022); err != nil {
		t.Errorf("VerifyYearSummary on a fresh roll-up: %v", err)
	}
}

func TestStore_RollUpYear_IsRepeatableAndPicksUpLateJobs(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
			t.Fatalf("RollUpYear #%d: %v", i+1, err)
		}
	}
	if got := countRows(t, f.db, &models.StatsYearTotal{}, ""); got != 1 {
		t.Errorf("stats_year_total rows after two roll-ups = %d, want 1", got)
	}
	if got := statValue(t, f.db, "stats_year_occupation", f.m1, models.DimensionTotal, 2022, 0); got != 5 {
		t.Errorf("M1 total after two roll-ups = %d, want 5 (not doubled)", got)
	}

	late := seedOccupationLeaf(t, f.db, f.m1)
	seedJobSpec(t, f.db, jobSpec{postedAt: utcDate(2022, 8, 1, 0), vacancies: 10, occupation: &late})
	if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
		t.Fatalf("RollUpYear after a late job: %v", err)
	}
	if got := statValue(t, f.db, "stats_year_occupation", f.m1, models.DimensionTotal, 2022, 0); got != 15 {
		t.Errorf("M1 total after a late job = %d, want 15", got)
	}
	total, _ := f.store.GetYearTotal(ctx, 2022)
	if total.VacancyCount != 24 {
		t.Errorf("year total after a late job = %d, want 24", total.VacancyCount)
	}
}

func TestStore_RollUpYear_RefusesAYearWhosePurgeHasStarted(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)
	ctx := context.Background()
	if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}
	if err := f.store.MarkYearPurging(ctx, 2022); err != nil {
		t.Fatalf("MarkYearPurging: %v", err)
	}

	// A cut-off purge: 2 of the 4 jobs of 2022 are gone.
	if n, err := f.store.PurgeBatch(ctx, from, to, 2); err != nil || n != 2 {
		t.Fatalf("PurgeBatch = %d, %v, want 2", n, err)
	}
	err := f.store.RollUpYear(ctx, 2022, from, to)
	if !errors.Is(err, repositories.ErrYearPurgeStarted) {
		t.Fatalf("RollUpYear on a purging year = %v, want ErrYearPurgeStarted", err)
	}
	total, _ := f.store.GetYearTotal(ctx, 2022)
	if total.VacancyCount != 14 {
		t.Errorf("year total after the refused roll-up = %d, want 14: it must not be recounted from the rows that remain", total.VacancyCount)
	}
	if got := statValue(t, f.db, "stats_year_occupation", f.m1, models.DimensionTotal, 2022, 0); got != 5 {
		t.Errorf("M1 total after the refused roll-up = %d, want 5 (summary untouched)", got)
	}

	// Finish the purge: a purged year is refused as well.
	if _, err := f.store.PurgeBatch(ctx, from, to, 100); err != nil {
		t.Fatalf("PurgeBatch: %v", err)
	}
	if err := f.store.MarkYearPurged(ctx, 2022); err != nil {
		t.Fatalf("MarkYearPurged: %v", err)
	}
	if err := f.store.RollUpYear(ctx, 2022, from, to); !errors.Is(err, repositories.ErrYearPurgeStarted) {
		t.Fatalf("RollUpYear on a purged year = %v, want ErrYearPurgeStarted", err)
	}
	if got := statValue(t, f.db, "stats_year_occupation", f.m1, models.DimensionTotal, 2022, 0); got != 5 {
		t.Errorf("M1 total after the purge = %d, want 5", got)
	}
}

func TestStore_RollUpYear_FailsLoudlyOnAWorkModeWithoutACode(t *testing.T) {
	f := newRetentionFixture(t)
	seedJobSpec(t, f.db, jobSpec{postedAt: utcDate(2021, 4, 1, 0), vacancies: 1, workMode: "freelance",
		occupation: ptr(seedOccupationLeaf(t, f.db, f.m1))})
	from, to := year(2021)

	if err := f.store.RollUpYear(context.Background(), 2021, from, to); err == nil {
		t.Fatal("RollUpYear accepted a work mode that has no code")
	}
	if total, _ := f.store.GetYearTotal(context.Background(), 2021); total != nil {
		t.Error("a failed roll-up left a summary row behind: the transaction was not rolled back")
	}
}

func TestStore_RollUpYear_AnEmptyYear(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2015)
	ctx := context.Background()

	if err := f.store.RollUpYear(ctx, 2015, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}
	if total, _ := f.store.GetYearTotal(ctx, 2015); total == nil || total.VacancyCount != 0 {
		t.Errorf("empty year total = %+v, want a row with 0", total)
	}
	if err := f.store.VerifyYearSummary(ctx, 2015); err != nil {
		t.Errorf("VerifyYearSummary on an empty year: %v", err)
	}
}

// ---------- VerifyYearSummary ----------

func TestStore_VerifyYearSummary_CatchesAWrongDimension(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)
	ctx := context.Background()
	if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}

	err := f.db.Model(&models.StatsYearOccupation{}).
		Where("root_id = ? AND dimension = ? AND value_id = ?", f.m1, models.DimensionGender, f.male).
		Update("vacancy_count", 99).Error
	if err != nil {
		t.Fatalf("failed to tamper with the summary: %v", err)
	}
	if err := f.store.VerifyYearSummary(ctx, 2022); err == nil {
		t.Error("VerifyYearSummary accepted a gender breakdown that does not add up to the total")
	}
}

func TestStore_VerifyYearSummary_CatchesAMissingDimension(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)
	ctx := context.Background()
	if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}
	if err := f.db.Where("dimension = ?", models.DimensionProvince).Delete(&models.StatsYearIndustry{}).Error; err != nil {
		t.Fatalf("failed to delete rows: %v", err)
	}
	if err := f.store.VerifyYearSummary(ctx, 2022); err == nil {
		t.Error("VerifyYearSummary accepted a summary with a whole dimension missing")
	}
}

func TestStore_VerifyYearSummary_CatchesMistakesThatCancelOutAcrossRoots(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)
	ctx := context.Background()
	if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}

	// M1 totals 5 and M2 totals 4. Make M1's gender figures add up to 6 and
	// M2's to 3: each root is wrong, but the two gender sums together are 9, the
	// same as the two totals together, so a check across roots would pass it.
	err := f.db.Model(&models.StatsYearOccupation{}).
		Where("root_id = ? AND dimension = ? AND value_id = ?", f.m1, models.DimensionGender, f.male).
		Update("vacancy_count", 4).Error
	if err != nil {
		t.Fatalf("failed to tamper with M1: %v", err)
	}
	err = f.db.Model(&models.StatsYearOccupation{}).
		Where("root_id = ? AND dimension = ? AND value_id = ?", f.m2, models.DimensionGender, models.UnknownValueID).
		Update("vacancy_count", 3).Error
	if err != nil {
		t.Fatalf("failed to tamper with M2: %v", err)
	}

	err = f.store.VerifyYearSummary(ctx, 2022)
	if err == nil {
		t.Fatal("VerifyYearSummary accepted two roots whose mistakes cancel each other out")
	}
	if want := fmt.Sprintf("root %d", f.m1); !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to name %s, the first wrong root", err, want)
	}
}

func TestStore_VerifyYearSummary_CatchesARootThatHasNoTotal(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)
	ctx := context.Background()
	if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}
	// M2 loses its 'total' row but keeps its other dimensions.
	err := f.db.Where("root_id = ? AND dimension = ?", f.m2, models.DimensionTotal).
		Delete(&models.StatsYearOccupation{}).Error
	if err != nil {
		t.Fatalf("failed to delete the total row: %v", err)
	}
	if err := f.store.VerifyYearSummary(ctx, 2022); err == nil {
		t.Error("VerifyYearSummary accepted a root whose dimensions have no total")
	}
}

// ---------- PurgeBatch / year statuses / EarliestPostedAt ----------

func TestStore_PurgeBatch_DeletesInBatchesAndCascades(t *testing.T) {
	db := testutil.SetupTestDB(t)
	store := repositories.NewStore(db)
	ctx := context.Background()
	from, to := year(2022)

	var old []uint
	for i := 0; i < 5; i++ {
		old = append(old, seedJobSpec(t, db, jobSpec{postedAt: utcDate(2022, 5, 1+i, 0), vacancies: 1,
			bucketKeys: []string{"k" + string(rune('a'+i)), "k" + string(rune('A'+i))}, skillNames: []string{"go"}}))
	}
	keep := seedJobSpec(t, db, jobSpec{postedAt: utcDate(2023, 1, 1, 0), vacancies: 1, bucketKeys: []string{"keep"}})

	for _, want := range []int64{2, 2, 1, 0} {
		got, err := store.PurgeBatch(ctx, from, to, 2)
		if err != nil {
			t.Fatalf("PurgeBatch: %v", err)
		}
		if got != want {
			t.Fatalf("PurgeBatch deleted %d, want %d", got, want)
		}
	}

	if got := countRows(t, db, &models.JobPost{}, "id IN ?", old); got != 0 {
		t.Errorf("%d old job_post rows left", got)
	}
	if got := countRows(t, db, &models.JobMetaData{}, "job_post_id IN ?", old); got != 0 {
		t.Errorf("%d old meta_data rows left: the delete did not cascade", got)
	}
	if got := countRows(t, db, &models.LshIndex{}, "job_post_id IN ?", old); got != 0 {
		t.Errorf("%d old lsh_index rows left: the delete did not cascade", got)
	}
	var skillLinks int64
	db.Table("job_post_skills").Where("job_post_id IN ?", old).Count(&skillLinks)
	if skillLinks != 0 {
		t.Errorf("%d old job_post_skills rows left: the delete did not cascade", skillLinks)
	}

	if got := countRows(t, db, &models.JobPost{}, "id = ?", keep); got != 1 {
		t.Error("a job outside the range was deleted")
	}
	if got := countRows(t, db, &models.LshIndex{}, "job_post_id = ?", keep); got != 1 {
		t.Error("an lsh row of a job outside the range was deleted")
	}
}

func TestStore_YearStatusesGoInOrder(t *testing.T) {
	f := newRetentionFixture(t)
	ctx := context.Background()
	from, to := year(2022)

	if err := f.store.MarkYearPurging(ctx, 2022); err == nil {
		t.Error("MarkYearPurging accepted a year that was never rolled up")
	}
	if err := f.store.RollUpYear(ctx, 2022, from, to); err != nil {
		t.Fatalf("RollUpYear: %v", err)
	}
	if err := f.store.MarkYearPurged(ctx, 2022); err == nil {
		t.Error("MarkYearPurged accepted a year whose purge never started")
	}

	if err := f.store.MarkYearPurging(ctx, 2022); err != nil {
		t.Fatalf("MarkYearPurging: %v", err)
	}
	if total, _ := f.store.GetYearTotal(ctx, 2022); total.Status != models.YearStatusPurging {
		t.Errorf("status = %q, want purging", total.Status)
	}
	if err := f.store.MarkYearPurging(ctx, 2022); err == nil {
		t.Error("MarkYearPurging accepted a year that is already purging")
	}

	if err := f.store.MarkYearPurged(ctx, 2022); err != nil {
		t.Fatalf("MarkYearPurged: %v", err)
	}
	total, _ := f.store.GetYearTotal(ctx, 2022)
	if total.Status != models.YearStatusPurged || total.PurgedAt == nil {
		t.Errorf("after MarkYearPurged: %+v, want status purged with purged_at set", total)
	}
	if err := f.store.MarkYearPurged(ctx, 2022); err == nil {
		t.Error("MarkYearPurged accepted a year that is already purged")
	}
}

func TestStore_PurgingYears(t *testing.T) {
	db := testutil.SetupTestDB(t)
	store := repositories.NewStore(db)
	ctx := context.Background()

	if got, err := store.PurgingYears(ctx); err != nil || len(got) != 0 {
		t.Fatalf("PurgingYears with no summaries = %v, %v, want none", got, err)
	}
	now := time.Now()
	rows := []models.StatsYearTotal{
		{Year: 2022, VacancyCount: 1, Status: models.YearStatusPurging, RolledUpAt: now},
		{Year: 2019, VacancyCount: 1, Status: models.YearStatusPurged, RolledUpAt: now, PurgedAt: &now},
		{Year: 2021, VacancyCount: 1, Status: models.YearStatusRolledUp, RolledUpAt: now},
		{Year: 2020, VacancyCount: 1, Status: models.YearStatusPurging, RolledUpAt: now},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("failed to seed summaries: %v", err)
	}

	got, err := store.PurgingYears(ctx)
	if err != nil {
		t.Fatalf("PurgingYears: %v", err)
	}
	if len(got) != 2 || got[0] != 2020 || got[1] != 2022 {
		t.Errorf("PurgingYears = %v, want [2020 2022]: only purging years, oldest first", got)
	}
}

func TestStore_EarliestPostedAt(t *testing.T) {
	db := testutil.SetupTestDB(t)
	store := repositories.NewStore(db)
	ctx := context.Background()

	if got, err := store.EarliestPostedAt(ctx); err != nil || got != nil {
		t.Fatalf("EarliestPostedAt on no jobs = %v, %v, want nil", got, err)
	}
	seedJobSpec(t, db, jobSpec{postedAt: utcDate(2024, 2, 1, 0), vacancies: 1})
	seedJobSpec(t, db, jobSpec{postedAt: utcDate(2021, 9, 15, 0), vacancies: 1})

	got, err := store.EarliestPostedAt(ctx)
	if err != nil || got == nil {
		t.Fatalf("EarliestPostedAt = %v, %v", got, err)
	}
	if got.Year() != 2021 || got.Month() != time.September {
		t.Errorf("EarliestPostedAt = %v, want 2021-09-15", got)
	}
}

func TestStore_RawVacancyTotal_CountsUnclassifiedJobsAndRespectsTheRange(t *testing.T) {
	f := newRetentionFixture(t)
	from, to := year(2022)
	got, err := f.store.RawVacancyTotal(context.Background(), from, to)
	if err != nil {
		t.Fatalf("RawVacancyTotal: %v", err)
	}
	if got != 14 {
		t.Errorf("RawVacancyTotal = %d, want 14", got)
	}
}

// ---------- GetDashboardCardData ----------

func TestStore_GetDashboardCardData_AddsPurgedYearsButNotRolledUpOrPurgingOnes(t *testing.T) {
	db := testutil.SetupTestDB(t)
	store := repositories.NewStore(db)
	ctx := context.Background()

	seedJobSpec(t, db, jobSpec{postedAt: utcDate(2025, 5, 1, 0), vacancies: 4})
	seedJobSpec(t, db, jobSpec{postedAt: utcDate(2025, 6, 1, 0), vacancies: 6})

	card, err := store.GetDashboardCardData(ctx)
	if err != nil {
		t.Fatalf("GetDashboardCardData: %v", err)
	}
	if card.AllTimeVacancyCount != 10 {
		t.Fatalf("all-time count with no summaries = %d, want 10", card.AllTimeVacancyCount)
	}

	now := time.Now()
	rows := []models.StatsYearTotal{
		{Year: 2021, VacancyCount: 1000, Status: models.YearStatusPurged, RolledUpAt: now, PurgedAt: &now},
		{Year: 2022, VacancyCount: 2000, Status: models.YearStatusPurged, RolledUpAt: now, PurgedAt: &now},
		// Rolled up, raw rows still there: already counted in the raw sum.
		{Year: 2023, VacancyCount: 500, Status: models.YearStatusRolledUp, RolledUpAt: now},
		// Purging: some raw rows are still there and counted, so not added either.
		{Year: 2020, VacancyCount: 700, Status: models.YearStatusPurging, RolledUpAt: now},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("failed to seed summaries: %v", err)
	}

	card, err = store.GetDashboardCardData(ctx)
	if err != nil {
		t.Fatalf("GetDashboardCardData: %v", err)
	}
	if card.AllTimeVacancyCount != 3010 {
		t.Errorf("all-time count = %d, want 3010 (10 raw + 3000 from purged years; the rolled-up and purging years not added)", card.AllTimeVacancyCount)
	}
}
