package config

import (
	"fmt"

	"github.com/discord-subscriptions/shared/config"
	"github.com/discord-subscriptions/shared/database"
)

type Config struct {
	Port               int
	Environment        string
	LogLevel           string
	RabbitMQURL        string
	PayPalClientID     string
	PayPalClientSecret string
	PayPalWebhookID    string
	PayPalMode         string
	JWTSecret          string
	DB                 database.Config
}

func Load() Config {
	rabbitUser := config.GetString("RABBITMQ_USER", "guest")
	rabbitPass := config.GetString("RABBITMQ_PASS", "guest")
	rabbitHost := config.GetString("RABBITMQ_HOST", "localhost")
	rabbitPort := config.GetInt("RABBITMQ_PORT", 5672)

	return Config{
		Port:               config.GetInt("BILLING_SVC_PORT", 8082),
		Environment:        config.GetString("ENV", "development"),
		LogLevel:           config.GetString("LOG_LEVEL", "info"),
		RabbitMQURL:        fmt.Sprintf("amqp://%s:%s@%s:%d/", rabbitUser, rabbitPass, rabbitHost, rabbitPort),
		PayPalClientID:     config.GetString("PAYPAL_CLIENT_ID", ""),
		PayPalClientSecret: config.GetString("PAYPAL_CLIENT_SECRET", ""),
		PayPalWebhookID:    config.GetString("PAYPAL_WEBHOOK_ID", ""),
		PayPalMode:         config.GetString("PAYPAL_MODE", "sandbox"),
		JWTSecret:          config.GetString("JWT_SECRET", "super-secret-development-jwt-key-replace-in-production"),
		DB: database.Config{
			Host:     config.GetString("POSTGRES_HOST", "localhost"),
			Port:     config.GetInt("POSTGRES_PORT", 5432),
			User:     config.GetString("POSTGRES_USER", "billing_user"),
			Password: config.GetString("POSTGRES_PASSWORD", "billing_pass"),
			Database: config.GetString("BILLING_DB_NAME", "billing_db"),
			SSLMode:  config.GetString("POSTGRES_SSLMODE", "disable"),
		},
	}
}
