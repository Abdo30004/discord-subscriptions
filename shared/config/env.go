package config

import (
	"os"
	"strconv"
	"time"
)

// GetString retrieves an environment variable or returns the fallback value.
func GetString(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

// GetInt retrieves an environment variable as an integer or returns the fallback value.
func GetInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

// GetBool retrieves an environment variable as a boolean or returns the fallback value.
func GetBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}

// GetDuration retrieves an environment variable as a time.Duration or returns the fallback value.
func GetDuration(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}
