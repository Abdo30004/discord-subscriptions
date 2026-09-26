package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
)

type PostgresDeploymentRepository struct {
	db *sql.DB
}

// NewPostgresDeploymentRepository creates a persistence adapter for deployments in PostgreSQL.
func NewPostgresDeploymentRepository(db *sql.DB) *PostgresDeploymentRepository {
	return &PostgresDeploymentRepository{db: db}
}

// Create inserts a new deployment record.
func (r *PostgresDeploymentRepository) Create(ctx context.Context, dep *domain.Deployment) error {
	query := `
		INSERT INTO deployments (
			id, subscription_id, user_id, guild_id, bot_type, instance_label,
			k8s_namespace, k8s_deployment_name, image_name, image_tag, 
			status, is_zero_setup, client_id, error_message, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		dep.ID, dep.SubscriptionID, dep.UserID, dep.GuildID, dep.BotType, dep.InstanceLabel,
		dep.K8sNamespace, dep.K8sDeploymentName, dep.ImageName, dep.ImageTag,
		string(dep.Status), dep.IsZeroSetup, dep.ClientID, dep.ErrorMessage, dep.CreatedAt, dep.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed inserting deployment: %w", err)
	}
	return nil
}

// GetByID finds a deployment by primary key.
func (r *PostgresDeploymentRepository) GetByID(ctx context.Context, id string) (*domain.Deployment, error) {
	query := `
		SELECT id, subscription_id, user_id, guild_id, bot_type, instance_label,
		       k8s_namespace, k8s_deployment_name, image_name, image_tag, 
		       status, is_zero_setup, client_id, error_message, created_at, updated_at
		FROM deployments
		WHERE id = $1
	`
	return r.scanSingle(r.db.QueryRowContext(ctx, query, id))
}

// GetByGuildID finds the active deployment for a given Discord guild.
func (r *PostgresDeploymentRepository) GetByGuildID(ctx context.Context, guildID string) (*domain.Deployment, error) {
	query := `
		SELECT id, subscription_id, user_id, guild_id, bot_type, instance_label,
		       k8s_namespace, k8s_deployment_name, image_name, image_tag, 
		       status, is_zero_setup, client_id, error_message, created_at, updated_at
		FROM deployments
		WHERE guild_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	return r.scanSingle(r.db.QueryRowContext(ctx, query, guildID))
}

// ListByGuildID returns all deployments for a given Discord guild.
func (r *PostgresDeploymentRepository) ListByGuildID(ctx context.Context, guildID string) ([]domain.Deployment, error) {
	query := `
		SELECT id, subscription_id, user_id, guild_id, bot_type, instance_label,
		       k8s_namespace, k8s_deployment_name, image_name, image_tag, 
		       status, is_zero_setup, client_id, error_message, created_at, updated_at
		FROM deployments
		WHERE guild_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, guildID)
	if err != nil {
		return nil, fmt.Errorf("failed querying guild deployments: %w", err)
	}
	defer rows.Close()

	var deps []domain.Deployment
	for rows.Next() {
		var d domain.Deployment
		var status string
		var clientID, errMsg sql.NullString
		err := rows.Scan(
			&d.ID, &d.SubscriptionID, &d.UserID, &d.GuildID, &d.BotType, &d.InstanceLabel,
			&d.K8sNamespace, &d.K8sDeploymentName, &d.ImageName, &d.ImageTag,
			&status, &d.IsZeroSetup, &clientID, &errMsg, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning deployment row: %w", err)
		}
		d.Status = domain.DeploymentStatus(status)
		if clientID.Valid {
			d.ClientID = clientID.String
		}
		if errMsg.Valid {
			d.ErrorMessage = errMsg.String
		}
		deps = append(deps, d)
	}

	return deps, nil
}

// GetBySubscriptionID finds the deployment attached to a specific subscription.
func (r *PostgresDeploymentRepository) GetBySubscriptionID(ctx context.Context, subID string) (*domain.Deployment, error) {
	query := `
		SELECT id, subscription_id, user_id, guild_id, bot_type, instance_label,
		       k8s_namespace, k8s_deployment_name, image_name, image_tag, 
		       status, is_zero_setup, client_id, error_message, created_at, updated_at
		FROM deployments
		WHERE subscription_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	return r.scanSingle(r.db.QueryRowContext(ctx, query, subID))
}

// Update modifies deployment state, status, or error messages.
func (r *PostgresDeploymentRepository) Update(ctx context.Context, dep *domain.Deployment) error {
	query := `
		UPDATE deployments
		SET status = $1, error_message = $2, updated_at = $3, image_tag = $4
		WHERE id = $5
	`
	res, err := r.db.ExecContext(ctx, query, string(dep.Status), dep.ErrorMessage, dep.UpdatedAt, dep.ImageTag, dep.ID)
	if err != nil {
		return fmt.Errorf("failed updating deployment %s: %w", dep.ID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sharedErrors.ErrNotFound
	}

	return nil
}

// List returns a paginated list of deployments.
func (r *PostgresDeploymentRepository) List(ctx context.Context, limit, offset int) ([]domain.Deployment, error) {
	query := `
		SELECT id, subscription_id, user_id, guild_id, bot_type, instance_label,
		       k8s_namespace, k8s_deployment_name, image_name, image_tag, 
		       status, is_zero_setup, client_id, error_message, created_at, updated_at
		FROM deployments
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed querying deployments: %w", err)
	}
	defer rows.Close()

	var deps []domain.Deployment
	for rows.Next() {
		var d domain.Deployment
		var status string
		var clientID, errMsg sql.NullString
		err := rows.Scan(
			&d.ID, &d.SubscriptionID, &d.UserID, &d.GuildID, &d.BotType, &d.InstanceLabel,
			&d.K8sNamespace, &d.K8sDeploymentName, &d.ImageName, &d.ImageTag,
			&status, &d.IsZeroSetup, &clientID, &errMsg, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning deployment row: %w", err)
		}
		d.Status = domain.DeploymentStatus(status)
		if clientID.Valid {
			d.ClientID = clientID.String
		}
		if errMsg.Valid {
			d.ErrorMessage = errMsg.String
		}
		deps = append(deps, d)
	}

	return deps, nil
}

func (r *PostgresDeploymentRepository) scanSingle(row *sql.Row) (*domain.Deployment, error) {
	var d domain.Deployment
	var status string
	var clientID, errMsg sql.NullString
	err := row.Scan(
		&d.ID, &d.SubscriptionID, &d.UserID, &d.GuildID, &d.BotType, &d.InstanceLabel,
		&d.K8sNamespace, &d.K8sDeploymentName, &d.ImageName, &d.ImageTag,
		&status, &d.IsZeroSetup, &clientID, &errMsg, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed scanning deployment: %w", err)
	}

	d.Status = domain.DeploymentStatus(status)
	if clientID.Valid {
		d.ClientID = clientID.String
	}
	if errMsg.Valid {
		d.ErrorMessage = errMsg.String
	}
	return &d, nil
}
