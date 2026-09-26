package services

import (
	"context"
	"fmt"

	"github.com/discord-subscriptions/catalog-svc/internal/core/domain"
	"github.com/discord-subscriptions/catalog-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/errors"
)

type CatalogServiceImpl struct {
	botRepo  ports.BotRepository
	planRepo ports.PlanRepository
}

// NewCatalogService initializes the service with repository dependencies.
func NewCatalogService(botRepo ports.BotRepository, planRepo ports.PlanRepository) *CatalogServiceImpl {
	return &CatalogServiceImpl{
		botRepo:  botRepo,
		planRepo: planRepo,
	}
}

// ListBots returns all available bots along with their subscription plans.
func (s *CatalogServiceImpl) ListBots(ctx context.Context) ([]domain.BotTemplate, error) {
	bots, err := s.botRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve bots: %w", err)
	}

	for i := range bots {
		plans, err := s.planRepo.GetByBotID(ctx, bots[i].ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve plans for bot %s: %w", bots[i].ID, err)
		}
		bots[i].Plans = plans
	}

	return bots, nil
}

// GetBotDetails retrieves a specific bot by either UUID or Slug, including plans.
func (s *CatalogServiceImpl) GetBotDetails(ctx context.Context, idOrSlug string) (*domain.BotTemplate, error) {
	bot, err := s.botRepo.GetByID(ctx, idOrSlug)
	if err != nil && errors.ErrNotFound == err {
		// Try slug lookup
		bot, err = s.botRepo.GetBySlug(ctx, idOrSlug)
	}

	if err != nil {
		return nil, err
	}

	plans, err := s.planRepo.GetByBotID(ctx, bot.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve plans for bot %s: %w", bot.ID, err)
	}
	bot.Plans = plans

	return bot, nil
}

// GetPlan retrieves a single plan by its ID.
func (s *CatalogServiceImpl) GetPlan(ctx context.Context, planID string) (*domain.Plan, error) {
	return s.planRepo.GetByID(ctx, planID)
}
