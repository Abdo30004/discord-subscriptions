package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
)

type PostgresVoucherRepository struct {
	db *sql.DB
}

func NewPostgresVoucherRepository(db *sql.DB) *PostgresVoucherRepository {
	return &PostgresVoucherRepository{db: db}
}

func (r *PostgresVoucherRepository) Create(ctx context.Context, voucher *domain.VoucherCode) error {
	query := `
		INSERT INTO voucher_codes (
			id, code, plan_id, bot_type, duration_days, is_dedicated, is_redeemed, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		voucher.ID, domain.NormalizeVoucher(voucher.Code), voucher.PlanID,
		voucher.BotType, voucher.DurationDays, voucher.IsDedicated,
		voucher.IsRedeemed, voucher.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed creating voucher: %w", err)
	}
	return nil
}

func (r *PostgresVoucherRepository) GetByCode(ctx context.Context, code string) (*domain.VoucherCode, error) {
	query := `
		SELECT id, code, plan_id, bot_type, duration_days, is_dedicated, 
		       is_redeemed, redeemed_by_user_id, redeemed_by_guild_id, redeemed_at, created_at
		FROM voucher_codes
		WHERE code = $1
	`
	var v domain.VoucherCode
	var redeemedUser, redeemedGuild sql.NullString
	var redeemedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, domain.NormalizeVoucher(code)).Scan(
		&v.ID, &v.Code, &v.PlanID, &v.BotType, &v.DurationDays,
		&v.IsDedicated, &v.IsRedeemed, &redeemedUser, &redeemedGuild,
		&redeemedAt, &v.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound
		}
		return nil, err
	}

	if redeemedUser.Valid {
		v.RedeemedByUserID = redeemedUser.String
	}
	if redeemedGuild.Valid {
		v.RedeemedByGuildID = redeemedGuild.String
	}
	if redeemedAt.Valid {
		v.RedeemedAt = &redeemedAt.Time
	}
	return &v, nil
}

func (r *PostgresVoucherRepository) Update(ctx context.Context, voucher *domain.VoucherCode) error {
	query := `
		UPDATE voucher_codes
		SET is_redeemed = $1, redeemed_by_user_id = $2, redeemed_by_guild_id = $3, redeemed_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(
		ctx, query,
		voucher.IsRedeemed, voucher.RedeemedByUserID, voucher.RedeemedByGuildID,
		voucher.RedeemedAt, voucher.ID,
	)
	return err
}
