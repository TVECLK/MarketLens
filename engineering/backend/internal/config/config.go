package config

import (
	"os"
	"strconv"
	"time"

	"marketlens-go-backend/internal/auth"
	"marketlens-go-backend/internal/ingestion"
	"marketlens-go-backend/internal/llm"
)

// DBConfig holds Postgres connection settings for ConnectDatabase.
type DBConfig struct {
	Host, User, Password, Name, Port string
	SSLMode, Timezone                string
	MaxOpenConns, MaxIdleConns       int
	ConnMaxLifetimeMinutes           int
	ConnMaxIdleTimeMinutes           int
}

// MCPConfig holds the MCP server's own settings.
type MCPConfig struct {
	ResourceID    string
	PublicBaseURL string
}

// ServerConfig holds the HTTP server's own settings. GinModeEnv is the raw
// GIN_MODE value (possibly empty) — resolving it to an actual gin mode is
// left to main.go, which already imports gin.
type ServerConfig struct {
	Port       string
	GinModeEnv string
}

// Config bundles every environment-driven setting the app uses, read once
// at startup and passed down explicitly from there.
type Config struct {
	DB        DBConfig
	Auth      auth.Config
	LLM       llm.Config
	Ingestion ingestion.PipelineConfig
	MCP       MCPConfig
	Server    ServerConfig
}

// LoadConfig reads every environment variable the app uses into one
// struct. It does not validate — each sub-config's own Validate() (or
// ConnectDatabase's existing check) is called explicitly by whichever
// binary actually needs that piece, since e.g. cmd/mcp needs DB but not
// Auth/LLM.
func LoadConfig() *Config {
	return &Config{
		DB: DBConfig{
			Host:                   os.Getenv("DB_HOST"),
			User:                   os.Getenv("DB_USER"),
			Password:               os.Getenv("DB_PASSWORD"),
			Name:                   os.Getenv("DB_NAME"),
			Port:                   os.Getenv("DB_PORT"),
			SSLMode:                envOrDefault("DB_SSLMODE", "disable"),
			Timezone:               envOrDefault("DB_TIMEZONE", "UTC"),
			MaxOpenConns:           envIntOrDefault("DB_MAX_OPEN_CONNS", 30),
			MaxIdleConns:           envIntOrDefault("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetimeMinutes: envIntOrDefault("DB_CONN_MAX_LIFETIME_MINUTES", 30),
			ConnMaxIdleTimeMinutes: envIntOrDefault("DB_CONN_MAX_IDLE_TIME_MINUTES", 5),
		},
		Auth: auth.Config{
			BaseURL:     os.Getenv("THUNDER_BASE_URL"),
			Issuer:      os.Getenv("THUNDER_ISSUER"),
			Audience:    os.Getenv("THUNDER_AUDIENCE"),
			InsecureTLS: os.Getenv("THUNDER_INSECURE_TLS") == "true",
		},
		LLM: llm.Config{
			APIKey: os.Getenv("LLM_API_KEY"),
			APIURL: os.Getenv("LLM_API_URL"),
			Model:  os.Getenv("LLM_MODEL"),
		},
		Ingestion: ingestion.PipelineConfig{
			DedupThreshold:   envFloatOrDefault("INGESTION_DEDUP_THRESHOLD", 0.65),
			MetadataCacheTTL: time.Duration(envIntOrDefault("INGESTION_METADATA_CACHE_TTL_MINUTES", 30)) * time.Minute,
		},
		MCP: MCPConfig{
			ResourceID:    os.Getenv("MCP_RESOURCE_ID"),
			PublicBaseURL: os.Getenv("MCP_PUBLIC_BASE_URL"),
		},
		Server: ServerConfig{
			Port:       envOrDefault("PORT", "8080"),
			GinModeEnv: os.Getenv("GIN_MODE"),
		},
	}
}

// envOrDefault reads a string env var, falling back to def if it's unset.
func envOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
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

// envFloatOrDefault reads a float env var, falling back to def if it's
// unset or not a valid float.
func envFloatOrDefault(key string, def float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return def
	}
	return n
}
