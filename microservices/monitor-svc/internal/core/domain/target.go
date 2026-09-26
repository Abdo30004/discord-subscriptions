package domain

import (
	"errors"
	"net/url"
	"time"

	"github.com/discord-subscriptions/shared/events"
)

var (
	ErrInvalidHealthURL = errors.New("invalid or empty health check URL")
	ErrInvalidGuildID   = errors.New("guild ID cannot be empty")
)

// MonitoringTarget represents an endpoint being actively polled for uptime and health.
type MonitoringTarget struct {
	ID                  string           `json:"id"`
	BotID               string           `json:"bot_id"`
	GuildID             string           `json:"guild_id"`
	InstanceLabel       string           `json:"instance_label"`
	HealthURL           string           `json:"health_url"`
	PollIntervalSec     int              `json:"poll_interval_sec"`
	IsActive            bool             `json:"is_active"`
	CurrentStatus       events.BotStatus `json:"current_status"`
	ConsecutiveFailures int              `json:"consecutive_failures"`
	LastCheckedAt       *time.Time       `json:"last_checked_at,omitempty"`
	CreatedAt           time.Time        `json:"created_at"`
	UpdatedAt           time.Time        `json:"updated_at"`
}

// Validate checks target configuration.
func (t *MonitoringTarget) Validate() error {
	if t.GuildID == "" {
		return ErrInvalidGuildID
	}
	if t.HealthURL == "" {
		return ErrInvalidHealthURL
	}
	u, err := url.ParseRequestURI(t.HealthURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ErrInvalidHealthURL
	}
	return nil
}

// ApplyCheckResult updates the target's current state and returns whether the status changed.
func (t *MonitoringTarget) ApplyCheckResult(res CheckResult) (statusChanged bool, oldStatus events.BotStatus) {
	oldStatus = t.CurrentStatus
	now := time.Now().UTC()
	t.LastCheckedAt = &now
	t.UpdatedAt = now

	if res.Status == events.BotStatusOnline {
		t.ConsecutiveFailures = 0
		t.CurrentStatus = events.BotStatusOnline
	} else if res.Status == events.BotStatusDegraded {
		t.CurrentStatus = events.BotStatusDegraded
	} else {
		t.ConsecutiveFailures++
		// After 2 consecutive failures, mark offline
		if t.ConsecutiveFailures >= 2 {
			t.CurrentStatus = events.BotStatusOffline
		}
	}

	statusChanged = (oldStatus != t.CurrentStatus)
	return statusChanged, oldStatus
}
