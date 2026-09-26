package ports

import (
	"context"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
)

// TargetRepository defines data persistence operations for monitoring targets and health logs.
type TargetRepository interface {
	CreateTarget(ctx context.Context, target *domain.MonitoringTarget) error
	GetTargetByID(ctx context.Context, id string) (*domain.MonitoringTarget, error)
	GetTargetByGuildID(ctx context.Context, guildID string) (*domain.MonitoringTarget, error)
	ListTargetsByGuildID(ctx context.Context, guildID string) ([]domain.MonitoringTarget, error)
	GetActiveTargets(ctx context.Context) ([]domain.MonitoringTarget, error)
	UpdateTarget(ctx context.Context, target *domain.MonitoringTarget) error
	SaveCheckLog(ctx context.Context, log *domain.CheckResult) error
	GetRecentLogs(ctx context.Context, targetID string, limit int) ([]domain.CheckResult, error)
}

// Pinger abstracts active HTTP health checking against a target.
type Pinger interface {
	Ping(ctx context.Context, target *domain.MonitoringTarget) domain.CheckResult
}

// MonitorService defines the application operations for target management and health checks.
type MonitorService interface {
	RegisterTarget(ctx context.Context, botID, guildID, instanceLabel, healthURL string, pollInterval int) (*domain.MonitoringTarget, error)
	DeactivateTarget(ctx context.Context, targetID string) error
	GetGuildStatus(ctx context.Context, guildID string) (*domain.MonitoringTarget, []domain.CheckResult, error)
	GetGuildTargets(ctx context.Context, guildID string) ([]domain.MonitoringTarget, error)
	CheckTarget(ctx context.Context, target *domain.MonitoringTarget) error
	ListTargets(ctx context.Context) ([]domain.MonitoringTarget, error)
}
