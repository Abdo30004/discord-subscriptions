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

	"github.com/discord-subscriptions/catalog-svc/internal/adapters/handlers"
	"github.com/discord-subscriptions/catalog-svc/internal/adapters/repositories"
	"github.com/discord-subscriptions/catalog-svc/internal/config"
	"github.com/discord-subscriptions/catalog-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/database"
	"github.com/discord-subscriptions/shared/health"
	"github.com/discord-subscriptions/shared/logger"
	"github.com/discord-subscriptions/shared/telemetry"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	log := logger.New(logger.Config{
		ServiceName: "catalog-svc",
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		JSONFormat:  cfg.Environment == "production",
	})

	log.Info("starting catalog-svc", slog.Int("port", cfg.Port), slog.String("env", cfg.Environment))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to PostgreSQL database
	db, err := database.Connect(ctx, "pgx", cfg.DB)
	if err != nil {
		log.Error("failed to connect to database", slog.String("error", err.Error()))
		// Continue even if DB is down locally so service can be inspected/tested
	} else {
		defer db.Close()
		log.Info("connected to catalog_db successfully")
	}

	// Wire up Clean Architecture layers
	botRepo := repositories.NewPostgresBotRepository(db)
	planRepo := repositories.NewPostgresPlanRepository(db)
	catalogService := services.NewCatalogService(botRepo, planRepo)

	// Health checker with PostgreSQL probe
	healthChecker := health.NewChecker("catalog-svc", "1.0.0")
	healthChecker.AddDatabaseCheck("postgres", db)

	mux := http.NewServeMux()
	handler := handlers.NewHTTPHandler(catalogService, healthChecker, log)
	handler.RegisterRoutes(mux)
	mux.Handle("GET /metrics", telemetry.Handler())

	telemetryMiddleware := telemetry.HTTPMiddleware("catalog-svc")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      telemetryMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-shutdownChan
	log.Info("shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	} else {
		log.Info("server exited cleanly")
	}
}
