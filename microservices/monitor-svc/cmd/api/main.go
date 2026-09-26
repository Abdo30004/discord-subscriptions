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

	eventsAdapter "github.com/discord-subscriptions/monitor-svc/internal/adapters/events"
	"github.com/discord-subscriptions/monitor-svc/internal/adapters/handlers"
	"github.com/discord-subscriptions/monitor-svc/internal/adapters/pinger"
	"github.com/discord-subscriptions/monitor-svc/internal/adapters/repositories"
	"github.com/discord-subscriptions/monitor-svc/internal/config"
	"github.com/discord-subscriptions/monitor-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/database"
	"github.com/discord-subscriptions/shared/health"
	"github.com/discord-subscriptions/shared/logger"
	"github.com/discord-subscriptions/shared/messaging"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	log := logger.New(logger.Config{
		ServiceName: "monitor-svc",
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		JSONFormat:  cfg.Environment == "production",
	})

	log.Info("starting monitor-svc", slog.Int("port", cfg.Port), slog.String("env", cfg.Environment))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Connect to PostgreSQL
	db, err := database.Connect(ctx, "pgx", cfg.DB)
	if err != nil {
		log.Error("failed connecting to monitor_db", slog.String("error", err.Error()))
	} else {
		defer db.Close()
		log.Info("connected to monitor_db successfully")
	}

	// 2. Connect to RabbitMQ
	var rmqClient *messaging.RabbitMQClient
	rmqClient, err = messaging.NewRabbitMQClient(cfg.RabbitMQURL, log)
	if err != nil {
		log.Warn("rabbitmq not available, status change events will not be broadcast", slog.String("error", err.Error()))
	} else {
		defer rmqClient.Close()
		log.Info("connected to rabbitmq successfully")
	}

	// 3. Wire Clean Architecture layers
	targetRepo := repositories.NewPostgresTargetRepository(db)
	httpPinger := pinger.NewHTTPPinger()
	monitorService := services.NewMonitorService(targetRepo, httpPinger, rmqClient, log)

	// 4. Start Background Poller Engine
	poller := services.NewPollerEngine(monitorService, cfg.PollInterval, cfg.WorkerCount, log)
	poller.Start(ctx)

	// 5. Start RabbitMQ Event Consumer
	if rmqClient != nil {
		consumer := eventsAdapter.NewEventConsumer(monitorService, rmqClient, log)
		if err := consumer.Start(ctx); err != nil {
			log.Error("failed starting event consumer", slog.String("error", err.Error()))
		}
	}

	// 6. Health Checker with PostgreSQL and RabbitMQ probes
	healthChecker := health.NewChecker("monitor-svc", "1.0.0")
	healthChecker.AddDatabaseCheck("postgres", db)
	if rmqClient != nil {
		healthChecker.AddRabbitMQCheck("rabbitmq", rmqClient)
	}

	// 7. Setup HTTP server
	mux := http.NewServeMux()
	handler := handlers.NewHTTPHandler(monitorService, healthChecker, log)
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("monitor-svc HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-shutdownChan
	log.Info("shutting down monitor-svc gracefully...")

	// Stop background poller
	poller.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	} else {
		log.Info("monitor-svc exited cleanly")
	}
}
