package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"marketlens-go-backend/internal/models"
)

// retentionStore is the part of repositories.Store the retention job uses.
type retentionStore interface {
	EarliestPostedAt(ctx context.Context) (*time.Time, error)
	GetYearTotal(ctx context.Context, year int) (*models.StatsYearTotal, error)
	RawVacancyTotal(ctx context.Context, from, to time.Time) (int64, error)
	RollUpYear(ctx context.Context, year int, from, to time.Time) error
	VerifyYearSummary(ctx context.Context, year int) error
	PurgeBatch(ctx context.Context, from, to time.Time, limit int) (int64, error)
	MarkYearPurging(ctx context.Context, year int) error
	MarkYearPurged(ctx context.Context, year int) error
}

// RetentionConfig sets which years are rolled up and purged.
type RetentionConfig struct {
	// RawYears is how many calendar years before the current one keep their
	// raw rows: with 3 in 2026, raw data starts on 1 Jan 2023 and 2022 is the
	// newest year that gets rolled up. The frontend date picker uses the same
	// boundary.
	RawYears int
	// Location decides where a year starts and ends.
	Location *time.Location
	// BatchSize is how many jobs one purge transaction deletes.
	BatchSize int
	// PurgeEnabled turns the delete step on. With it off the job only rolls
	// up and verifies, so the summaries can be checked first.
	PurgeEnabled bool
	// Now returns the current time. A zero value means time.Now.
	Now func() time.Time
}

// RetentionService rolls old years up into the yearly summary tables and
// deletes their raw rows.
type RetentionService struct {
	store retentionStore
	cfg   RetentionConfig
}

func NewRetentionService(store retentionStore, cfg RetentionConfig) (*RetentionService, error) {
	if cfg.RawYears < 1 {
		return nil, errors.New("retention: RawYears must be at least 1")
	}
	if cfg.BatchSize < 1 {
		return nil, errors.New("retention: BatchSize must be at least 1")
	}
	if cfg.Location == nil {
		return nil, errors.New("retention: Location is required")
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &RetentionService{store: store, cfg: cfg}, nil
}

// Run processes every year that has left the raw-data window, oldest first,
// and stops at the first error. Years already purged are skipped, and a year
// whose purge was cut off is not rolled up again but has its deletes finished,
// so running it again after a failure or a pod kill continues where it stopped.
func (s *RetentionService) Run(ctx context.Context) error {
	years, err := s.eligibleYears(ctx)
	if err != nil {
		return err
	}
	if len(years) == 0 {
		log.Println("retention: no year has left the raw-data window")
		return nil
	}
	for _, year := range years {
		if err := s.processYear(ctx, year); err != nil {
			return fmt.Errorf("year %d: %w", year, err)
		}
	}
	return nil
}

// yearRange returns [start of year, start of the next year) in the
// configured location.
func (s *RetentionService) yearRange(year int) (time.Time, time.Time) {
	loc := s.cfg.Location
	return time.Date(year, 1, 1, 0, 0, 0, 0, loc), time.Date(year+1, 1, 1, 0, 0, 0, 0, loc)
}

// eligibleYears lists the years from the one holding the oldest raw job up
// to the newest one that has left the raw-data window.
func (s *RetentionService) eligibleYears(ctx context.Context) ([]int, error) {
	earliest, err := s.store.EarliestPostedAt(ctx)
	if err != nil {
		return nil, err
	}
	if earliest == nil {
		return nil, nil
	}
	cutoffYear := s.cfg.Now().In(s.cfg.Location).Year() - s.cfg.RawYears
	var years []int
	for y := earliest.In(s.cfg.Location).Year(); y < cutoffYear; y++ {
		years = append(years, y)
	}
	return years, nil
}

func (s *RetentionService) processYear(ctx context.Context, year int) error {
	total, err := s.store.GetYearTotal(ctx, year)
	if err != nil {
		return err
	}
	from, to := s.yearRange(year)

	if total != nil {
		switch total.Status {
		case models.YearStatusPurged:
			return nil
		case models.YearStatusPurging:
			return s.resumePurge(ctx, year, from, to)
		}
	}

	// Roll up every time while no delete has started, even when the year is
	// already rolled up: the summary is replaced, so jobs that arrived late are
	// included. Once the purge starts the summary is final and this is skipped.
	if err := s.store.RollUpYear(ctx, year, from, to); err != nil {
		return err
	}
	if err := s.checkSummary(ctx, year, from, to); err != nil {
		return err
	}
	log.Printf("retention: year %d rolled up and verified", year)

	if !s.cfg.PurgeEnabled {
		log.Printf("retention: purge is disabled, raw jobs of year %d kept", year)
		return nil
	}
	// From this call on the year is never rolled up again: a run cut off in the
	// middle of the deletes must not recount the rows that are left.
	if err := s.store.MarkYearPurging(ctx, year); err != nil {
		return err
	}
	return s.finishPurge(ctx, year, from, to)
}

// resumePurge finishes the deletes of a year whose purge was cut off. Some of
// its raw rows are already gone, so the stored summary is the only complete
// record: it is not rebuilt and not compared with the raw rows. It is only
// checked for adding up within itself.
func (s *RetentionService) resumePurge(ctx context.Context, year int, from, to time.Time) error {
	if !s.cfg.PurgeEnabled {
		log.Printf("retention: purge is disabled, the purge of year %d stays half done", year)
		return nil
	}
	if err := s.store.VerifyYearSummary(ctx, year); err != nil {
		return fmt.Errorf("%w; nothing was deleted", err)
	}
	log.Printf("retention: resuming the purge of year %d", year)
	return s.finishPurge(ctx, year, from, to)
}

// finishPurge deletes the raw jobs of the year and marks it purged.
func (s *RetentionService) finishPurge(ctx context.Context, year int, from, to time.Time) error {
	deleted, err := s.purge(ctx, from, to)
	if err != nil {
		return err
	}
	if err := s.store.MarkYearPurged(ctx, year); err != nil {
		return err
	}
	log.Printf("retention: year %d purged (%d jobs deleted in this run)", year, deleted)
	return nil
}

// checkSummary stops the run before any delete if the summary does not match
// the raw data. A mismatch means a job arrived while the roll-up ran, or the
// roll-up missed something; running again repairs the first, and the error
// shows the second.
func (s *RetentionService) checkSummary(ctx context.Context, year int, from, to time.Time) error {
	raw, err := s.store.RawVacancyTotal(ctx, from, to)
	if err != nil {
		return err
	}
	stored, err := s.store.GetYearTotal(ctx, year)
	if err != nil {
		return err
	}
	if stored == nil || stored.VacancyCount != raw {
		var got int64
		if stored != nil {
			got = stored.VacancyCount
		}
		return fmt.Errorf("summary total %d does not match the %d raw vacancies; nothing was deleted", got, raw)
	}
	if err := s.store.VerifyYearSummary(ctx, year); err != nil {
		return fmt.Errorf("%w; nothing was deleted", err)
	}
	return nil
}

// purge deletes the raw jobs of [from, to) in batches until none are left.
func (s *RetentionService) purge(ctx context.Context, from, to time.Time) (int64, error) {
	var deleted int64
	for {
		if err := ctx.Err(); err != nil {
			return deleted, err
		}
		n, err := s.store.PurgeBatch(ctx, from, to, s.cfg.BatchSize)
		if err != nil {
			return deleted, err
		}
		if n == 0 {
			return deleted, nil
		}
		deleted += n
	}
}
