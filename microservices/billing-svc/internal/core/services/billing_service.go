package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
	"github.com/discord-subscriptions/billing-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/events"
	"github.com/discord-subscriptions/shared/messaging"
	"github.com/google/uuid"
)

type BillingServiceImpl struct {
	subRepo   ports.SubscriptionRepository
	promoRepo ports.PromoRepository
	vouchRepo ports.VoucherRepository
	paypal    ports.PayPalGateway
	publisher messaging.Publisher
	logger    *slog.Logger
}

func NewBillingService(
	subRepo ports.SubscriptionRepository,
	promoRepo ports.PromoRepository,
	vouchRepo ports.VoucherRepository,
	paypal ports.PayPalGateway,
	publisher messaging.Publisher,
	logger *slog.Logger,
) *BillingServiceImpl {
	if logger == nil {
		logger = slog.Default()
	}
	return &BillingServiceImpl{
		subRepo:   subRepo,
		promoRepo: promoRepo,
		vouchRepo: vouchRepo,
		paypal:    paypal,
		publisher: publisher,
		logger:    logger,
	}
}

// InitiateCheckout starts a checkout session, applies promo discounts, or activates immediately if 100% off.
func (s *BillingServiceImpl) InitiateCheckout(ctx context.Context, req ports.CheckoutRequest) (*ports.CheckoutResponse, error) {
	// Base price (in a full setup this is fetched from catalog-svc; here default $7.99 / 799 cents)
	basePriceCents := int64(799)
	if req.IsZeroSetup {
		basePriceCents += 299 // $2.99/mo Turnkey Managed Token Pool Add-on
	}
	discountCents := int64(0)
	finalPriceCents := basePriceCents

	// Determine instance label with fallback auto-increment
	instanceLabel := strings.TrimSpace(req.InstanceLabel)
	if instanceLabel == "" {
		existing, _ := s.subRepo.ListByGuildID(ctx, req.GuildID)
		if len(existing) == 0 {
			instanceLabel = "Default"
		} else {
			instanceLabel = fmt.Sprintf("%s #%d", strings.ToUpper(req.BotType[:1])+req.BotType[1:], len(existing)+1)
		}
	}

	// 1. Process Promo Code if provided
	if req.PromoCode != "" {
		promo, err := s.promoRepo.GetByCode(ctx, req.PromoCode)
		if err != nil {
			return nil, fmt.Errorf("promo code not found: %w", err)
		}

		if err := promo.IsUsable(); err != nil {
			return nil, err
		}

		discountCents, finalPriceCents = promo.CalculateDiscount(basePriceCents)
		_ = s.promoRepo.IncrementUsage(ctx, promo.Code)
		s.logger.Info("applied promo code", slog.String("code", promo.Code), slog.Int64("discount", discountCents))
	}

	// 2. If final price is 0 (100% discount promo or free plan), activate instantly!
	if finalPriceCents == 0 {
		sub, err := s.createAndPublishSubscription(
			ctx,
			req.UserID,
			req.GuildID,
			req.BotType,
			req.PlanID,
			instanceLabel,
			domain.ProviderPayPal, // 100% discounted checkout
			"",
			true,
			req.IsZeroSetup,
			30*24*time.Hour,
		)
		if err != nil {
			return nil, err
		}

		return &ports.CheckoutResponse{
			IsFreeInstantActive: true,
			Subscription:        sub,
			OriginalPriceCents:  basePriceCents,
			DiscountCents:       discountCents,
			FinalPriceCents:     finalPriceCents,
		}, nil
	}

	// 3. Initiate PayPal Subscription Order
	approvalURL, _, err := s.paypal.CreateSubscriptionOrder(ctx, req.PlanID, req.ReturnURL, req.CancelURL)
	if err != nil {
		return nil, fmt.Errorf("failed creating paypal subscription: %w", err)
	}

	return &ports.CheckoutResponse{
		IsFreeInstantActive: false,
		ApprovalURL:         approvalURL,
		OriginalPriceCents:  basePriceCents,
		DiscountCents:       discountCents,
		FinalPriceCents:     finalPriceCents,
	}, nil
}

// RedeemVoucher redeems a one-time gift card / voucher code for a guild.
func (s *BillingServiceImpl) RedeemVoucher(ctx context.Context, code, userID, guildID string) (*domain.Subscription, error) {
	voucher, err := s.vouchRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("voucher not found: %w", err)
	}

	if err := voucher.Redeem(userID, guildID); err != nil {
		return nil, err
	}

	if err := s.vouchRepo.Update(ctx, voucher); err != nil {
		return nil, fmt.Errorf("failed saving redeemed voucher: %w", err)
	}

	existing, _ := s.subRepo.ListByGuildID(ctx, guildID)
	instanceLabel := "Default"
	if len(existing) > 0 {
		instanceLabel = fmt.Sprintf("%s #%d", strings.ToUpper(voucher.BotType[:1])+voucher.BotType[1:], len(existing)+1)
	}

	duration := time.Duration(voucher.DurationDays) * 24 * time.Hour
	sub, err := s.createAndPublishSubscription(
		ctx,
		userID,
		guildID,
		voucher.BotType,
		voucher.PlanID,
		instanceLabel,
		domain.ProviderVoucher,
		voucher.Code,
		voucher.IsDedicated,
		false,
		duration,
	)
	if err != nil {
		return nil, err
	}

	s.logger.Info("voucher redeemed successfully",
		slog.String("code", code),
		slog.String("guild_id", guildID),
		slog.String("user_id", userID),
	)

	return sub, nil
}

