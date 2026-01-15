package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port string
	CORSOrigin string

	GithubToken string

	RateLimitRPS int

	CacheTTL time.Duration

	RequestTimeout time.Duration

	LogLevel string
	LogFormat string
}

func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		CORSOrigin:    getEnv("CORS_ORIGIN", "*"),
		GithubToken:   os.Getenv("GITHUB_TOKEN"),
		RateLimitRPS:  getEnvAsInt("RATE_LIMIT_RPS", 10),
		CacheTTL:      getEnvAsDuration("CACHE_TTL", 5*time.Minute),
		RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 10*time.Second),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		LogFormat:     getEnv("LOG_FORMAT", "json"),
	}
}

//helper functions
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
