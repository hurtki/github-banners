package config

import (
	"os"
)

type Config struct {
	LogLevel  string
	LogFormat string

	ServiceSecret      string
	BannersStoragePath string
}

func Load() *Config {
	return &Config{
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		ServiceSecret:      getEnv("SERVICES_SECRET_KEY", "1234"),
		BannersStoragePath: getEnv("BANNERS_STORAGE_PATH", "/var/www/banners/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