// AdminGrantSubscription allows administrators to grant a free or custom duration subscription directly to a server.
func (s *BillingServiceImpl) AdminGrantSubscription(
	ctx context.Context,
	userID, guildID, botType, planID, instanceLabel string,
	durationDays int,
	isDedicated, isZeroSetup bool,
) (*domain.Subscription, error) {
	if durationDays <= 0 {
		durationDays = 30
	}
	if instanceLabel == "" {
		instanceLabel = "Default"
	}

	duration := time.Duration(durationDays) * 24 * time.Hour
	sub, err := s.createAndPublishSubscription(
		ctx,
		userID,
		guildID,
		botType,
		planID,
		instanceLabel,
		domain.ProviderAdminGrant,
		"",
		isDedicated,
		isZeroSetup,
		duration,
	)
	if err != nil {
		return nil, err
	}

	s.logger.Info("admin granted subscription",
		slog.String("guild_id", guildID),
		slog.String("plan_id", planID),
		slog.String("instance_label", instanceLabel),
		slog.Int("days", durationDays),
	)

	return sub, nil
}

// HandlePayPalWebhook processes asynchronous subscription events sent by PayPal.
func (s *BillingServiceImpl) HandlePayPalWebhook(ctx context.Context, eventType string, payload []byte) error {
	s.logger.Info("handling paypal webhook", slog.String("event_type", eventType))

	var webhookData struct {
		Resource struct {
			ID               string `json:"id"`
			Status           string `json:"status"`
			CustomID         string `json:"custom_id"` // Holds JSON or guild_id
			PlanID           string `json:"plan_id"`
			SubscriberUserID string `json:"subscriber_user_id"`
		} `json:"resource"`
	}

	if err := json.Unmarshal(payload, &webhookData); err != nil {
		return fmt.Errorf("failed unmarshaling paypal webhook payload: %w", err)
	}

	switch eventType {
	case "BILLING.SUBSCRIPTION.ACTIVATED":
		sub, err := s.subRepo.GetByExternalSubID(ctx, webhookData.Resource.ID)
		if err == nil && sub != nil {
			sub.Status = domain.StatusActive
			sub.ValidUntil = time.Now().UTC().Add(30 * 24 * time.Hour)
			_ = s.subRepo.Update(ctx, sub)

			if s.publisher != nil {
				evt := events.NewSubscriptionActivatedEvent(
					sub.ID, sub.UserID, sub.GuildID, sub.BotType, sub.PlanID,
					sub.InstanceLabel, sub.IsDedicated, sub.IsZeroSetup, sub.ValidUntil, string(domain.ProviderPayPal),
				)
				_ = s.publisher.Publish(ctx, "discord.events", "subscription.activated", evt)
			}
		}

	case "BILLING.SUBSCRIPTION.CANCELLED":
		sub, err := s.subRepo.GetByExternalSubID(ctx, webhookData.Resource.ID)
		if err == nil && sub != nil {
			sub.Cancel()
			_ = s.subRepo.Update(ctx, sub)

			if s.publisher != nil {
				evt := events.NewSubscriptionCancelledEvent(
					sub.ID, sub.UserID, sub.GuildID, "PayPal subscription cancelled", time.Now().UTC(),
				)
				_ = s.publisher.Publish(ctx, "discord.events", "subscription.cancelled", evt)
			}
		}
	}

	return nil
}

func (s *BillingServiceImpl) CreatePromoCode(ctx context.Context, promo *domain.PromoCode) error {
	return s.promoRepo.Create(ctx, promo)
}

func (s *BillingServiceImpl) CreateVoucher(ctx context.Context, voucher *domain.VoucherCode) error {
	return s.vouchRepo.Create(ctx, voucher)
}

func (s *BillingServiceImpl) GetGuildSubscription(ctx context.Context, guildID string) (*domain.Subscription, error) {
	return s.subRepo.GetActiveByGuildID(ctx, guildID)
}

func (s *BillingServiceImpl) GetGuildSubscriptions(ctx context.Context, guildID string) ([]domain.Subscription, error) {
	return s.subRepo.ListByGuildID(ctx, guildID)
}

func (s *BillingServiceImpl) createAndPublishSubscription(
	ctx context.Context,
	userID, guildID, botType, planID, instanceLabel string,
	provider domain.PaymentProvider,
	extID string,
	isDedicated, isZeroSetup bool,
	duration time.Duration,
) (*domain.Subscription, error) {
	subID := fmt.Sprintf("sub-%s", uuid.New().String()[:8])
	now := time.Now().UTC()

	sub := &domain.Subscription{
		ID:            subID,
		UserID:        userID,
		GuildID:       guildID,
		PlanID:        planID,
		BotType:       botType,
		InstanceLabel: instanceLabel,
		Status:        domain.StatusActive,
		Provider:      provider,
		ExternalSubID: extID,
		IsDedicated:   isDedicated,
		IsZeroSetup:   isZeroSetup,
		ValidUntil:    now.Add(duration),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.subRepo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed saving subscription: %w", err)
	}

	// Emit SubscriptionActivatedEvent
	if s.publisher != nil {
		evt := events.NewSubscriptionActivatedEvent(
			sub.ID, sub.UserID, sub.GuildID, sub.BotType, sub.PlanID,
			sub.InstanceLabel, sub.IsDedicated, sub.IsZeroSetup, sub.ValidUntil, string(provider),
		)
		err := s.publisher.Publish(ctx, "discord.events", "subscription.activated", evt)
		if err != nil {
			s.logger.Error("failed publishing SubscriptionActivatedEvent", slog.String("error", err.Error()))
		}
	}

	return sub, nil
}
