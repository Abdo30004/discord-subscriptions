package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
)

type PostgresSubscriptionRepository struct {
	db *sql.DB
}

func NewPostgresSubscriptionRepository(db *sql.DB) *PostgresSubscriptionRepository {
	return &PostgresSubscriptionRepository{db: db}
}

func (r *PostgresSubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	query := `
		INSERT INTO subscriptions (
			id, user_id, guild_id, plan_id, bot_type, instance_label, status, 
			provider, external_sub_id, is_dedicated, is_zero_setup, valid_until, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		sub.ID, sub.UserID, sub.GuildID, sub.PlanID, sub.BotType, sub.InstanceLabel,
		string(sub.Status), string(sub.Provider), sub.ExternalSubID,
		sub.IsDedicated, sub.IsZeroSetup, sub.ValidUntil, sub.CreatedAt, sub.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed creating subscription: %w", err)
	}
	return nil
}

func (r *PostgresSubscriptionRepository) GetByID(ctx context.Context, id string) (*domain.Subscription, error) {
	query := `
		SELECT id, user_id, guild_id, plan_id, bot_type, instance_label, status, 
		       provider, external_sub_id, is_dedicated, is_zero_setup, valid_until, created_at, updated_at
		FROM subscriptions
		WHERE id = $1
	`
	return r.scanSubscription(r.db.QueryRowContext(ctx, query, id))
}

func (r *PostgresSubscriptionRepository) GetActiveByGuildID(ctx context.Context, guildID string) (*domain.Subscription, error) {
	query := `
		SELECT id, user_id, guild_id, plan_id, bot_type, instance_label, status, 
		       provider, external_sub_id, is_dedicated, is_zero_setup, valid_until, created_at, updated_at
		FROM subscriptions
		WHERE guild_id = $1 AND status = 'active' AND valid_until > CURRENT_TIMESTAMP
		ORDER BY valid_until DESC
		LIMIT 1
	`
	return r.scanSubscription(r.db.QueryRowContext(ctx, query, guildID))
}

func (r *PostgresSubscriptionRepository) ListByGuildID(ctx context.Context, guildID string) ([]domain.Subscription, error) {
	query := `
		SELECT id, user_id, guild_id, plan_id, bot_type, instance_label, status, 
		       provider, external_sub_id, is_dedicated, is_zero_setup, valid_until, created_at, updated_at
		FROM subscriptions
		WHERE guild_id = $1 AND status = 'active' AND valid_until > CURRENT_TIMESTAMP
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, guildID)
	if err != nil {
		return nil, fmt.Errorf("failed listing subscriptions for guild: %w", err)
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var s domain.Subscription
		var status, provider string
		var extID sql.NullString

		err := rows.Scan(
			&s.ID, &s.UserID, &s.GuildID, &s.PlanID, &s.BotType, &s.InstanceLabel,
			&status, &provider, &extID, &s.IsDedicated, &s.IsZeroSetup, &s.ValidUntil,
			&s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning subscription row: %w", err)
		}
		s.Status = domain.SubscriptionStatus(status)
		s.Provider = domain.PaymentProvider(provider)
		if extID.Valid {
			s.ExternalSubID = extID.String
		}
		subs = append(subs, s)
	}

	return subs, nil
}

func (r *PostgresSubscriptionRepository) GetByExternalSubID(ctx context.Context, extID string) (*domain.Subscription, error) {
	query := `
		SELECT id, user_id, guild_id, plan_id, bot_type, instance_label, status, 
		       provider, external_sub_id, is_dedicated, is_zero_setup, valid_until, created_at, updated_at
		FROM subscriptions
		WHERE external_sub_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	return r.scanSubscription(r.db.QueryRowContext(ctx, query, extID))
}

func (r *PostgresSubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	query := `
		UPDATE subscriptions
		SET status = $1, valid_until = $2, updated_at = $3, external_sub_id = $4, instance_label = $5
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query, string(sub.Status), sub.ValidUntil, sub.UpdatedAt, sub.ExternalSubID, sub.InstanceLabel, sub.ID)
	if err != nil {
		return fmt.Errorf("failed updating subscription: %w", err)
	}
	return nil
}

func (r *PostgresSubscriptionRepository) scanSubscription(row *sql.Row) (*domain.Subscription, error) {
	var s domain.Subscription
	var status, provider string
	var extID sql.NullString

	err := row.Scan(
		&s.ID, &s.UserID, &s.GuildID, &s.PlanID, &s.BotType, &s.InstanceLabel,
		&status, &provider, &extID, &s.IsDedicated, &s.IsZeroSetup, &s.ValidUntil,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound
		}
		return nil, err
	}

	s.Status = domain.SubscriptionStatus(status)
	s.Provider = domain.PaymentProvider(provider)
	if extID.Valid {
		s.ExternalSubID = extID.String
	}
	return &s, nil
}
