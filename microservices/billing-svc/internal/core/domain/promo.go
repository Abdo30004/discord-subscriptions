package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrPromoExpired  = errors.New("promo code has expired")
	ErrPromoDepleted = errors.New("promo code has reached maximum uses")
	ErrPromoInactive = errors.New("promo code is inactive")
	ErrInvalidPromo  = errors.New("invalid promo code parameters")
)

type DiscountType string

const (
	DiscountPercentage DiscountType = "percentage" // e.g. 25 for 25% off
	DiscountFixed      DiscountType = "fixed"      // e.g. 200 for $2.00 off in cents
)

// PromoCode represents a discount coupon applicable during checkout.
type PromoCode struct {
	ID            string       `json:"id"`
	Code          string       `json:"code"`
	DiscountType  DiscountType `json:"discount_type"`
	DiscountValue int64        `json:"discount_value"` // Percentage (0-100) or fixed amount in cents
	MaxUses       int          `json:"max_uses"`       // 0 for unlimited
	CurrentUses   int          `json:"current_uses"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
	IsActive      bool         `json:"is_active"`
	CreatedAt     time.Time    `json:"created_at"`
}

// IsUsable validates whether the promo code is active, not expired, and under max uses.
func (p *PromoCode) IsUsable() error {
	if !p.IsActive {
		return ErrPromoInactive
	}
	if p.ExpiresAt != nil && time.Now().UTC().After(*p.ExpiresAt) {
		return ErrPromoExpired
	}
	if p.MaxUses > 0 && p.CurrentUses >= p.MaxUses {
		return ErrPromoDepleted
	}
	return nil
}

// CalculateDiscount computes the discounted amount in cents for an original price.
func (p *PromoCode) CalculateDiscount(originalPriceCents int64) (discountCents int64, finalPriceCents int64) {
	if originalPriceCents <= 0 {
		return 0, 0
	}

	if p.DiscountType == DiscountPercentage {
		percent := p.DiscountValue
		if percent > 100 {
			percent = 100
		} else if percent < 0 {
			percent = 0
		}
		discountCents = (originalPriceCents * percent) / 100
	} else {
		discountCents = p.DiscountValue
		if discountCents > originalPriceCents {
			discountCents = originalPriceCents
		}
	}

	finalPriceCents = originalPriceCents - discountCents
	if finalPriceCents < 0 {
		finalPriceCents = 0
	}

	return discountCents, finalPriceCents
}

// NormalizeCode standardizes the promo code string.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
