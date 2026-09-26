package config

import (
	"github.com/discord-subscriptions/shared/config"
	"github.com/discord-subscriptions/shared/database"
)

type Config struct {
	Port        int
	Environment string
	LogLevel    string
	DB          database.Config
}

func Load() Config {
	return Config{
		Port:        config.GetInt("CATALOG_SVC_PORT", 8081),
		Environment: config.GetString("ENV", "development"),
		LogLevel:    config.GetString("LOG_LEVEL", "info"),
		DB: database.Config{
			Host:     config.GetString("POSTGRES_HOST", "localhost"),
			Port:     config.GetInt("POSTGRES_PORT", 5432),
			User:     config.GetString("POSTGRES_USER", "postgres"),
			Password: config.GetString("POSTGRES_PASSWORD", "postgres"),
			Database: config.GetString("CATALOG_DB_NAME", "catalog_db"),
			SSLMode:  config.GetString("POSTGRES_SSLMODE", "disable"),
		},
	}
}
