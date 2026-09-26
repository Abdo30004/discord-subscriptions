package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/catalog-svc/internal/core/domain"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
)

type PostgresPlanRepository struct {
	db *sql.DB
}

// NewPostgresPlanRepository creates a new repository instance for subscription plans.
func NewPostgresPlanRepository(db *sql.DB) *PostgresPlanRepository {
	return &PostgresPlanRepository{db: db}
}

// GetByBotID retrieves all pricing tiers and plans associated with a bot.
func (r *PostgresPlanRepository) GetByBotID(ctx context.Context, botID string) ([]domain.Plan, error) {
	query := `
		SELECT id, bot_id, name, description, interval, price_cents, currency, is_dedicated, features, created_at, updated_at
		FROM subscription_plans
		WHERE bot_id = $1
		ORDER BY price_cents ASC
	`
	rows, err := r.db.QueryContext(ctx, query, botID)
	if err != nil {
		return nil, fmt.Errorf("failed querying subscription_plans: %w", err)
	}
	defer rows.Close()

	var plans []domain.Plan
	for rows.Next() {
		var p domain.Plan
		var rawFeatures []byte
		var interval string

		err := rows.Scan(
			&p.ID, &p.BotID, &p.Name, &p.Description, &interval,
			&p.PriceCents, &p.Currency, &p.IsDedicated, &rawFeatures,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning plan row: %w", err)
		}

		p.Interval = domain.BillingInterval(interval)
		if len(rawFeatures) > 0 {
			if err := json.Unmarshal(rawFeatures, &p.Features); err != nil {
				p.Features = []string{}
			}
		}

		plans = append(plans, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return plans, nil
}

// GetByID retrieves a single plan by its ID.
func (r *PostgresPlanRepository) GetByID(ctx context.Context, id string) (*domain.Plan, error) {
	query := `
		SELECT id, bot_id, name, description, interval, price_cents, currency, is_dedicated, features, created_at, updated_at
		FROM subscription_plans
		WHERE id = $1
	`
	var p domain.Plan
	var rawFeatures []byte
	var interval string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.BotID, &p.Name, &p.Description, &interval,
		&p.PriceCents, &p.Currency, &p.IsDedicated, &rawFeatures,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed scanning plan: %w", err)
	}

	p.Interval = domain.BillingInterval(interval)
	if len(rawFeatures) > 0 {
		_ = json.Unmarshal(rawFeatures, &p.Features)
	}

	return &p, nil
}

// Create persists a new subscription plan tier.
func (r *PostgresPlanRepository) Create(ctx context.Context, plan *domain.Plan) error {
	query := `
		INSERT INTO subscription_plans (id, bot_id, name, description, interval, price_cents, currency, is_dedicated, features)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return fmt.Errorf("failed marshaling features json: %w", err)
	}

	_, err = r.db.ExecContext(
		ctx, query,
		plan.ID, plan.BotID, plan.Name, plan.Description, string(plan.Interval),
		plan.PriceCents, plan.Currency, plan.IsDedicated, featuresJSON,
	)
	if err != nil {
		return fmt.Errorf("failed inserting subscription plan: %w", err)
	}

	return nil
}
