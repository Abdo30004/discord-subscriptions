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

	"github.com/discord-subscriptions/billing-svc/internal/adapters/handlers"
	"github.com/discord-subscriptions/billing-svc/internal/adapters/paypal"
	"github.com/discord-subscriptions/billing-svc/internal/adapters/repositories"
	"github.com/discord-subscriptions/billing-svc/internal/config"
	"github.com/discord-subscriptions/billing-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/auth"
	"github.com/discord-subscriptions/shared/database"
	"github.com/discord-subscriptions/shared/health"
	"github.com/discord-subscriptions/shared/logger"
	"github.com/discord-subscriptions/shared/messaging"
	"github.com/discord-subscriptions/shared/telemetry"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	log := logger.New(logger.Config{
		ServiceName: "billing-svc",
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		JSONFormat:  cfg.Environment == "production",
	})

	log.Info("starting billing-svc", slog.Int("port", cfg.Port), slog.String("env", cfg.Environment))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Connect to PostgreSQL
	db, err := database.Connect(ctx, "pgx", cfg.DB)
	if err != nil {
		log.Error("failed connecting to billing_db", slog.String("error", err.Error()))
	} else {
		defer db.Close()
		log.Info("connected to billing_db successfully")
	}

	// 2. Connect to RabbitMQ
	var rmqClient *messaging.RabbitMQClient
	rmqClient, err = messaging.NewRabbitMQClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Warn("rabbitmq not available, billing events will not be broadcast", slog.String("error", err.Error()))
	} else {
		defer rmqClient.Close()
		if rmqClient.IsConnected() {
			log.Info("connected to rabbitmq successfully")
		} else {
			log.Warn("rabbitmq initial connection pending, background reconnect active")
		}
	}

	// 3. Initialize PayPal Adapter
	paypalClient := paypal.NewClient(cfg.PayPalClientID, cfg.PayPalClientSecret, cfg.PayPalWebhookID, cfg.PayPalMode, log)

	// 4. Wire Clean Architecture layers
	subRepo := repositories.NewPostgresSubscriptionRepository(db)
	promoRepo := repositories.NewPostgresPromoRepository(db)
	vouchRepo := repositories.NewPostgresVoucherRepository(db)
	billingService := services.NewBillingService(subRepo, promoRepo, vouchRepo, paypalClient, rmqClient, log)

	// 5. Health Checker with PostgreSQL and RabbitMQ probes
	healthChecker := health.NewChecker("billing-svc", "1.0.0")
	healthChecker.AddDatabaseCheck("postgres", db)
	if rmqClient != nil {
		healthChecker.AddRabbitMQCheck("rabbitmq", rmqClient)
	}

	// 6. Setup HTTP REST API
	authValidator := auth.NewValidator(cfg.JWTSecret)
	mux := http.NewServeMux()
	handler := handlers.NewHTTPHandler(billingService, healthChecker, log, authValidator)
	handler.RegisterRoutes(mux)
	mux.Handle("GET /metrics", telemetry.Handler())

	telemetryMiddleware := telemetry.HTTPMiddleware("billing-svc")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      telemetryMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("billing-svc HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-shutdownChan
	log.Info("shutting down billing-svc gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	} else {
		log.Info("billing-svc exited cleanly")
	}
}
