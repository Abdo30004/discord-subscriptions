package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// Upsert inserts or updates a user profile on login.
func (r *PostgresUserRepository) Upsert(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, username, global_name, avatar, email, 
			access_token, refresh_token, token_expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			global_name = EXCLUDED.global_name,
			avatar = EXCLUDED.avatar,
			email = EXCLUDED.email,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			token_expires_at = EXCLUDED.token_expires_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.ExecContext(
		ctx, query,
		user.ID, user.Username, user.GlobalName, user.Avatar, user.Email,
		user.AccessToken, user.RefreshToken, user.TokenExpiresAt,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed upserting user: %w", err)
	}
	return nil
}

// GetByID finds a user by Discord Snowflake ID.
func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, global_name, avatar, email, 
		       access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	var globalName, avatar, email, accToken, refToken sql.NullString
	var tokenExpires sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &globalName, &avatar, &email,
		&accToken, &refToken, &tokenExpires, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound
		}
		return nil, err
	}

	if globalName.Valid {
		u.GlobalName = globalName.String
	}
	if avatar.Valid {
		u.Avatar = avatar.String
	}
	if email.Valid {
		u.Email = email.String
	}
	if accToken.Valid {
		u.AccessToken = accToken.String
	}
	if refToken.Valid {
		u.RefreshToken = refToken.String
	}
	if tokenExpires.Valid {
		u.TokenExpiresAt = tokenExpires.Time
	}

	return &u, nil
}
