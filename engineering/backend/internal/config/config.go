package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
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

// Validate reports missing or malformed MCP settings. PublicBaseURL is put
// into the WWW-Authenticate header that tells clients where to find the
// resource metadata, so a wrong value sends every client to a URL that does
// not exist and login fails with no clear cause. A path prefix is allowed
// (a proxy may serve the MCP server under one); a trailing slash is not,
// because the metadata path is appended to the value as it is.
func (c MCPConfig) Validate() error {
	var missing []string
	if c.ResourceID == "" {
		missing = append(missing, "MCP_RESOURCE_ID")
	}
	if c.PublicBaseURL == "" {
		missing = append(missing, "MCP_PUBLIC_BASE_URL")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}

	// None of the errors below repeat the value. It is published to every
	// client in the WWW-Authenticate header, so it may carry credentials
	// (user:password@host) that must not also end up in a startup log.
	u, err := url.Parse(c.PublicBaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return errors.New("MCP_PUBLIC_BASE_URL must be an absolute http(s) URL such as https://mcp.example.com")
	}
	if u.User != nil {
		return errors.New("MCP_PUBLIC_BASE_URL must not contain a username or password: the value is published to every client")
	}
	// Look for the literal delimiters, not u.RawQuery / u.Fragment: url.Parse
	// keeps a lone "?" as an empty query and drops a lone "#" altogether, so
	// "https://host?" and "https://host#" would pass those checks. The metadata
	// path is appended to the original string, so either would send clients to
	// the wrong path. Percent-encoded characters in a path prefix (%3F) are
	// still fine, because they contain no literal "?" or "#".
	if strings.ContainsAny(c.PublicBaseURL, "?#") {
		return errors.New("MCP_PUBLIC_BASE_URL must not contain a query or fragment")
	}
	if strings.HasSuffix(c.PublicBaseURL, "/") {
		return errors.New("MCP_PUBLIC_BASE_URL must not end with a slash")
	}
	// The value is written inside a quoted parameter of the WWW-Authenticate
	// header, so a literal quote, backslash or space in the path would break
	// the header. url.Parse accepts all of them. Requiring the value to survive
	// a parse and re-serialise unchanged rejects them, and also non-ASCII
	// characters and an uppercase scheme, while keeping valid prefixes such as
	// /team%22west (the encoded form of a quote).
	if u.String() != c.PublicBaseURL {
		return errors.New("MCP_PUBLIC_BASE_URL must be in canonical form: a lowercase http or https scheme, and special characters such as quotes, spaces and non-ASCII letters percent-encoded (for example %22 for a quote)")
	}
	return nil
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
