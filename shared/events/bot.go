package events

const (
	TypeBotStatusChanged EventType = "bot.status.changed"
	TypeBotHeartbeat     EventType = "bot.heartbeat"
)

// BotStatus indicates the operational state of a bot instance.
type BotStatus string

const (
	BotStatusOnline   BotStatus = "online"
	BotStatusDegraded BotStatus = "degraded"
	BotStatusOffline  BotStatus = "offline"
)

// BotStatusChangedEvent is emitted by monitor-svc when a bot transitions state.
type BotStatusChangedEvent struct {
	BaseEvent
	BotID        string    `json:"bot_id"`
	GuildID      string    `json:"guild_id"`
	Previous     BotStatus `json:"previous_status"`
	Current      BotStatus `json:"current_status"`
	Reason       string    `json:"reason,omitempty"`
	LatencyMs    int64     `json:"latency_ms"`
}

// NewBotStatusChangedEvent constructs a ready-to-publish BotStatusChangedEvent.
func NewBotStatusChangedEvent(botID, guildID string, prev, curr BotStatus, reason string, latencyMs int64) BotStatusChangedEvent {
	return BotStatusChangedEvent{
		BaseEvent: NewBaseEvent(TypeBotStatusChanged),
		BotID:     botID,
		GuildID:   guildID,
		Previous:  prev,
		Current:   curr,
		Reason:    reason,
		LatencyMs: latencyMs,
	}
}
