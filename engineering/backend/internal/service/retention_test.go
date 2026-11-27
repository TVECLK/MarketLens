package service

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"gorm.io/gorm"

	"marketlens-go-backend/internal/models"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/testutil"
)

// fakeRetentionStore records every call so the tests can check the order of
// the steps, in particular that nothing is deleted before the summary checks
// pass.
type fakeRetentionStore struct {
	earliest  *time.Time
	totals    map[int]*models.StatsYearTotal
	rawTotal  int64 // what RawVacancyTotal returns
	rollUpSum int64 // what RollUpYear stores as the year total
	verifyErr error
	purgeLeft int64 // jobs a year still has; PurgeBatch drains it
	purgeErr  error
	calls     []string
}

func newFakeStore(earliestYear int) *fakeRetentionStore {
	e := time.Date(earliestYear, 6, 1, 0, 0, 0, 0, time.UTC)
	return &fakeRetentionStore{earliest: &e, totals: map[int]*models.StatsYearTotal{}, rawTotal: 10, rollUpSum: 10, purgeLeft: 5}
}

func (f *fakeRetentionStore) EarliestPostedAt(context.Context) (*time.Time, error) {
	return f.earliest, nil
}
func (f *fakeRetentionStore) PurgingYears(context.Context) ([]int, error) {
	var years []int
	for y, t := range f.totals {
		if t.Status == models.YearStatusPurging {
			years = append(years, y)
		}
	}
	sort.Ints(years)
	return years, nil
}
func (f *fakeRetentionStore) GetYearTotal(_ context.Context, year int) (*models.StatsYearTotal, error) {
	return f.totals[year], nil
}
func (f *fakeRetentionStore) RawVacancyTotal(context.Context, time.Time, time.Time) (int64, error) {
	return f.rawTotal, nil
}
func (f *fakeRetentionStore) RollUpYear(_ context.Context, year int, _, _ time.Time) error {
	f.calls = append(f.calls, "rollup")
	// Like the real store: a year whose purge has started is never rolled up.
	if t := f.totals[year]; t != nil && t.Status != models.YearStatusRolledUp {
		return errors.New("fake: roll-up of a year whose purge has started")
	}
	f.totals[year] = &models.StatsYearTotal{Year: year, VacancyCount: f.rollUpSum, Status: models.YearStatusRolledUp}
	return nil
}
func (f *fakeRetentionStore) VerifyYearSummary(context.Context, int) error {
	f.calls = append(f.calls, "verify")
	return f.verifyErr
}
func (f *fakeRetentionStore) PurgeBatch(_ context.Context, _, _ time.Time, limit int) (int64, error) {
	f.calls = append(f.calls, "purge")
	if f.purgeErr != nil {
		return 0, f.purgeErr
	}
	n := f.purgeLeft
	if n > int64(limit) {
		n = int64(limit)
	}
	f.purgeLeft -= n
	return n, nil
}
func (f *fakeRetentionStore) MarkYearPurging(_ context.Context, year int) error {
	f.calls = append(f.calls, "purging")
	f.totals[year].Status = models.YearStatusPurging
	return nil
}
func (f *fakeRetentionStore) MarkYearPurged(_ context.Context, year int) error {
	f.calls = append(f.calls, "mark")
	f.totals[year].Status = models.YearStatusPurged
	return nil
}

