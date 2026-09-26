package config

import (
	"fmt"
	"time"

	"github.com/discord-subscriptions/shared/config"
	"github.com/discord-subscriptions/shared/database"
)

type Config struct {
	Port         int
	Environment  string
	LogLevel     string
	PollInterval time.Duration
	WorkerCount  int
	RabbitMQURL  string
	DB           database.Config
}

func Load() Config {
	rabbitUser := config.GetString("RABBITMQ_USER", "guest")
	rabbitPass := config.GetString("RABBITMQ_PASS", "guest")
	rabbitHost := config.GetString("RABBITMQ_HOST", "localhost")
	rabbitPort := config.GetInt("RABBITMQ_PORT", 5672)

	return Config{
		Port:         config.GetInt("MONITOR_SVC_PORT", 8084),
		Environment:  config.GetString("ENV", "development"),
		LogLevel:     config.GetString("LOG_LEVEL", "info"),
		PollInterval: config.GetDuration("MONITOR_POLL_INTERVAL", 15*time.Second),
		WorkerCount:  config.GetInt("MONITOR_WORKERS", 5),
		RabbitMQURL:  fmt.Sprintf("amqp://%s:%s@%s:%d/", rabbitUser, rabbitPass, rabbitHost, rabbitPort),
		DB: database.Config{
			Host:     config.GetString("POSTGRES_HOST", "localhost"),
			Port:     config.GetInt("POSTGRES_PORT", 5432),
			User:     config.GetString("POSTGRES_USER", "postgres"),
			Password: config.GetString("POSTGRES_PASSWORD", "postgres"),
			Database: config.GetString("MONITOR_DB_NAME", "monitor_db"),
			SSLMode:  config.GetString("POSTGRES_SSLMODE", "disable"),
		},
	}
}
