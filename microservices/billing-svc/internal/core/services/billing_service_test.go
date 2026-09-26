package services_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
	"github.com/discord-subscriptions/billing-svc/internal/core/ports"
	"github.com/discord-subscriptions/billing-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/events"
)

type mockSubRepo struct {
	subs map[string]*domain.Subscription
}

func (m *mockSubRepo) Create(ctx context.Context, sub *domain.Subscription) error {
	m.subs[sub.ID] = sub
	return nil
}
func (m *mockSubRepo) GetByID(ctx context.Context, id string) (*domain.Subscription, error) {
	return m.subs[id], nil
}
func (m *mockSubRepo) GetActiveByGuildID(ctx context.Context, guildID string) (*domain.Subscription, error) {
	for _, s := range m.subs {
		if s.GuildID == guildID && s.IsActiveNow() {
			return s, nil
		}
	}
	return nil, nil
}
func (m *mockSubRepo) GetByExternalSubID(ctx context.Context, extID string) (*domain.Subscription, error) {
	for _, s := range m.subs {
		if s.ExternalSubID == extID {
			return s, nil
		}
	}
	return nil, nil
}
func (m *mockSubRepo) ListByGuildID(ctx context.Context, guildID string) ([]domain.Subscription, error) {
	var res []domain.Subscription
	for _, s := range m.subs {
		if s.GuildID == guildID {
			res = append(res, *s)
		}
	}
	return res, nil
}
func (m *mockSubRepo) Update(ctx context.Context, sub *domain.Subscription) error {
	m.subs[sub.ID] = sub
	return nil
}

type mockPromoRepo struct {
	promos map[string]*domain.PromoCode
}

func (m *mockPromoRepo) Create(ctx context.Context, promo *domain.PromoCode) error {
	m.promos[promo.Code] = promo
	return nil
}
func (m *mockPromoRepo) GetByCode(ctx context.Context, code string) (*domain.PromoCode, error) {
	p, ok := m.promos[code]
	if !ok {
		return nil, domain.ErrInvalidPromo
	}
	return p, nil
}
func (m *mockPromoRepo) IncrementUsage(ctx context.Context, code string) error {
	if p, ok := m.promos[code]; ok {
		p.CurrentUses++
	}
	return nil
}

type mockVoucherRepo struct {
	vouchers map[string]*domain.VoucherCode
}

func (m *mockVoucherRepo) Create(ctx context.Context, voucher *domain.VoucherCode) error {
	m.vouchers[voucher.Code] = voucher
	return nil
}
func (m *mockVoucherRepo) GetByCode(ctx context.Context, code string) (*domain.VoucherCode, error) {
	v, ok := m.vouchers[code]
	if !ok {
		return nil, domain.ErrInvalidVoucherCode
	}
	return v, nil
}
func (m *mockVoucherRepo) Update(ctx context.Context, voucher *domain.VoucherCode) error {
	m.vouchers[voucher.Code] = voucher
	return nil
}

type mockPayPal struct{}

func (m *mockPayPal) CreateSubscriptionOrder(ctx context.Context, planID, returnURL, cancelURL string) (string, string, error) {
	return "https://paypal.com/approval?id=sub123", "sub123", nil
}
func (m *mockPayPal) VerifyWebhookSignature(r *http.Request, webhookID string) bool {
	return true
}

type mockPublisher struct {
	published []events.Event
}

func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, event events.Event) error {
	m.published = append(m.published, event)
	return nil
}

func TestBillingService_RedeemVoucher(t *testing.T) {
	subRepo := &mockSubRepo{subs: make(map[string]*domain.Subscription)}
	promoRepo := &mockPromoRepo{promos: make(map[string]*domain.PromoCode)}
	vouchRepo := &mockVoucherRepo{
		vouchers: map[string]*domain.VoucherCode{
			"GIFT-PRO-30D": {
				ID:           "vouch-1",
				Code:         "GIFT-PRO-30D",
				PlanID:       "pro-plan",
				BotType:      "music",
				DurationDays: 30,
				IsDedicated:  true,
				IsRedeemed:   false,
			},
		},
	}
	pub := &mockPublisher{}

	svc := services.NewBillingService(subRepo, promoRepo, vouchRepo, &mockPayPal{}, pub, nil)

	sub, err := svc.RedeemVoucher(context.Background(), "GIFT-PRO-30D", "user-1", "guild-100")
	if err != nil {
		t.Fatalf("unexpected error redeeming voucher: %v", err)
	}

	if sub.Provider != domain.ProviderVoucher {
		t.Fatalf("expected provider gift_code, got %s", sub.Provider)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected 1 activated event published, got %d", len(pub.published))
	}

	if pub.published[0].GetType() != events.TypeSubscriptionActivated {
		t.Fatalf("expected TypeSubscriptionActivated, got %s", pub.published[0].GetType())
	}
}

func TestBillingService_AdminGrant(t *testing.T) {
	subRepo := &mockSubRepo{subs: make(map[string]*domain.Subscription)}
	pub := &mockPublisher{}

	svc := services.NewBillingService(subRepo, &mockPromoRepo{}, &mockVoucherRepo{}, &mockPayPal{}, pub, nil)

	sub, err := svc.AdminGrantSubscription(context.Background(), "admin-user", "guild-200", "moderation", "enterprise", "Mod Squad", 90, true, false)
	if err != nil {
		t.Fatalf("unexpected error granting subscription: %v", err)
	}

	if sub.Provider != domain.ProviderAdminGrant {
		t.Fatalf("expected provider admin_grant, got %s", sub.Provider)
	}

	if sub.InstanceLabel != "Mod Squad" {
		t.Fatalf("expected instance label 'Mod Squad', got %s", sub.InstanceLabel)
	}

	if sub.ValidUntil.Before(time.Now().UTC().Add(89 * 24 * time.Hour)) {
		t.Fatal("expected at least 89 days of validity")
	}
}

