package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectDatabase opens a connection to Postgres, retrying until it's
// reachable, and returns the configured *gorm.DB. Callers own the result —
// there is no package-level global, since config.DB previously had exactly
// one consumer (the binary's main.go).
func ConnectDatabase() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	required := map[string]string{
		"DB_HOST":     dbHost,
		"DB_USER":     dbUser,
		"DB_PASSWORD": dbPassword,
		"DB_NAME":     dbName,
		"DB_PORT":     dbPort,
	}
	for name, val := range required {
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", name)
		}
	}

	sslMode := os.Getenv("DB_SSLMODE")
	if sslMode == "" {
		sslMode = "disable"
	}

	timezone := os.Getenv("DB_TIMEZONE")
	if timezone == "" {
		timezone = "UTC"
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, sslMode, timezone,
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

	sqlDB.SetMaxOpenConns(envIntOrDefault("DB_MAX_OPEN_CONNS", 30))
	sqlDB.SetMaxIdleConns(envIntOrDefault("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetConnMaxLifetime(time.Duration(envIntOrDefault("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(envIntOrDefault("DB_CONN_MAX_IDLE_TIME_MINUTES", 5)) * time.Minute)

	log.Println("Database connected successfully!")
	return database, nil
}

// envIntOrDefault reads an integer env var, falling back to def if it's
// unset or not a valid integer.
func envIntOrDefault(key string, def int) int {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return def
	}
	return n
}