func newTestRetention(t *testing.T, store retentionStore, purge bool) *RetentionService {
	t.Helper()
	s, err := NewRetentionService(store, RetentionConfig{
		RawYears:     3,
		Location:     time.UTC,
		BatchSize:    2,
		PurgeEnabled: purge,
		Now:          func() time.Time { return time.Date(2026, 11, 26, 9, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewRetentionService: %v", err)
	}
	return s
}

func TestRetention_OnlyYearsOutsideTheRawWindowAreProcessed(t *testing.T) {
	// In 2026 the raw window starts on 1 Jan 2023, so 2022 is the newest year
	// to roll up; with data since 2020 that is 2020, 2021 and 2022.
	store := newFakeStore(2020)
	s := newTestRetention(t, store, false)

	years, err := s.eligibleYears(context.Background())
	if err != nil {
		t.Fatalf("eligibleYears: %v", err)
	}
	if want := []int{2020, 2021, 2022}; !reflect.DeepEqual(years, want) {
		t.Errorf("eligible years = %v, want %v", years, want)
	}
}

func TestRetention_NothingToDoWhenAllDataIsRecent(t *testing.T) {
	for _, tc := range []struct {
		name string
		year int
	}{{"data starts in the window", 2023}, {"data starts this year", 2026}} {
		t.Run(tc.name, func(t *testing.T) {
			store := newFakeStore(tc.year)
			if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
				t.Fatalf("Run: %v", err)
			}
			if len(store.calls) != 0 {
				t.Errorf("calls = %v, want none", store.calls)
			}
		})
	}
}

func TestRetention_NothingToDoWithoutJobs(t *testing.T) {
	store := newFakeStore(2020)
	store.earliest = nil
	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(store.calls) != 0 {
		t.Errorf("calls = %v, want none", store.calls)
	}
}

func TestRetention_RollsUpVerifiesThenPurgesInBatches(t *testing.T) {
	store := newFakeStore(2022)
	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// 5 jobs in batches of 2: three batches, then a batch that finds none.
	want := []string{"rollup", "verify", "purging", "purge", "purge", "purge", "purge", "mark"}
	if !reflect.DeepEqual(store.calls, want) {
		t.Errorf("calls = %v, want %v", store.calls, want)
	}
	if store.totals[2022].Status != models.YearStatusPurged {
		t.Errorf("status = %q, want purged", store.totals[2022].Status)
	}
}

func TestRetention_WithPurgeDisabledItOnlyRollsUp(t *testing.T) {
	store := newFakeStore(2022)
	if err := newTestRetention(t, store, false).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if want := []string{"rollup", "verify"}; !reflect.DeepEqual(store.calls, want) {
		t.Errorf("calls = %v, want %v: nothing may be deleted while purge is off", store.calls, want)
	}
	if store.totals[2022].Status != models.YearStatusRolledUp {
		t.Errorf("status = %q, want rolled_up", store.totals[2022].Status)
	}
}

func TestRetention_ATotalMismatchStopsBeforeAnyDelete(t *testing.T) {
	store := newFakeStore(2022)
	store.rawTotal, store.rollUpSum = 12, 10 // a job arrived while the roll-up ran

	err := newTestRetention(t, store, true).Run(context.Background())
	if err == nil {
		t.Fatal("Run accepted a summary that does not match the raw data")
	}
	for _, c := range store.calls {
		if c == "purging" || c == "purge" || c == "mark" {
			t.Fatalf("calls = %v: a delete happened after a failed check", store.calls)
		}
	}
}

func TestRetention_AFailedVerifyStopsBeforeAnyDelete(t *testing.T) {
	store := newFakeStore(2022)
	store.verifyErr = errors.New("gender does not add up")

	if err := newTestRetention(t, store, true).Run(context.Background()); err == nil {
		t.Fatal("Run ignored a failed verification")
	}
	for _, c := range store.calls {
		if c == "purging" || c == "purge" || c == "mark" {
			t.Fatalf("calls = %v: a delete happened after a failed verification", store.calls)
		}
	}
}

func TestRetention_APurgeErrorLeavesTheYearUnmarked(t *testing.T) {
	store := newFakeStore(2022)
	store.purgeErr = errors.New("connection lost")

	if err := newTestRetention(t, store, true).Run(context.Background()); err == nil {
		t.Fatal("Run swallowed a purge error")
	}
	if got := store.totals[2022].Status; got != models.YearStatusPurging {
		t.Errorf("status = %q, want purging: the deletes started but did not finish", got)
	}
}

func TestRetention_APurgedYearIsSkippedAndACutOffPurgeResumesWithoutARollUp(t *testing.T) {
	store := newFakeStore(2021)
	store.totals[2021] = &models.StatsYearTotal{Year: 2021, Status: models.YearStatusPurged}
	// 2022 was cut off during its purge: some raw rows are already gone, so the
	// raw total no longer matches the summary and must not be looked at.
	store.totals[2022] = &models.StatsYearTotal{Year: 2022, VacancyCount: 100, Status: models.YearStatusPurging}
	store.rawTotal = 60
	store.purgeLeft = 3

	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// No roll-up and no "purging" mark: the summary is final. The 3 jobs left
	// go in two batches and a final empty one.
	want := []string{"verify", "purge", "purge", "purge", "mark"}
	if !reflect.DeepEqual(store.calls, want) {
		t.Errorf("calls = %v, want %v", store.calls, want)
	}
	if store.totals[2022].VacancyCount != 100 {
		t.Errorf("summary total = %d, want 100: it must not be recounted from the rows that remain", store.totals[2022].VacancyCount)
	}
}

func TestRetention_ARolledUpYearIsRolledUpAgainBeforeItsPurgeStarts(t *testing.T) {
	// A run that was cut off after the roll-up but before the first delete left
	// the year rolled_up with all its raw rows: rolling up again is safe.
	store := newFakeStore(2022)
	store.totals[2022] = &models.StatsYearTotal{Year: 2022, VacancyCount: 10, Status: models.YearStatusRolledUp}

	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(store.calls) == 0 || store.calls[0] != "rollup" {
		t.Errorf("calls = %v, want the roll-up first", store.calls)
	}
}

func TestRetention_ACutOffPurgeStaysUntouchedWhilePurgeIsDisabled(t *testing.T) {
	store := newFakeStore(2022)
	store.totals[2022] = &models.StatsYearTotal{Year: 2022, VacancyCount: 100, Status: models.YearStatusPurging}

	if err := newTestRetention(t, store, false).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(store.calls) != 0 {
		t.Errorf("calls = %v, want none: no roll-up and no delete while purge is off", store.calls)
	}
}

func TestRetention_AResumedPurgeStopsIfTheSummaryDoesNotVerify(t *testing.T) {
	store := newFakeStore(2022)
	store.totals[2022] = &models.StatsYearTotal{Year: 2022, VacancyCount: 100, Status: models.YearStatusPurging}
	store.verifyErr = errors.New("gender does not add up")

	if err := newTestRetention(t, store, true).Run(context.Background()); err == nil {
		t.Fatal("Run resumed a purge although the stored summary does not verify")
	}
	for _, c := range store.calls {
		if c == "purge" || c == "mark" {
			t.Fatalf("calls = %v: a delete happened after a failed verification", store.calls)
		}
	}
}

func TestRetention_AYearWithNoRawJobsLeftIsStillFinalizedWhenItIsPurging(t *testing.T) {
	// All batches of 2021 were deleted but the year was never marked purged, so
	// no raw job is left to point at it: there are no raw jobs at all.
	store := newFakeStore(2021)
	store.earliest = nil
	store.totals[2021] = &models.StatsYearTotal{Year: 2021, VacancyCount: 100, Status: models.YearStatusPurging}
	store.purgeLeft = 0

	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if want := []string{"verify", "purge", "mark"}; !reflect.DeepEqual(store.calls, want) {
		t.Errorf("calls = %v, want %v: no roll-up, the summary is final", store.calls, want)
	}
	if got := store.totals[2021]; got.Status != models.YearStatusPurged || got.VacancyCount != 100 {
		t.Errorf("year 2021 = %+v, want purged with the total still 100", got)
	}
}

func TestRetention_AnUnfinishedYearIsProcessedBeforeTheYearsFoundInTheRawData(t *testing.T) {
	store := newFakeStore(2022) // raw data starts in 2022
	store.totals[2021] = &models.StatsYearTotal{Year: 2021, VacancyCount: 100, Status: models.YearStatusPurging}
	store.purgeLeft = 0

	s := newTestRetention(t, store, true)
	years, err := s.eligibleYears(context.Background())
	if err != nil {
		t.Fatalf("eligibleYears: %v", err)
	}
	if want := []int{2021, 2022}; !reflect.DeepEqual(years, want) {
		t.Errorf("years = %v, want %v: oldest first, no duplicates", years, want)
	}
	if err := s.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, y := range []int{2021, 2022} {
		if store.totals[y].Status != models.YearStatusPurged {
			t.Errorf("year %d status = %q, want purged", y, store.totals[y].Status)
		}
	}
}

func TestRetention_AnUnfinishedYearInsideTheWindowIsStillFinalized(t *testing.T) {
	// RawYears was raised after a purge started: 2024 is inside the window now,
	// but half a purge is not left as it is.
	store := newFakeStore(2024)
	store.totals[2024] = &models.StatsYearTotal{Year: 2024, VacancyCount: 100, Status: models.YearStatusPurging}
	store.purgeLeft = 0

	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if store.totals[2024].Status != models.YearStatusPurged {
		t.Errorf("status = %q, want purged", store.totals[2024].Status)
	}
}

func TestRetention_ARolledUpYearWithoutRawJobsIsNotPulledIn(t *testing.T) {
	// Only purging years are added from the summary table. A rolled_up year has
	// all its raw rows, and rebuilding one from an empty range would write 0.
	store := newFakeStore(2022)
	store.earliest = nil
	store.totals[2021] = &models.StatsYearTotal{Year: 2021, VacancyCount: 100, Status: models.YearStatusRolledUp}

	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(store.calls) != 0 {
		t.Errorf("calls = %v, want none", store.calls)
	}
	if store.totals[2021].VacancyCount != 100 {
		t.Errorf("total = %d, want 100: it must not be rebuilt from an empty range", store.totals[2021].VacancyCount)
	}
}

func TestRetention_StopsWhenTheContextIsCancelled(t *testing.T) {
	store := newFakeStore(2022)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := newTestRetention(t, store, true).Run(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Run on a cancelled context = %v, want context.Canceled", err)
	}
	if store.totals[2022] != nil && store.totals[2022].Status == models.YearStatusPurged {
		t.Error("the year was marked purged after the context was cancelled")
	}
}

func TestRetention_YearBoundariesUseTheConfiguredLocation(t *testing.T) {
	colombo, err := time.LoadLocation("Asia/Colombo")
	if err != nil {
		t.Skipf("no tz database: %v", err)
	}
	s, _ := NewRetentionService(newFakeStore(2022), RetentionConfig{RawYears: 3, Location: colombo, BatchSize: 1})
	from, to := s.yearRange(2022)
	if from.UTC() != time.Date(2021, 12, 31, 18, 30, 0, 0, time.UTC) || to.UTC() != time.Date(2022, 12, 31, 18, 30, 0, 0, time.UTC) {
		t.Errorf("2022 in Asia/Colombo = [%v, %v), want 18:30 UTC on 31 Dec to 31 Dec", from.UTC(), to.UTC())
	}
}

func TestNewRetentionService_RejectsBadConfig(t *testing.T) {
	good := RetentionConfig{RawYears: 3, Location: time.UTC, BatchSize: 10}
	for name, mutate := range map[string]func(*RetentionConfig){
		"no raw years":  func(c *RetentionConfig) { c.RawYears = 0 },
		"no batch size": func(c *RetentionConfig) { c.BatchSize = 0 },
		"no location":   func(c *RetentionConfig) { c.Location = nil },
	} {
		cfg := good
		mutate(&cfg)
		if _, err := NewRetentionService(newFakeStore(2020), cfg); err == nil {
			t.Errorf("%s: NewRetentionService accepted the config", name)
		}
	}
}

// cutOffStore is the real store whose PurgeBatch fails after a number of
// batches, like a pod that is killed in the middle of a purge.
type cutOffStore struct {
	*repositories.Store
	allowed int // batches that succeed before the "kill"
	done    int
}

func (c *cutOffStore) PurgeBatch(ctx context.Context, from, to time.Time, limit int) (int64, error) {
	if c.done >= c.allowed {
		return 0, errors.New("pod killed")
	}
	c.done++
	return c.Store.PurgeBatch(ctx, from, to, limit)
}

func seedJobs(t *testing.T, db *gorm.DB, postedAt time.Time, vacancies ...int) {
	t.Helper()
	for _, v := range vacancies {
		job := models.JobPost{
			JobRole:       "role",
			WorkMode:      models.WorkModeOnsite,
			NoOfVacancies: v,
			MetaData:      models.JobMetaData{PostedAt: postedAt},
		}
		if err := db.Create(&job).Error; err != nil {
			t.Fatalf("failed to create job: %v", err)
		}
	}
}

func newRealRetention(t *testing.T, store retentionStore) *RetentionService {
	t.Helper()
	s, err := NewRetentionService(store, RetentionConfig{
		RawYears: 3, Location: time.UTC, BatchSize: 3, PurgeEnabled: true,
		Now: func() time.Time { return time.Date(2026, 11, 26, 9, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatalf("NewRetentionService: %v", err)
	}
	return s
}

// A purge that is cut off half way and then run again must keep the year total
// of the complete data. Recounting the rows that are left would store a total
// that is too low and lose the deleted jobs from the record.
func TestRetention_ACutOffPurgeKeepsTheCompleteYearTotal(t *testing.T) {
	db := testutil.SetupTestDB(t)
	real := repositories.NewStore(db)
	ctx := context.Background()

	// 10 jobs of 2022 worth 1+2+...+10 = 55 vacancies, and 2 jobs inside the window.
	seedJobs(t, db, time.Date(2022, 6, 1, 9, 0, 0, 0, time.UTC), 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	seedJobs(t, db, time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC), 100, 200)

	// First run: batches of 3, killed after 2 of them, so 6 jobs are deleted and 4 remain.
	if err := newRealRetention(t, &cutOffStore{Store: real, allowed: 2}).Run(ctx); err == nil {
		t.Fatal("the cut-off run reported success")
	}
	total, err := real.GetYearTotal(ctx, 2022)
	if err != nil || total == nil {
		t.Fatalf("GetYearTotal: %v, %v", total, err)
	}
	if total.Status != models.YearStatusPurging || total.VacancyCount != 55 {
		t.Fatalf("after the cut-off run: %+v, want status purging with total 55", total)
	}
	var left int64
	db.Model(&models.JobMetaData{}).Where("posted_at < ?", "2023-01-01").Count(&left)
	if left != 4 {
		t.Fatalf("%d jobs of 2022 left after the cut-off run, want 4", left)
	}

	// Second run on the same data: it must finish the deletes and keep 55.
	if err := newRealRetention(t, real).Run(ctx); err != nil {
		t.Fatalf("the second run: %v", err)
	}
	total, _ = real.GetYearTotal(ctx, 2022)
	if total.VacancyCount != 55 {
		t.Errorf("year total = %d, want 55: it was recounted from the rows left after the cut-off", total.VacancyCount)
	}
	if total.Status != models.YearStatusPurged || total.PurgedAt == nil {
		t.Errorf("after the second run: %+v, want status purged", total)
	}
	db.Model(&models.JobMetaData{}).Where("posted_at < ?", "2023-01-01").Count(&left)
	if left != 0 {
		t.Errorf("%d jobs of 2022 left after the second run, want 0", left)
	}
	db.Model(&models.JobMetaData{}).Where("posted_at >= ?", "2024-01-01").Count(&left)
	if left != 2 {
		t.Errorf("%d jobs inside the window left, want 2 (untouched)", left)
	}

	// A third run has nothing to do.
	if err := newRealRetention(t, real).Run(ctx); err != nil {
		t.Fatalf("the third run: %v", err)
	}
}

// markPurgedFailsStore is the real store whose MarkYearPurged fails once, like a
// write that fails or a pod that is killed after the last batch was deleted.
type markPurgedFailsStore struct {
	*repositories.Store
	failures int
}

func (m *markPurgedFailsStore) MarkYearPurged(ctx context.Context, year int) error {
	if m.failures > 0 {
		m.failures--
		return errors.New("write failed")
	}
	return m.Store.MarkYearPurged(ctx, year)
}

// If every batch is deleted but marking the year purged fails, no raw job of
// that year is left to point at it. The next run must still find the year and
// finish it, or it stays purging and its vacancies drop out of the dashboard's
// all-time count.
func TestRetention_AYearWhoseLastStepFailedIsFinalizedOnTheNextRun(t *testing.T) {
	db := testutil.SetupTestDB(t)
	real := repositories.NewStore(db)
	ctx := context.Background()

	// 2022: 1+2+...+10 = 55 vacancies. Inside the window: 100+200 = 300.
	seedJobs(t, db, time.Date(2022, 6, 1, 9, 0, 0, 0, time.UTC), 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	seedJobs(t, db, time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC), 100, 200)

	// First run: all batches commit, then the final write fails.
	if err := newRealRetention(t, &markPurgedFailsStore{Store: real, failures: 1}).Run(ctx); err == nil {
		t.Fatal("the first run reported success although marking the year purged failed")
	}
	var left int64
	db.Model(&models.JobMetaData{}).Where("posted_at < ?", "2023-01-01").Count(&left)
	if left != 0 {
		t.Fatalf("%d jobs of 2022 left after the first run, want 0 (every batch committed)", left)
	}
	total, _ := real.GetYearTotal(ctx, 2022)
	if total.Status != models.YearStatusPurging || total.VacancyCount != 55 {
		t.Fatalf("after the first run: %+v, want status purging with total 55", total)
	}
	card, err := real.GetDashboardCardData(ctx)
	if err != nil {
		t.Fatalf("GetDashboardCardData: %v", err)
	}
	if card.AllTimeVacancyCount != 300 {
		t.Fatalf("all-time count while stuck = %d, want 300 (the 55 are in neither place yet)", card.AllTimeVacancyCount)
	}

	// The oldest raw job is now from 2024, inside the window, so only the
	// summary table can point at 2022.
	if earliest, _ := real.EarliestPostedAt(ctx); earliest == nil || earliest.Year() != 2024 {
		t.Fatalf("earliest raw job = %v, want one from 2024", earliest)
	}

	if err := newRealRetention(t, real).Run(ctx); err != nil {
		t.Fatalf("the second run: %v", err)
	}
	total, _ = real.GetYearTotal(ctx, 2022)
	if total.Status != models.YearStatusPurged || total.PurgedAt == nil || total.VacancyCount != 55 {
		t.Errorf("after the second run: %+v, want purged with the total still 55", total)
	}
	card, _ = real.GetDashboardCardData(ctx)
	if card.AllTimeVacancyCount != 355 {
		t.Errorf("all-time count after the second run = %d, want 355 (300 raw + 55 purged)", card.AllTimeVacancyCount)
	}
	if err := newRealRetention(t, real).Run(ctx); err != nil {
		t.Fatalf("the third run: %v", err)
	}
}