func TestBillingService_MultiBot_Labels(t *testing.T) {
	subRepo := &mockSubRepo{subs: make(map[string]*domain.Subscription)}
	promoRepo := &mockPromoRepo{
		promos: map[string]*domain.PromoCode{
			"FREE100": {
				Code:          "FREE100",
				DiscountType:  domain.DiscountPercentage,
				DiscountValue: 100,
				IsActive:      true,
			},
		},
	}
	pub := &mockPublisher{}

	svc := services.NewBillingService(subRepo, promoRepo, &mockVoucherRepo{}, &mockPayPal{}, pub, nil)

	// First checkout without label -> "Default"
	resp1, err := svc.InitiateCheckout(context.Background(), ports.CheckoutRequest{
		UserID:    "user-1",
		GuildID:   "guild-multi",
		PlanID:    "plan-music-pro",
		BotType:   "music",
		PromoCode: "FREE100",
	})
	if err != nil {
		t.Fatalf("checkout 1 failed: %v", err)
	}
	if resp1.Subscription.InstanceLabel != "Default" {
		t.Fatalf("expected first bot to have 'Default' label, got %s", resp1.Subscription.InstanceLabel)
	}

	// Second checkout without label -> auto-incremented to "Music #2"
	resp2, err := svc.InitiateCheckout(context.Background(), ports.CheckoutRequest{
		UserID:    "user-1",
		GuildID:   "guild-multi",
		PlanID:    "plan-music-pro",
		BotType:   "music",
		PromoCode: "FREE100",
	})
	if err != nil {
		t.Fatalf("checkout 2 failed: %v", err)
	}
	if resp2.Subscription.InstanceLabel != "Music #2" {
		t.Fatalf("expected second bot to have 'Music #2' label, got %s", resp2.Subscription.InstanceLabel)
	}

	// Third checkout with explicit custom label
	resp3, err := svc.InitiateCheckout(context.Background(), ports.CheckoutRequest{
		UserID:        "user-1",
		GuildID:       "guild-multi",
		PlanID:        "plan-moderation-pro",
		BotType:       "moderation",
		InstanceLabel: "VIP Moderation",
		PromoCode:     "FREE100",
	})
	if err != nil {
		t.Fatalf("checkout 3 failed: %v", err)
	}
	if resp3.Subscription.InstanceLabel != "VIP Moderation" {
		t.Fatalf("expected third bot to have 'VIP Moderation' label, got %s", resp3.Subscription.InstanceLabel)
	}

	// Test GetGuildSubscriptions returns all 3
	subs, err := svc.GetGuildSubscriptions(context.Background(), "guild-multi")
	if err != nil {
		t.Fatalf("GetGuildSubscriptions failed: %v", err)
	}
	if len(subs) != 3 {
		t.Fatalf("expected 3 subscriptions for guild, got %d", len(subs))
	}
}

func TestBillingService_InitiateCheckout_100PercentPromo(t *testing.T) {
	subRepo := &mockSubRepo{subs: make(map[string]*domain.Subscription)}
	promoRepo := &mockPromoRepo{
		promos: map[string]*domain.PromoCode{
			"FREE100": {
				Code:          "FREE100",
				DiscountType:  domain.DiscountPercentage,
				DiscountValue: 100,
				IsActive:      true,
			},
		},
	}
	pub := &mockPublisher{}

	svc := services.NewBillingService(subRepo, promoRepo, &mockVoucherRepo{}, &mockPayPal{}, pub, nil)

	resp, err := svc.InitiateCheckout(context.Background(), ports.CheckoutRequest{
		UserID:    "user-99",
		GuildID:   "guild-99",
		PlanID:    "plan-music-pro",
		BotType:   "music",
		PromoCode: "FREE100",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.IsFreeInstantActive {
		t.Fatal("expected 100% discount promo to activate instantly without going to PayPal")
	}

	if resp.FinalPriceCents != 0 {
		t.Fatalf("expected final price 0, got %d", resp.FinalPriceCents)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected 1 activation event published, got %d", len(pub.published))
	}
}

func TestBillingService_InitiateCheckout_ZeroSetup(t *testing.T) {
	subRepo := &mockSubRepo{subs: make(map[string]*domain.Subscription)}
	promoRepo := &mockPromoRepo{
		promos: map[string]*domain.PromoCode{
			"FREE100": {
				Code:          "FREE100",
				DiscountType:  domain.DiscountPercentage,
				DiscountValue: 100,
				IsActive:      true,
			},
		},
	}
	pub := &mockPublisher{}

	svc := services.NewBillingService(subRepo, promoRepo, &mockVoucherRepo{}, &mockPayPal{}, pub, nil)

	// Checkout with Zero-Setup
	resp, err := svc.InitiateCheckout(context.Background(), ports.CheckoutRequest{
		UserID:      "user-zero",
		GuildID:     "guild-zero",
		PlanID:      "plan-music-pro",
		BotType:     "music",
		IsZeroSetup: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 799 base + 299 zero setup = 1098 cents ($10.98)
	if resp.FinalPriceCents != 1098 {
		t.Fatalf("expected final price 1098 cents with zero setup, got %d", resp.FinalPriceCents)
	}

	if resp.ApprovalURL == "" {
		t.Fatal("expected approval URL for paid order")
	}
}
