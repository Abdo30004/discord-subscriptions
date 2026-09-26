package ports

import (
	"context"

	"github.com/discord-subscriptions/catalog-svc/internal/core/domain"
)

// BotRepository handles database persistence operations for bot templates.
type BotRepository interface {
	GetAll(ctx context.Context) ([]domain.BotTemplate, error)
	GetByID(ctx context.Context, id string) (*domain.BotTemplate, error)
	GetBySlug(ctx context.Context, slug string) (*domain.BotTemplate, error)
	Create(ctx context.Context, bot *domain.BotTemplate) error
}

// PlanRepository handles persistence operations for subscription plans.
type PlanRepository interface {
	GetByBotID(ctx context.Context, botID string) ([]domain.Plan, error)
	GetByID(ctx context.Context, id string) (*domain.Plan, error)
	Create(ctx context.Context, plan *domain.Plan) error
}

// CatalogService defines the core business logic interface for the store catalog.
type CatalogService interface {
	ListBots(ctx context.Context) ([]domain.BotTemplate, error)
	GetBotDetails(ctx context.Context, idOrSlug string) (*domain.BotTemplate, error)
	GetPlan(ctx context.Context, planID string) (*domain.Plan, error)
}
