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
			access_token, refresh_token, token_expires_at, is_admin, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
		user.IsAdmin, user.CreatedAt, user.UpdatedAt,
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
		       is_admin, admin_promoted_at, admin_promoted_by,
		       access_token, refresh_token, token_expires_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var u domain.User
	var globalName, avatar, email, accToken, refToken, adminPromotedBy sql.NullString
	var tokenExpires, adminPromotedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&u.ID, &u.Username, &globalName, &avatar, &email,
		&u.IsAdmin, &adminPromotedAt, &adminPromotedBy,
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
	if adminPromotedBy.Valid {
		u.AdminPromotedBy = adminPromotedBy.String
	}
	if adminPromotedAt.Valid {
		t := adminPromotedAt.Time
		u.AdminPromotedAt = &t
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

// ListAdmins returns all appointed administrators in the database.
func (r *PostgresUserRepository) ListAdmins(ctx context.Context) ([]domain.User, error) {
	query := `
		SELECT id, username, global_name, avatar, email, 
		       is_admin, admin_promoted_at, admin_promoted_by,
		       created_at, updated_at
		FROM users
		WHERE is_admin = true
		ORDER BY admin_promoted_at DESC NULLS LAST, created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying admins: %w", err)
	}
	defer rows.Close()

	var admins []domain.User
	for rows.Next() {
		var u domain.User
		var globalName, avatar, email, adminPromotedBy sql.NullString
		var adminPromotedAt sql.NullTime

		if err := rows.Scan(
			&u.ID, &u.Username, &globalName, &avatar, &email,
			&u.IsAdmin, &adminPromotedAt, &adminPromotedBy,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning admin row: %w", err)
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
		if adminPromotedBy.Valid {
			u.AdminPromotedBy = adminPromotedBy.String
		}
		if adminPromotedAt.Valid {
			t := adminPromotedAt.Time
			u.AdminPromotedAt = &t
		}
		admins = append(admins, u)
	}

	return admins, nil
}

// SearchUsers searches registered users by Discord ID or Username.
func (r *PostgresUserRepository) SearchUsers(ctx context.Context, q string, limit int) ([]domain.User, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	searchPattern := "%" + q + "%"
	query := `
		SELECT id, username, global_name, avatar, email, 
		       is_admin, admin_promoted_at, admin_promoted_by,
		       created_at, updated_at
		FROM users
		WHERE id ILIKE $1 OR username ILIKE $1 OR global_name ILIKE $1
		ORDER BY is_admin DESC, username ASC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed searching users: %w", err)
	}
	defer rows.Close()

	var results []domain.User
	for rows.Next() {
		var u domain.User
		var globalName, avatar, email, adminPromotedBy sql.NullString
		var adminPromotedAt sql.NullTime

		if err := rows.Scan(
			&u.ID, &u.Username, &globalName, &avatar, &email,
			&u.IsAdmin, &adminPromotedAt, &adminPromotedBy,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed scanning search row: %w", err)
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
		if adminPromotedBy.Valid {
			u.AdminPromotedBy = adminPromotedBy.String
		}
		if adminPromotedAt.Valid {
			t := adminPromotedAt.Time
			u.AdminPromotedAt = &t
		}
		results = append(results, u)
	}

	return results, nil
}

// SetAdmin updates a user's is_admin status in the database.
func (r *PostgresUserRepository) SetAdmin(ctx context.Context, discordID string, isAdmin bool, promotedBy string) error {
	query := `
		UPDATE users
		SET is_admin = $2,
		    admin_promoted_at = CASE WHEN $2 = true THEN CURRENT_TIMESTAMP ELSE NULL END,
		    admin_promoted_by = CASE WHEN $2 = true THEN $3 ELSE NULL END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, query, discordID, isAdmin, promotedBy)
	if err != nil {
		return fmt.Errorf("failed updating admin status: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sharedErrors.ErrNotFound
	}
	return nil
}
