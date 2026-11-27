package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"marketlens-go-backend/internal/models"
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
func (f *fakeRetentionStore) GetYearTotal(_ context.Context, year int) (*models.StatsYearTotal, error) {
	return f.totals[year], nil
}
func (f *fakeRetentionStore) RawVacancyTotal(context.Context, time.Time, time.Time) (int64, error) {
	return f.rawTotal, nil
}
func (f *fakeRetentionStore) RollUpYear(_ context.Context, year int, _, _ time.Time) error {
	f.calls = append(f.calls, "rollup")
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
	want := []string{"rollup", "verify", "purge", "purge", "purge", "purge", "mark"}
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
		if c == "purge" || c == "mark" {
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
		if c == "purge" || c == "mark" {
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
	if store.totals[2022].Status == models.YearStatusPurged {
		t.Error("the year was marked purged although the purge did not finish")
	}
}

func TestRetention_APurgedYearIsSkippedAndARestartContinues(t *testing.T) {
	store := newFakeStore(2021)
	store.totals[2021] = &models.StatsYearTotal{Year: 2021, Status: models.YearStatusPurged}
	// 2022 was rolled up by an earlier run that was killed during the purge.
	store.totals[2022] = &models.StatsYearTotal{Year: 2022, VacancyCount: 10, Status: models.YearStatusRolledUp}
	store.purgeLeft = 3

	if err := newTestRetention(t, store, true).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// Only 2022 is processed: rolled up again (to catch late jobs), then the
	// 3 jobs left are deleted in two batches and a final empty one.
	want := []string{"rollup", "verify", "purge", "purge", "purge", "mark"}
	if !reflect.DeepEqual(store.calls, want) {
		t.Errorf("calls = %v, want %v", store.calls, want)
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
