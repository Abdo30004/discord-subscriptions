package domain_test

import (
	"testing"
	"time"

	"github.com/discord-subscriptions/billing-svc/internal/core/domain"
)

func TestPromoCode_CalculateDiscount(t *testing.T) {
	// 50% discount on $10.00 (1000 cents)
	p50 := domain.PromoCode{
		DiscountType:  domain.DiscountPercentage,
		DiscountValue: 50,
		IsActive:      true,
	}

	disc, final := p50.CalculateDiscount(1000)
	if disc != 500 || final != 500 {
		t.Fatalf("expected 500 disc and 500 final, got disc=%d, final=%d", disc, final)
	}

	// Fixed $3.00 off (300 cents) on $10.00 (1000 cents)
	pFixed := domain.PromoCode{
		DiscountType:  domain.DiscountFixed,
		DiscountValue: 300,
		IsActive:      true,
	}
	disc, final = pFixed.CalculateDiscount(1000)
	if disc != 300 || final != 700 {
		t.Fatalf("expected 300 disc and 700 final, got disc=%d, final=%d", disc, final)
	}

	// Fixed discount greater than total price cap to 0
	disc, final = pFixed.CalculateDiscount(200)
	if disc != 200 || final != 0 {
		t.Fatalf("expected cap to price, got disc=%d, final=%d", disc, final)
	}
}

func TestPromoCode_IsUsable(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-1 * time.Hour)
	future := now.Add(24 * time.Hour)

	// Inactive promo
	inactive := domain.PromoCode{IsActive: false}
	if err := inactive.IsUsable(); err != domain.ErrPromoInactive {
		t.Fatalf("expected ErrPromoInactive, got %v", err)
	}

	// Expired promo
	expired := domain.PromoCode{IsActive: true, ExpiresAt: &past}
	if err := expired.IsUsable(); err != domain.ErrPromoExpired {
		t.Fatalf("expected ErrPromoExpired, got %v", err)
	}

	// Depleted promo
	depleted := domain.PromoCode{IsActive: true, MaxUses: 5, CurrentUses: 5, ExpiresAt: &future}
	if err := depleted.IsUsable(); err != domain.ErrPromoDepleted {
		t.Fatalf("expected ErrPromoDepleted, got %v", err)
	}

	// Valid promo
	valid := domain.PromoCode{IsActive: true, MaxUses: 10, CurrentUses: 2, ExpiresAt: &future}
	if err := valid.IsUsable(); err != nil {
		t.Fatalf("expected valid promo, got %v", err)
	}
}

func TestVoucherCode_Redeem(t *testing.T) {
	voucher := domain.VoucherCode{
		Code:       "GIFT-TEST-123",
		IsRedeemed: false,
	}

	err := voucher.Redeem("user-1", "guild-1")
	if err != nil {
		t.Fatalf("unexpected error redeeming voucher: %v", err)
	}

	if !voucher.IsRedeemed || voucher.RedeemedByUserID != "user-1" {
		t.Fatal("expected voucher to be marked as redeemed")
	}

	// Attempting to redeem again must fail
	err = voucher.Redeem("user-2", "guild-2")
	if err != domain.ErrVoucherAlreadyRedeemed {
		t.Fatalf("expected ErrVoucherAlreadyRedeemed, got %v", err)
	}
}

func TestSubscription_Lifecycle(t *testing.T) {
	sub := domain.Subscription{
		Status:     domain.StatusActive,
		ValidUntil: time.Now().UTC().Add(30 * 24 * time.Hour),
	}

	if !sub.IsActiveNow() {
		t.Fatal("expected subscription to be active")
	}

	sub.Cancel()
	if sub.IsActiveNow() {
		t.Fatal("expected subscription to be inactive after cancellation")
	}

	// Extend subscription by 7 days
	sub.Extend(7 * 24 * time.Hour)
	if !sub.IsActiveNow() {
		t.Fatal("expected subscription to be reactivated after extension")
	}
}
