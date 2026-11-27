// Command retention keeps the database from growing for ever. It rolls every
// calendar year that has left the raw-data window (the last RETENTION_RAW_YEARS
// years) up into the stats_year_* summary tables, and then deletes that year's
// raw job rows. It is meant to run as a Kubernetes CronJob, about once a year.
//
// The delete step is off unless PURGE_ENABLED=true. It deletes for good and the
// job takes no backup, so take a database backup before turning it on. It is
// safe to run again at any time: summaries are replaced, purged years are
// skipped, and a run that was killed half way simply finishes the deletes. Only
// one run works at a time.
//
// Settings, besides the DB_* ones the other binaries use:
//
//	RETENTION_RAW_YEARS   years of raw data to keep (default 3)
//	RETENTION_TZ          time zone of year boundaries (default Asia/Colombo)
//	RETENTION_BATCH_SIZE  jobs per delete transaction (default 5000)
//	PURGE_ENABLED         "true" to delete (default off)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	// Embeds the time zone database: the runtime image is a bare Alpine with none.
	_ "time/tzdata"

	"marketlens-go-backend/internal/config"
	"marketlens-go-backend/internal/repositories"
	"marketlens-go-backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("retention: %v", err)
	}
}

func run() error {
	cfg := config.LoadConfig()
	if err := cfg.Retention.Validate(); err != nil {
		return fmt.Errorf("misconfigured: %w", err)
	}
	loc, err := cfg.Retention.Location()
	if err != nil {
		return err
	}

	db, err := config.ConnectDatabase(cfg.DB)
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}
	store := repositories.NewStore(db)

	// SIGTERM (a pod stop) or Ctrl-C ends the run between delete batches.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	release, acquired, err := store.TryRetentionLock(ctx)
	if err != nil {
		return err
	}
	if !acquired {
		log.Println("retention: another run holds the lock, exiting")
		return nil
	}
	defer release()

	retentionCfg := service.RetentionConfig{
		RawYears:     cfg.Retention.RawYears,
		Location:     loc,
		BatchSize:    cfg.Retention.BatchSize,
		PurgeEnabled: cfg.Retention.PurgeEnabled,
	}
	svc, err := service.NewRetentionService(store, retentionCfg)
	if err != nil {
		return err
	}
	if err := svc.Run(ctx); err != nil {
		return err
	}
	log.Println("retention: done")
	return nil
}
