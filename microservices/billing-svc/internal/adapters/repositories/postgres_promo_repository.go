package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
)

type PostgresPromoRepository struct {
	db *sql.DB
}

func NewPostgresPromoRepository(db *sql.DB) *PostgresPromoRepository {
	return &PostgresPromoRepository{db: db}
}

func (r *PostgresPromoRepository) Create(ctx context.Context, promo *domain.PromoCode) error {
	query := `
		INSERT INTO promo_codes (
			id, code, discount_type, discount_value, max_uses, current_uses, expires_at, is_active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		promo.ID, domain.NormalizeCode(promo.Code), string(promo.DiscountType),
		promo.DiscountValue, promo.MaxUses, promo.CurrentUses, promo.ExpiresAt,
		promo.IsActive, promo.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed creating promo code: %w", err)
	}
	return nil
}

func (r *PostgresPromoRepository) GetByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	query := `
		SELECT id, code, discount_type, discount_value, max_uses, current_uses, expires_at, is_active, created_at
		FROM promo_codes
		WHERE code = $1
	`
	var p domain.PromoCode
	var discType string
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, domain.NormalizeCode(code)).Scan(
		&p.ID, &p.Code, &discType, &p.DiscountValue, &p.MaxUses,
		&p.CurrentUses, &expiresAt, &p.IsActive, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound
		}
		return nil, err
	}

	p.DiscountType = domain.DiscountType(discType)
	if expiresAt.Valid {
		p.ExpiresAt = &expiresAt.Time
	}
	return &p, nil
}

func (r *PostgresPromoRepository) IncrementUsage(ctx context.Context, code string) error {
	query := `UPDATE promo_codes SET current_uses = current_uses + 1 WHERE code = $1`
	_, err := r.db.ExecContext(ctx, query, domain.NormalizeCode(code))
	return err
}
