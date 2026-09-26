package ports

import (
	"context"
	"net/http"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
)

// CheckoutRequest holds parameters for initiating a subscription checkout.
type CheckoutRequest struct {
	UserID        string `json:"user_id"`
	GuildID       string `json:"guild_id"`
	PlanID        string `json:"plan_id"`
	BotType       string `json:"bot_type"`
	InstanceLabel string `json:"instance_label,omitempty"`
	PromoCode     string `json:"promo_code,omitempty"`
	IsZeroSetup   bool   `json:"is_zero_setup"`
	ReturnURL     string `json:"return_url"`
	CancelURL     string `json:"cancel_url"`
}

// CheckoutResponse contains the result of a checkout attempt.
type CheckoutResponse struct {
	IsFreeInstantActive bool                 `json:"is_free_instant_active"`
	Subscription        *domain.Subscription `json:"subscription,omitempty"`
	ApprovalURL         string               `json:"approval_url,omitempty"`
	OriginalPriceCents  int64                `json:"original_price_cents"`
	DiscountCents       int64                `json:"discount_cents"`
	FinalPriceCents     int64                `json:"final_price_cents"`
}

// SubscriptionRepository persists and retrieves subscription records.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id string) (*domain.Subscription, error)
	GetActiveByGuildID(ctx context.Context, guildID string) (*domain.Subscription, error)
	ListByGuildID(ctx context.Context, guildID string) ([]domain.Subscription, error)
	GetByExternalSubID(ctx context.Context, extID string) (*domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
}

// PromoRepository manages promo code records.
type PromoRepository interface {
	Create(ctx context.Context, promo *domain.PromoCode) error
	GetByCode(ctx context.Context, code string) (*domain.PromoCode, error)
	IncrementUsage(ctx context.Context, code string) error
}

// VoucherRepository manages gift / voucher code records.
type VoucherRepository interface {
	Create(ctx context.Context, voucher *domain.VoucherCode) error
	GetByCode(ctx context.Context, code string) (*domain.VoucherCode, error)
	Update(ctx context.Context, voucher *domain.VoucherCode) error
}

// PayPalGateway abstracts PayPal API operations.
type PayPalGateway interface {
	CreateSubscriptionOrder(ctx context.Context, planID, returnURL, cancelURL string) (approvalURL, subID string, err error)
	VerifyWebhookSignature(r *http.Request, webhookID string) bool
}

// BillingService defines the business operations of the billing domain.
type BillingService interface {
	InitiateCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResponse, error)
	RedeemVoucher(ctx context.Context, code, userID, guildID string) (*domain.Subscription, error)
	AdminGrantSubscription(ctx context.Context, userID, guildID, botType, planID, instanceLabel string, durationDays int, isDedicated, isZeroSetup bool) (*domain.Subscription, error)
	CreatePromoCode(ctx context.Context, promo *domain.PromoCode) error
	CreateVoucher(ctx context.Context, voucher *domain.VoucherCode) error
	GetGuildSubscription(ctx context.Context, guildID string) (*domain.Subscription, error)
	GetGuildSubscriptions(ctx context.Context, guildID string) ([]domain.Subscription, error)
	HandlePayPalWebhook(ctx context.Context, eventType string, payload []byte) error
}
