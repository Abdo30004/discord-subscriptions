package domain

import (
	"time"

	"github.com/discord-subscriptions/shared/events"
)

// CheckResult represents the outcome of a single active health probe.
type CheckResult struct {
	TargetID      string           `json:"target_id"`
	Status        events.BotStatus `json:"status"`
	StatusCode    int              `json:"status_code"`
	LatencyMs     int64            `json:"latency_ms"`
	ErrorMessage  string           `json:"error_message,omitempty"`
	DiscordPingMs int64            `json:"discord_ping_ms,omitempty"`
	MemoryUsageMb int64            `json:"memory_usage_mb,omitempty"`
	CheckedAt     time.Time        `json:"checked_at"`
}

// BotHealthResponse is the expected JSON response payload from a bot's internal /health route.
type BotHealthResponse struct {
	Status        string `json:"status"` // "ok", "degraded", "error"
	BotID         string `json:"bot_id"`
	GuildID       string `json:"guild_id"`
	DiscordPingMs int64  `json:"discord_ping_ms"`
	MemoryUsageMb int64  `json:"memory_usage_mb"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}
