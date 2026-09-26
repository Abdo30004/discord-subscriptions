package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
)

type PostgresTokenPoolRepository struct {
	db *sql.DB
}

func NewPostgresTokenPoolRepository(db *sql.DB) *PostgresTokenPoolRepository {
	return &PostgresTokenPoolRepository{db: db}
}

func (r *PostgresTokenPoolRepository) Add(ctx context.Context, entry *domain.BotTokenPoolEntry) error {
	query := `
		INSERT INTO token_pool (
			id, bot_type, client_id, token_vault_path, token_encrypted, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		entry.ID, entry.BotType, entry.ClientID,
		entry.TokenVaultPath, entry.TokenEncrypted, string(entry.Status),
		entry.CreatedAt, entry.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed inserting token to pool: %w", err)
	}
	return nil
}

func (r *PostgresTokenPoolRepository) ClaimToken(ctx context.Context, botType, guildID, subscriptionID string) (*domain.BotTokenPoolEntry, error) {
	query := `
		UPDATE token_pool
		SET status = 'assigned',
		    assigned_guild_id = $1,
		    assigned_subscription_id = $2,
		    assigned_at = CURRENT_TIMESTAMP,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id
			FROM token_pool
			WHERE bot_type = $3 AND status = 'available'
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, bot_type, client_id, token_vault_path, token_encrypted, status, 
		          assigned_guild_id, assigned_subscription_id, assigned_at, created_at, updated_at
	`

	var entry domain.BotTokenPoolEntry
	var status string
	var assignedGuild, assignedSub sql.NullString
	var assignedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, guildID, subscriptionID, botType).Scan(
		&entry.ID, &entry.BotType, &entry.ClientID, &entry.TokenVaultPath, &entry.TokenEncrypted,
		&status, &assignedGuild, &assignedSub, &assignedAt, &entry.CreatedAt, &entry.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNoTokenAvailable
		}
		return nil, fmt.Errorf("failed claiming token from pool: %w", err)
	}

	entry.Status = domain.TokenStatus(status)
	if assignedGuild.Valid {
		entry.AssignedGuildID = assignedGuild.String
	}
	if assignedSub.Valid {
		entry.AssignedSubscriptionID = assignedSub.String
	}
	if assignedAt.Valid {
		entry.AssignedAt = &assignedAt.Time
	}

	return &entry, nil
}

func (r *PostgresTokenPoolRepository) ReleaseToken(ctx context.Context, guildID string) error {
	query := `
		UPDATE token_pool
		SET status = 'quarantined',
		    updated_at = CURRENT_TIMESTAMP
		WHERE assigned_guild_id = $1 AND status = 'assigned'
	`
	_, err := r.db.ExecContext(ctx, query, guildID)
	if err != nil {
		return fmt.Errorf("failed releasing token for guild: %w", err)
	}
	return nil
}

func (r *PostgresTokenPoolRepository) GetAvailableCount(ctx context.Context, botType string) (int, error) {
	query := `SELECT COUNT(*) FROM token_pool WHERE bot_type = $1 AND status = 'available'`
	var count int
	err := r.db.QueryRowContext(ctx, query, botType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed getting available count: %w", err)
	}
	return count, nil
}

func (r *PostgresTokenPoolRepository) GetPoolStats(ctx context.Context) (map[string]map[string]int, error) {
	query := `
		SELECT bot_type, status, COUNT(*)
		FROM token_pool
		GROUP BY bot_type, status
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying pool stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]map[string]int)
	for rows.Next() {
		var botType, status string
		var count int
		if err := rows.Scan(&botType, &status, &count); err != nil {
			return nil, err
		}
		if _, ok := stats[botType]; !ok {
			stats[botType] = make(map[string]int)
		}
		stats[botType][status] = count
	}

	return stats, nil
}

func (r *PostgresTokenPoolRepository) List(ctx context.Context, botType string, status domain.TokenStatus, limit, offset int) ([]domain.BotTokenPoolEntry, error) {
	query := `
		SELECT id, bot_type, client_id, token_vault_path, status,
		       assigned_guild_id, assigned_subscription_id, assigned_at, created_at, updated_at
		FROM token_pool
		WHERE ($1 = '' OR bot_type = $1)
		  AND ($2 = '' OR status = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, botType, string(status), limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed listing pool tokens: %w", err)
	}
	defer rows.Close()

	var entries []domain.BotTokenPoolEntry
	for rows.Next() {
		var e domain.BotTokenPoolEntry
		var st string
		var assignedGuild, assignedSub sql.NullString
		var assignedAt sql.NullTime

		if err := rows.Scan(
			&e.ID, &e.BotType, &e.ClientID, &e.TokenVaultPath, &st,
			&assignedGuild, &assignedSub, &assignedAt, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, err
		}

		e.Status = domain.TokenStatus(st)
		if assignedGuild.Valid {
			e.AssignedGuildID = assignedGuild.String
		}
		if assignedSub.Valid {
			e.AssignedSubscriptionID = assignedSub.String
		}
		if assignedAt.Valid {
			e.AssignedAt = &assignedAt.Time
		}
		e.TokenMasked = "••••••••••••••••"
		entries = append(entries, e)
	}

	return entries, nil
}
