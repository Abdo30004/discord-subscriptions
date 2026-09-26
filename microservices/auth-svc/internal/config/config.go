package config

import (
	"strings"

	"github.com/discord-subscriptions/shared/config"
	"github.com/discord-subscriptions/shared/database"
)

type Config struct {
	Port                int
	Environment         string
	LogLevel            string
	DiscordClientID     string
	DiscordClientSecret string
	JWTSecret           string
	SuperAdminIDs       []string
	DB                  database.Config
}

func Load() Config {
	superAdminRaw := config.GetString("SUPER_ADMIN_DISCORD_IDS", "")
	if superAdminRaw == "" {
		superAdminRaw = config.GetString("SUPER_ADMIN_DISCORD_ID", "")
	}
	var superAdminIDs []string
	for _, id := range strings.Split(superAdminRaw, ",") {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			superAdminIDs = append(superAdminIDs, trimmed)
		}
	}

	return Config{
		Port:                config.GetInt("AUTH_SVC_PORT", 8080),
		Environment:         config.GetString("ENV", "development"),
		LogLevel:            config.GetString("LOG_LEVEL", "info"),
		DiscordClientID:     config.GetString("DISCORD_CLIENT_ID", ""),
		DiscordClientSecret: config.GetString("DISCORD_CLIENT_SECRET", ""),
		JWTSecret:           config.GetString("JWT_SECRET", "super-secret-development-jwt-key-replace-in-production"),
		SuperAdminIDs:       superAdminIDs,
		DB: database.Config{
			Host:     config.GetString("POSTGRES_HOST", "localhost"),
			Port:     config.GetInt("POSTGRES_PORT", 5432),
			User:     config.GetString("POSTGRES_USER", "postgres"),
			Password: config.GetString("POSTGRES_PASSWORD", "postgres"),
			Database: config.GetString("AUTH_DB_NAME", "auth_db"),
			SSLMode:  config.GetString("POSTGRES_SSLMODE", "disable"),
		},
	}
}
