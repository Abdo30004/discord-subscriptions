package domain

import (
	"errors"
	"time"
)

var (
	ErrSubscriptionExpired   = errors.New("subscription has expired")
	ErrSubscriptionInactive  = errors.New("subscription is not active")
	ErrInvalidSubscriptionID = errors.New("invalid subscription ID")
	ErrInvalidGuildID        = errors.New("guild ID is required")
)

type SubscriptionStatus string

const (
	StatusActive    SubscriptionStatus = "active"
	StatusCancelled SubscriptionStatus = "cancelled"
	StatusExpired   SubscriptionStatus = "expired"
)

type PaymentProvider string

const (
	ProviderPayPal     PaymentProvider = "paypal"
	ProviderVoucher    PaymentProvider = "gift_code"
	ProviderAdminGrant PaymentProvider = "admin_grant"
)

// Subscription represents an active license for a bot on a Discord guild.
type Subscription struct {
	ID             string             `json:"id"`
	UserID         string             `json:"user_id"`
	GuildID        string             `json:"guild_id"`
	PlanID         string             `json:"plan_id"`
	BotType        string             `json:"bot_type"`
	InstanceLabel  string             `json:"instance_label"`
	Status         SubscriptionStatus `json:"status"`
	Provider       PaymentProvider    `json:"provider"`
	ExternalSubID  string             `json:"external_sub_id,omitempty"` // PayPal Subscription ID if applicable
	IsDedicated    bool               `json:"is_dedicated"`
	IsZeroSetup    bool               `json:"is_zero_setup"`
	ValidUntil     time.Time          `json:"valid_until"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// IsActiveNow checks whether the subscription is active and has not expired.
func (s *Subscription) IsActiveNow() bool {
	if s.Status != StatusActive {
		return false
	}
	return time.Now().UTC().Before(s.ValidUntil)
}

// Cancel updates the status to cancelled.
func (s *Subscription) Cancel() {
	s.Status = StatusCancelled
	s.UpdatedAt = time.Now().UTC()
}

// Extend adds duration to the subscription.
func (s *Subscription) Extend(duration time.Duration) {
	if time.Now().UTC().After(s.ValidUntil) {
		s.ValidUntil = time.Now().UTC().Add(duration)
	} else {
		s.ValidUntil = s.ValidUntil.Add(duration)
	}
	s.Status = StatusActive
	s.UpdatedAt = time.Now().UTC()
}
