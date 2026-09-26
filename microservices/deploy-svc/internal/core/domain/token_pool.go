package domain

import (
	"errors"
	"time"
)

type TokenStatus string

const (
	TokenStatusAvailable   TokenStatus = "available"
	TokenStatusAssigned    TokenStatus = "assigned"
	TokenStatusQuarantined TokenStatus = "quarantined"
)

var (
	ErrNoTokenAvailable = errors.New("no pre-warmed tokens available for this bot type")
	ErrTokenNotFound    = errors.New("token pool entry not found")
	ErrRateLimitExceeded = errors.New("discord rate limit reached: bot name/avatar can only be changed twice per hour")
)

// BotTokenPoolEntry represents a pre-warmed bot token available for 0-setup provisioning.
type BotTokenPoolEntry struct {
	ID                     string      `json:"id"`
	BotType                string      `json:"bot_type"`
	ClientID               string      `json:"client_id"`
	TokenVaultPath         string      `json:"token_vault_path,omitempty"`
	TokenEncrypted         string      `json:"token_encrypted,omitempty"`
	TokenMasked            string      `json:"token_masked,omitempty"`
	Status                 TokenStatus `json:"status"`
	AssignedGuildID        string      `json:"assigned_guild_id,omitempty"`
	AssignedSubscriptionID string      `json:"assigned_subscription_id,omitempty"`
	AssignedAt             *time.Time  `json:"assigned_at,omitempty"`
	CreatedAt              time.Time   `json:"created_at"`
	UpdatedAt              time.Time   `json:"updated_at"`
}

// IsAvailable returns true if the token is ready to be leased.
func (t *BotTokenPoolEntry) IsAvailable() bool {
	return t.Status == TokenStatusAvailable
}

// Assign marks the token as assigned to a specific guild and subscription.
func (t *BotTokenPoolEntry) Assign(guildID, subscriptionID string) {
	now := time.Now().UTC()
	t.Status = TokenStatusAssigned
	t.AssignedGuildID = guildID
	t.AssignedSubscriptionID = subscriptionID
	t.AssignedAt = &now
	t.UpdatedAt = now
}

// Quarantine marks the token as quarantined pending recycle/reset.
func (t *BotTokenPoolEntry) Quarantine() {
	t.Status = TokenStatusQuarantined
	t.UpdatedAt = time.Now().UTC()
}

// Release resets the token back to available.
func (t *BotTokenPoolEntry) Release() {
	t.Status = TokenStatusAvailable
	t.AssignedGuildID = ""
	t.AssignedSubscriptionID = ""
	t.AssignedAt = nil
	t.UpdatedAt = time.Now().UTC()
}
