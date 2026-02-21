package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	LogLevel  string
	LogFormat string

	ServiceSecret string
	StorageBaseURL string
}

func Load() *Config {
	return &Config{
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		ServiceSecret: getEnv("SERVICES_SECRET_KEY", "1234"),
		StorageBaseURL: getEnv("STORAGE_BASE_URL", "http://storage/"),
	}
}

// helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if dur, err := time.ParseDuration(value); err == nil {
			return dur
		}
	}
	return defaultValue
}
