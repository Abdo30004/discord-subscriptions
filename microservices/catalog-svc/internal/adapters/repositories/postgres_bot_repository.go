package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/discord-subscriptions/catalog-svc/internal/core/domain"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
)

type PostgresBotRepository struct {
	db *sql.DB
}

// NewPostgresBotRepository creates a new repository instance backed by PostgreSQL.
func NewPostgresBotRepository(db *sql.DB) *PostgresBotRepository {
	return &PostgresBotRepository{db: db}
}

// GetAll fetches all bot templates from the database.
func (r *PostgresBotRepository) GetAll(ctx context.Context) ([]domain.BotTemplate, error) {
	query := `
		SELECT id, slug, name, description, category, docker_image, default_image_tag, 
		       supports_dedicated, supports_shared, created_at, updated_at
		FROM bot_templates
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying bot_templates: %w", err)
	}
	defer rows.Close()

	var bots []domain.BotTemplate
	for rows.Next() {
		var b domain.BotTemplate
		var category string
		err := rows.Scan(
			&b.ID, &b.Slug, &b.Name, &b.Description, &category,
			&b.DockerImage, &b.DefaultImageTag, &b.SupportsDedicated,
			&b.SupportsShared, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed scanning bot_template row: %w", err)
		}
		b.Category = domain.BotCategory(category)
		bots = append(bots, b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return bots, nil
}

// GetByID finds a bot template by its primary key UUID.
func (r *PostgresBotRepository) GetByID(ctx context.Context, id string) (*domain.BotTemplate, error) {
	query := `
		SELECT id, slug, name, description, category, docker_image, default_image_tag, 
		       supports_dedicated, supports_shared, created_at, updated_at
		FROM bot_templates
		WHERE id = $1
	`
	return r.scanSingle(r.db.QueryRowContext(ctx, query, id))
}

// GetBySlug finds a bot template by its unique slug.
func (r *PostgresBotRepository) GetBySlug(ctx context.Context, slug string) (*domain.BotTemplate, error) {
	query := `
		SELECT id, slug, name, description, category, docker_image, default_image_tag, 
		       supports_dedicated, supports_shared, created_at, updated_at
		FROM bot_templates
		WHERE slug = $1
	`
	return r.scanSingle(r.db.QueryRowContext(ctx, query, slug))
}

func (r *PostgresBotRepository) scanSingle(row *sql.Row) (*domain.BotTemplate, error) {
	var b domain.BotTemplate
	var category string
	err := row.Scan(
		&b.ID, &b.Slug, &b.Name, &b.Description, &category,
		&b.DockerImage, &b.DefaultImageTag, &b.SupportsDedicated,
		&b.SupportsShared, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.ErrNotFound
		}
		return nil, fmt.Errorf("failed scanning bot template: %w", err)
	}
	b.Category = domain.BotCategory(category)
	return &b, nil
}

// Create inserts a new bot template.
func (r *PostgresBotRepository) Create(ctx context.Context, bot *domain.BotTemplate) error {
	query := `
		INSERT INTO bot_templates (id, slug, name, description, category, docker_image, default_image_tag, supports_dedicated, supports_shared)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		bot.ID, bot.Slug, bot.Name, bot.Description, string(bot.Category),
		bot.DockerImage, bot.DefaultImageTag, bot.SupportsDedicated, bot.SupportsShared,
	)
	if err != nil {
		return fmt.Errorf("failed inserting bot template: %w", err)
	}
	return nil
}
