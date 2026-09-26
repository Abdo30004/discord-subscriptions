package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
	"github.com/discord-subscriptions/monitor-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/events"
	"github.com/discord-subscriptions/shared/messaging"
	"github.com/google/uuid"
)

type MonitorServiceImpl struct {
	repo      ports.TargetRepository
	pinger    ports.Pinger
	publisher messaging.Publisher
	logger    *slog.Logger
}

// NewMonitorService creates an instance of MonitorServiceImpl.
func NewMonitorService(
	repo ports.TargetRepository,
	pinger ports.Pinger,
	publisher messaging.Publisher,
	logger *slog.Logger,
) *MonitorServiceImpl {
	if logger == nil {
		logger = slog.Default()
	}
	return &MonitorServiceImpl{
		repo:      repo,
		pinger:    pinger,
		publisher: publisher,
		logger:    logger,
	}
}

// RegisterTarget registers or updates an active bot health probe target.
func (s *MonitorServiceImpl) RegisterTarget(
	ctx context.Context,
	botID, guildID, instanceLabel, healthURL string,
	pollInterval int,
) (*domain.MonitoringTarget, error) {
	if pollInterval <= 0 {
		pollInterval = 15 // Default 15s interval
	}
	if instanceLabel == "" {
		instanceLabel = "Default"
	}

	target := &domain.MonitoringTarget{
		ID:                  fmt.Sprintf("tgt-%s", uuid.New().String()[:8]),
		BotID:               botID,
		GuildID:             guildID,
		InstanceLabel:       instanceLabel,
		HealthURL:           healthURL,
		PollIntervalSec:     pollInterval,
		IsActive:            true,
		CurrentStatus:       events.BotStatusOnline,
		ConsecutiveFailures: 0,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	if err := target.Validate(); err != nil {
		return nil, fmt.Errorf("invalid target: %w", err)
	}

	if err := s.repo.CreateTarget(ctx, target); err != nil {
		return nil, fmt.Errorf("failed saving target: %w", err)
	}

	s.logger.Info("registered monitoring target",
		slog.String("target_id", target.ID),
		slog.String("guild_id", target.GuildID),
		slog.String("health_url", target.HealthURL),
	)

	return target, nil
}

// DeactivateTarget stops active polling for a given target.
func (s *MonitorServiceImpl) DeactivateTarget(ctx context.Context, targetID string) error {
	target, err := s.repo.GetTargetByID(ctx, targetID)
	if err != nil {
		return err
	}

	target.IsActive = false
	target.UpdatedAt = time.Now().UTC()
	return s.repo.UpdateTarget(ctx, target)
}

// GetGuildStatus returns the target and its recent probe logs for a Discord server.
func (s *MonitorServiceImpl) GetGuildStatus(ctx context.Context, guildID string) (*domain.MonitoringTarget, []domain.CheckResult, error) {
	target, err := s.repo.GetTargetByGuildID(ctx, guildID)
	if err != nil {
		return nil, nil, err
	}

	logs, err := s.repo.GetRecentLogs(ctx, target.ID, 10)
	if err != nil {
		s.logger.Warn("failed fetching recent logs", slog.String("target_id", target.ID), slog.String("error", err.Error()))
	}

	return target, logs, nil
}

// GetGuildTargets returns all monitoring targets for a Discord server.
func (s *MonitorServiceImpl) GetGuildTargets(ctx context.Context, guildID string) ([]domain.MonitoringTarget, error) {
	return s.repo.ListTargetsByGuildID(ctx, guildID)
}

// CheckTarget runs an active health probe against a target and emits events on state transition.
func (s *MonitorServiceImpl) CheckTarget(ctx context.Context, target *domain.MonitoringTarget) error {
	result := s.pinger.Ping(ctx, target)

	// 1. Record health check log
	if err := s.repo.SaveCheckLog(ctx, &result); err != nil {
		s.logger.Error("failed saving check log", slog.String("target_id", target.ID), slog.String("error", err.Error()))
	}

	// 2. Evaluate state change
	statusChanged, oldStatus := target.ApplyCheckResult(result)

	// 3. Update target state in database
	if err := s.repo.UpdateTarget(ctx, target); err != nil {
		s.logger.Error("failed updating target state", slog.String("target_id", target.ID), slog.String("error", err.Error()))
	}

	// 4. Publish event if status transitioned
	if statusChanged {
		s.logger.Warn("bot status changed",
			slog.String("bot_id", target.BotID),
			slog.String("guild_id", target.GuildID),
			slog.String("old_status", string(oldStatus)),
			slog.String("new_status", string(target.CurrentStatus)),
		)

		if s.publisher != nil {
			evt := events.NewBotStatusChangedEvent(
				target.BotID,
				target.GuildID,
				oldStatus,
				target.CurrentStatus,
				result.ErrorMessage,
				result.LatencyMs,
			)
			_ = s.publisher.Publish(ctx, "discord.events", "bot.status.changed", evt)
		}
	}

	return nil
}

// ListTargets returns all active monitored targets.
func (s *MonitorServiceImpl) ListTargets(ctx context.Context) ([]domain.MonitoringTarget, error) {
	return s.repo.GetActiveTargets(ctx)
}
