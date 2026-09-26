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

	"github.com/discord-subscriptions/auth-svc/internal/adapters/discord"
	"github.com/discord-subscriptions/auth-svc/internal/adapters/handlers"
	jwtAdapter "github.com/discord-subscriptions/auth-svc/internal/adapters/jwt"
	"github.com/discord-subscriptions/auth-svc/internal/adapters/repositories"
	"github.com/discord-subscriptions/auth-svc/internal/config"
	"github.com/discord-subscriptions/auth-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/database"
	"github.com/discord-subscriptions/shared/health"
	"github.com/discord-subscriptions/shared/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.Load()

	log := logger.New(logger.Config{
		ServiceName: "auth-svc",
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		JSONFormat:  cfg.Environment == "production",
	})

	log.Info("starting auth-svc", slog.Int("port", cfg.Port), slog.String("env", cfg.Environment))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Connect to PostgreSQL (auth_db)
	db, err := database.Connect(ctx, "pgx", cfg.DB)
	if err != nil {
		log.Error("failed connecting to auth_db", slog.String("error", err.Error()))
	} else {
		defer db.Close()
		log.Info("connected to auth_db successfully")
	}

	// 2. Wire Clean Architecture layers
	userRepo := repositories.NewPostgresUserRepository(db)
	discordClient := discord.NewClient(cfg.DiscordClientID, cfg.DiscordClientSecret, log)
	tokenMgr := jwtAdapter.NewManager(cfg.JWTSecret, 7*24*time.Hour)
	authService := services.NewAuthService(userRepo, discordClient, tokenMgr, log)

	// 3. Health Checker with PostgreSQL ping probe
	healthChecker := health.NewChecker("auth-svc", "1.0.0")
	healthChecker.AddDatabaseCheck("postgres", db)

	// 4. Setup HTTP server
	mux := http.NewServeMux()
	handler := handlers.NewHTTPHandler(authService, healthChecker, log)
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
		log.Info("auth-svc HTTP server listening", slog.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}()

	<-shutdownChan
	log.Info("shutting down auth-svc gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", slog.String("error", err.Error()))
	} else {
		log.Info("auth-svc exited cleanly")
	}
}
