package services_test

import (
	"context"
	"testing"

	"github.com/discord-subscriptions/catalog-svc/internal/core/domain"
	"github.com/discord-subscriptions/catalog-svc/internal/core/services"
)

type mockBotRepo struct {
	bots []domain.BotTemplate
}

func (m *mockBotRepo) GetAll(ctx context.Context) ([]domain.BotTemplate, error) {
	return m.bots, nil
}
func (m *mockBotRepo) GetByID(ctx context.Context, id string) (*domain.BotTemplate, error) {
	for _, b := range m.bots {
		if b.ID == id {
			return &b, nil
		}
	}
	return nil, nil
}
func (m *mockBotRepo) GetBySlug(ctx context.Context, slug string) (*domain.BotTemplate, error) {
	for _, b := range m.bots {
		if b.Slug == slug {
			return &b, nil
		}
	}
	return nil, nil
}
func (m *mockBotRepo) Create(ctx context.Context, bot *domain.BotTemplate) error {
	m.bots = append(m.bots, *bot)
	return nil
}

type mockPlanRepo struct {
	plans []domain.Plan
}

func (m *mockPlanRepo) GetByBotID(ctx context.Context, botID string) ([]domain.Plan, error) {
	var result []domain.Plan
	for _, p := range m.plans {
		if p.BotID == botID {
			result = append(result, p)
		}
	}
	return result, nil
}
func (m *mockPlanRepo) GetByID(ctx context.Context, id string) (*domain.Plan, error) {
	for _, p := range m.plans {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, nil
}
func (m *mockPlanRepo) Create(ctx context.Context, plan *domain.Plan) error {
	m.plans = append(m.plans, *plan)
	return nil
}

func TestCatalogService_ListBots(t *testing.T) {
	botRepo := &mockBotRepo{
		bots: []domain.BotTemplate{
			{ID: "bot-1", Slug: "music-bot", Name: "Music Bot", Category: domain.CategoryMusic},
		},
	}
	planRepo := &mockPlanRepo{
		plans: []domain.Plan{
			{ID: "plan-1", BotID: "bot-1", Name: "Pro Tier", PriceCents: 999, Interval: domain.IntervalMonthly},
		},
	}

	service := services.NewCatalogService(botRepo, planRepo)
	bots, err := service.ListBots(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(bots) != 1 {
		t.Fatalf("expected 1 bot, got %d", len(bots))
	}

	if len(bots[0].Plans) != 1 {
		t.Fatalf("expected 1 plan attached to bot, got %d", len(bots[0].Plans))
	}

	if bots[0].Plans[0].Name != "Pro Tier" {
		t.Fatalf("expected plan name 'Pro Tier', got '%s'", bots[0].Plans[0].Name)
	}
}
