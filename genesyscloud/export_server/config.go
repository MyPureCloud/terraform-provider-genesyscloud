package export_server

import (
	"os"
	"strconv"
	"time"
)

// ServerConfig holds the configuration for the export server
type ServerConfig struct {
	Port              int           `json:"port"`
	ExportBaseDir     string        `json:"export_base_dir"`
	MaxConcurrentJobs int           `json:"max_concurrent_jobs"`
	JobTimeout        time.Duration `json:"job_timeout"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	MaxJobAge         time.Duration `json:"max_job_age"`
}

// DefaultConfig returns the default server configuration
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		Port:              getEnvAsInt("EXPORT_SERVER_PORT", 8080),
		ExportBaseDir:     getEnvAsString("EXPORT_SERVER_BASE_DIR", "./exports"),
		MaxConcurrentJobs: getEnvAsInt("EXPORT_SERVER_MAX_JOBS", 5),
		JobTimeout:        getEnvAsDuration("EXPORT_SERVER_JOB_TIMEOUT", 30*time.Minute),
		CleanupInterval:   getEnvAsDuration("EXPORT_SERVER_CLEANUP_INTERVAL", 1*time.Hour),
		MaxJobAge:         getEnvAsDuration("EXPORT_SERVER_MAX_JOB_AGE", 24*time.Hour),
	}
}

// getEnvAsString gets an environment variable as a string, with a default value
func getEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer, with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsDuration gets an environment variable as a duration, with a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
