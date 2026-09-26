package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	eventsAdapter "github.com/discord-subscriptions/deploy-svc/internal/adapters/events"
	"github.com/discord-subscriptions/deploy-svc/internal/adapters/handlers"
	"github.com/discord-subscriptions/deploy-svc/internal/adapters/k8s"
	"github.com/discord-subscriptions/deploy-svc/internal/adapters/repositories"
	"github.com/discord-subscriptions/deploy-svc/internal/config"
	"github.com/discord-subscriptions/deploy-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/auth"
	"github.com/discord-subscriptions/shared/database"
	"github.com/discord-subscriptions/shared/health"
	"github.com/discord-subscriptions/shared/logger"
	"github.com/discord-subscriptions/shared/messaging"
	"github.com/discord-subscriptions/shared/telemetry"
	"github.com/discord-subscriptions/shared/vault"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	log := logger.New(logger.Config{
		ServiceName: "deploy-svc",
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		JSONFormat:  cfg.Environment == "production",
	})

	log.Info("starting deploy-svc", slog.Int("port", cfg.Port), slog.String("env", cfg.Environment))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Connect to PostgreSQL
	db, err := database.Connect(ctx, "pgx", cfg.DB)
	if err != nil {
		log.Error("failed connecting to deploy_db", slog.String("error", err.Error()))
	} else {
		defer db.Close()
		log.Info("connected to deploy_db successfully")
	}

	// 2. Connect to RabbitMQ
	var rmqClient *messaging.RabbitMQClient
	rmqClient, err = messaging.NewRabbitMQClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Warn("rabbitmq not available, messaging will be disabled or simulated", slog.String("error", err.Error()))
	} else {
		defer rmqClient.Close()
		log.Info("connected to rabbitmq successfully")
	}

	// 3. Connect to Vault
	vaultClient := vault.NewVaultClient(cfg.VaultAddr, cfg.VaultToken)

	// 4. Initialize Kubernetes Orchestrator
	k8sClient, err := k8s.NewClient(cfg.KubeconfigPath, log)
	if err != nil {
		log.Error("failed initializing kubernetes client", slog.String("error", err.Error()))
	}

	// 5. Wire Clean Architecture layers
	depRepo := repositories.NewPostgresDeploymentRepository(db)
	poolRepo := repositories.NewPostgresTokenPoolRepository(db)
	deployService := services.NewDeploymentService(depRepo, poolRepo, k8sClient, vaultClient, rmqClient, log)

	// 6. Start RabbitMQ Event Consumer if connected
	if rmqClient != nil {
		consumer := eventsAdapter.NewEventConsumer(deployService, rmqClient, log)
		if err := consumer.Start(ctx); err != nil {
			log.Error("failed starting event consumer", slog.String("error", err.Error()))
		}
	}

	// 7. Health Checker with PostgreSQL, RabbitMQ, and Vault probes
	healthChecker := health.NewChecker("deploy-svc", "1.0.0")
	healthChecker.AddDatabaseCheck("postgres", db)
	if rmqClient != nil {
		healthChecker.AddRabbitMQCheck("rabbitmq", rmqClient)
	}
	if vaultClient != nil {
		healthChecker.AddVaultCheck("vault", vaultClient)
	}

	// 8. Setup HTTP server
	authValidator := auth.NewValidator(cfg.JWTSecret)
	mux := http.NewServeMux()
	handler := handlers.NewHTTPHandler(deployService, healthChecker, log, authValidator)
	handler.RegisterRoutes(mux)
	mux.Handle("GET /metrics", telemetry.Handler())

	telemetryMiddleware := telemetry.HTTPMiddleware("deploy-svc")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      telemetryMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("deploy-svc HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-shutdownChan
	log.Info("shutting down deploy-svc gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	} else {
		log.Info("deploy-svc exited cleanly")
	}
}
