package config

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDatabase opens a connection to Postgres, retrying until it's
// reachable, and returns the configured *gorm.DB. Callers own the result —
// there is no package-level global, since config.DB previously had exactly
// one consumer (the binary's main.go).
func ConnectDatabase(cfg DBConfig) (*gorm.DB, error) {
	required := map[string]string{
		"DB_HOST":     cfg.Host,
		"DB_USER":     cfg.User,
		"DB_PASSWORD": cfg.Password,
		"DB_NAME":     cfg.Name,
		"DB_PORT":     cfg.Port,
	}
	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", name)
		}
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, cfg.SSLMode, cfg.Timezone,
	)

	const maxAttempts = 10
	const retryDelay = 3 * time.Second

	var database *gorm.DB
	var sqlDB *sql.DB
	var err error

	// gorm.Open alone doesn't verify connectivity (the underlying sql.Open
	// is lazy and rarely errors on its own), so each attempt also pings
	// before declaring success - that's the only step that actually proves
	// the database is reachable.
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, err = database.DB()
			if err == nil {
				if err = sqlDB.Ping(); err == nil {
					break
				}
			}
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", attempt, maxAttempts, err)
		time.Sleep(retryDelay)
	}
	if err != nil {
		return nil, fmt.Errorf("could not connect to database after %d attempts: %w", maxAttempts, err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTimeMinutes) * time.Minute)

	log.Println("Database connected successfully!")
	return database, nil
}
