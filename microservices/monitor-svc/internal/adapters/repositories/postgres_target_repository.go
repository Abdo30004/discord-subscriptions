package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
	"github.com/discord-subscriptions/shared/events"
)

type PostgresTargetRepository struct {
	db *sql.DB
}

// NewPostgresTargetRepository creates a repository backed by PostgreSQL.
func NewPostgresTargetRepository(db *sql.DB) *PostgresTargetRepository {
	return &PostgresTargetRepository{db: db}
}

// CreateTarget inserts or updates a monitoring target based on bot_id conflict.
func (r *PostgresTargetRepository) CreateTarget(ctx context.Context, target *domain.MonitoringTarget) error {
	if r == nil || r.db == nil {
		return errors.New("database connection is not available")
	}
	query := `
		INSERT INTO monitoring_targets (
			id, bot_id, guild_id, instance_label, health_url, poll_interval_sec, 
			is_active, current_status, consecutive_failures, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (bot_id) DO UPDATE SET
			health_url = EXCLUDED.health_url,
			instance_label = EXCLUDED.instance_label,
			guild_id = EXCLUDED.guild_id,
			is_active = true,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(
		ctx, query,
		target.ID, target.BotID, target.GuildID, target.InstanceLabel, target.HealthURL,
		target.PollIntervalSec, target.IsActive, string(target.CurrentStatus),
		target.ConsecutiveFailures, target.CreatedAt, target.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed creating or updating target: %w", err)
	}
	return nil
}

// GetTargetByID finds a target by its ID.
func (r *PostgresTargetRepository) GetTargetByID(ctx context.Context, id string) (*domain.MonitoringTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	query := `
		SELECT id, bot_id, guild_id, instance_label, health_url, poll_interval_sec, 
		       is_active, current_status, consecutive_failures, last_checked_at, 
		       created_at, updated_at
		FROM monitoring_targets
		WHERE id = $1
	`
	return r.scanSingle(r.db.QueryRowContext(ctx, query, id))
}

// GetTargetByGuildID finds the first target for a given Discord guild.
func (r *PostgresTargetRepository) GetTargetByGuildID(ctx context.Context, guildID string) (*domain.MonitoringTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	query := `
		SELECT id, bot_id, guild_id, instance_label, health_url, poll_interval_sec, 
		       is_active, current_status, consecutive_failures, last_checked_at, 
		       created_at, updated_at
		FROM monitoring_targets
		WHERE guild_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	return r.scanSingle(r.db.QueryRowContext(ctx, query, guildID))
}

// ListTargetsByGuildID finds all targets for a given Discord guild.
func (r *PostgresTargetRepository) ListTargetsByGuildID(ctx context.Context, guildID string) ([]domain.MonitoringTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	query := `
		SELECT id, bot_id, guild_id, instance_label, health_url, poll_interval_sec, 
		       is_active, current_status, consecutive_failures, last_checked_at, 
		       created_at, updated_at
		FROM monitoring_targets
		WHERE guild_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, guildID)
	if err != nil {
		return nil, fmt.Errorf("failed querying guild targets: %w", err)
	}
	defer rows.Close()

	var targets []domain.MonitoringTarget
	for rows.Next() {
		t, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, *t)
	}
	return targets, nil
}

// GetActiveTargets returns all active targets configured for polling.
func (r *PostgresTargetRepository) GetActiveTargets(ctx context.Context) ([]domain.MonitoringTarget, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	query := `
		SELECT id, bot_id, guild_id, instance_label, health_url, poll_interval_sec, 
		       is_active, current_status, consecutive_failures, last_checked_at, 
		       created_at, updated_at
		FROM monitoring_targets
		WHERE is_active = true
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying active targets: %w", err)
	}
	defer rows.Close()

	var targets []domain.MonitoringTarget
	for rows.Next() {
		t, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, *t)
	}
	return targets, nil
}

// UpdateTarget updates status, failures, and check timestamps.
func (r *PostgresTargetRepository) UpdateTarget(ctx context.Context, target *domain.MonitoringTarget) error {
	if r == nil || r.db == nil {
		return errors.New("database connection is not available")
	}
	query := `
		UPDATE monitoring_targets
		SET current_status = $1, consecutive_failures = $2, 
		    last_checked_at = $3, updated_at = $4, is_active = $5
		WHERE id = $6
	`
	_, err := r.db.ExecContext(
		ctx, query,
		string(target.CurrentStatus), target.ConsecutiveFailures,
		target.LastCheckedAt, target.UpdatedAt, target.IsActive, target.ID,
	)
	if err != nil {
		return fmt.Errorf("failed updating target: %w", err)
	}
	return nil
}

// SaveCheckLog appends a probe result into the health_logs table.
func (r *PostgresTargetRepository) SaveCheckLog(ctx context.Context, log *domain.CheckResult) error {
	if r == nil || r.db == nil {
		return errors.New("database connection is not available")
	}
	query := `
		INSERT INTO health_logs (
			target_id, status, status_code, latency_ms, 
			error_message, discord_ping_ms, memory_usage_mb, checked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		log.TargetID, string(log.Status), log.StatusCode, log.LatencyMs,
		log.ErrorMessage, log.DiscordPingMs, log.MemoryUsageMb, log.CheckedAt,
	)
	if err != nil {
		return fmt.Errorf("failed saving health log: %w", err)
	}
	return nil
}

// GetRecentLogs returns the latest probe history for a target.
func (r *PostgresTargetRepository) GetRecentLogs(ctx context.Context, targetID string, limit int) ([]domain.CheckResult, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("database connection is not available")
	}
	query := `
		SELECT target_id, status, status_code, latency_ms, 
		       error_message, discord_ping_ms, memory_usage_mb, checked_at
		FROM health_logs
		WHERE target_id = $1
		ORDER BY checked_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, targetID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed querying health logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.CheckResult
	for rows.Next() {
		var res domain.CheckResult
		var status string
		var errMsg sql.NullString

		err := rows.Scan(
			&res.TargetID, &status, &res.StatusCode, &res.LatencyMs,
			&errMsg, &res.DiscordPingMs, &res.MemoryUsageMb, &res.CheckedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning log row: %w", err)
		}
		res.Status = events.BotStatus(status)
		if errMsg.Valid {
			res.ErrorMessage = errMsg.String
		}
		logs = append(logs, res)
	}
	return logs, nil
}

func (r *PostgresTargetRepository) scanSingle(row *sql.Row) (*domain.MonitoringTarget, error) {
	var t domain.MonitoringTarget
	var status string
	var lastChecked sql.NullTime

	err := row.Scan(
		&t.ID, &t.BotID, &t.GuildID, &t.InstanceLabel, &t.HealthURL, &t.PollIntervalSec,
		&t.IsActive, &status, &t.ConsecutiveFailures, &lastChecked,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("target not found")
		}
		return nil, err
	}

	t.CurrentStatus = events.BotStatus(status)
	if lastChecked.Valid {
		t.LastCheckedAt = &lastChecked.Time
	}
	return &t, nil
}

func (r *PostgresTargetRepository) scanRow(rows *sql.Rows) (*domain.MonitoringTarget, error) {
	var t domain.MonitoringTarget
	var status string
	var lastChecked sql.NullTime

	err := rows.Scan(
		&t.ID, &t.BotID, &t.GuildID, &t.InstanceLabel, &t.HealthURL, &t.PollIntervalSec,
		&t.IsActive, &status, &t.ConsecutiveFailures, &lastChecked,
		&t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.CurrentStatus = events.BotStatus(status)
	if lastChecked.Valid {
		t.LastCheckedAt = &lastChecked.Time
	}
	return &t, nil
}
