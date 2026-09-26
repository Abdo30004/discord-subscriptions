package config

import (
	"fmt"

	"github.com/discord-subscriptions/shared/config"
	"github.com/discord-subscriptions/shared/database"
)

type Config struct {
	Port           int
	Environment    string
	LogLevel       string
	KubeconfigPath string
	RabbitMQURL    string
	VaultAddr      string
	VaultToken     string
	JWTSecret      string
	DB             database.Config
}

func Load() Config {
	rabbitUser := config.GetString("RABBITMQ_USER", "guest")
	rabbitPass := config.GetString("RABBITMQ_PASS", "guest")
	rabbitHost := config.GetString("RABBITMQ_HOST", "localhost")
	rabbitPort := config.GetInt("RABBITMQ_PORT", 5672)

	return Config{
		Port:           config.GetInt("DEPLOY_SVC_PORT", 8083),
		Environment:    config.GetString("ENV", "development"),
		LogLevel:       config.GetString("LOG_LEVEL", "info"),
		KubeconfigPath: config.GetString("KUBECONFIG", ""),
		RabbitMQURL:    fmt.Sprintf("amqp://%s:%s@%s:%d/", rabbitUser, rabbitPass, rabbitHost, rabbitPort),
		VaultAddr:      config.GetString("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:     config.GetString("VAULT_DEV_ROOT_TOKEN", "root"),
		JWTSecret:      config.GetString("JWT_SECRET", "super-secret-development-jwt-key-replace-in-production"),
		DB: database.Config{
			Host:     config.GetString("POSTGRES_HOST", "localhost"),
			Port:     config.GetInt("POSTGRES_PORT", 5432),
			User:     config.GetString("POSTGRES_USER", "deploy_user"),
			Password: config.GetString("POSTGRES_PASSWORD", "deploy_pass"),
			Database: config.GetString("DEPLOY_DB_NAME", "deploy_db"),
			SSLMode:  config.GetString("POSTGRES_SSLMODE", "disable"),
		},
	}
}
