package events

import (
	"time"
)

const (
	TypeSubscriptionActivated EventType = "subscription.activated"
	TypeSubscriptionCancelled EventType = "subscription.cancelled"
	TypeSubscriptionRenewed   EventType = "subscription.renewed"
)

// SubscriptionActivatedEvent is emitted when a payment is confirmed via PayPal or Manager Bot.
type SubscriptionActivatedEvent struct {
	BaseEvent
	SubscriptionID string    `json:"subscription_id"`
	UserID         string    `json:"user_id"`         // Discord User Snowflake
	GuildID        string    `json:"guild_id"`        // Discord Guild Snowflake
	BotType        string    `json:"bot_type"`        // e.g. "music", "game", "utility"
	PlanID         string    `json:"plan_id"`         // e.g. "basic", "pro"
	InstanceLabel  string    `json:"instance_label"`  // e.g. "Main Stage", "VIP Room"
	IsDedicated    bool      `json:"is_dedicated"`    // true for single-tenant, false for multi-tenant
	IsZeroSetup    bool      `json:"is_zero_setup"`   // true if turnkey managed bot token pool used
	ValidUntil     time.Time `json:"valid_until"`
	Provider       string    `json:"provider"`        // "paypal"
}

// NewSubscriptionActivatedEvent creates a ready-to-publish SubscriptionActivatedEvent.
func NewSubscriptionActivatedEvent(subID, userID, guildID, botType, planID, instanceLabel string, isDedicated, isZeroSetup bool, validUntil time.Time, provider string) SubscriptionActivatedEvent {
	return SubscriptionActivatedEvent{
		BaseEvent:      NewBaseEvent(TypeSubscriptionActivated),
		SubscriptionID: subID,
		UserID:         userID,
		GuildID:        guildID,
		BotType:        botType,
		PlanID:         planID,
		InstanceLabel:  instanceLabel,
		IsDedicated:    isDedicated,
		IsZeroSetup:    isZeroSetup,
		ValidUntil:     validUntil,
		Provider:       provider,
	}
}

// SubscriptionCancelledEvent is emitted when a user or billing system cancels a subscription.
type SubscriptionCancelledEvent struct {
	BaseEvent
	SubscriptionID string    `json:"subscription_id"`
	UserID         string    `json:"user_id"`
	GuildID        string    `json:"guild_id"`
	Reason         string    `json:"reason"`
	EffectiveAt    time.Time `json:"effective_at"`
}

// NewSubscriptionCancelledEvent creates a ready-to-publish SubscriptionCancelledEvent.
func NewSubscriptionCancelledEvent(subID, userID, guildID, reason string, effectiveAt time.Time) SubscriptionCancelledEvent {
	return SubscriptionCancelledEvent{
		BaseEvent:      NewBaseEvent(TypeSubscriptionCancelled),
		SubscriptionID: subID,
		UserID:         userID,
		GuildID:        guildID,
		Reason:         reason,
		EffectiveAt:    effectiveAt,
	}
}
