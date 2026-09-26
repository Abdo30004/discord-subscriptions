package pinger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
	"github.com/discord-subscriptions/shared/events"
)

type HTTPPinger struct {
	client *http.Client
}

// NewHTTPPinger creates an active HTTP probe with a 5-second timeout.
func NewHTTPPinger() *HTTPPinger {
	return &HTTPPinger{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Ping performs an HTTP GET check against the target health endpoint.
func (p *HTTPPinger) Ping(ctx context.Context, target *domain.MonitoringTarget) domain.CheckResult {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.HealthURL, nil)
	if err != nil {
		return domain.CheckResult{
			TargetID:     target.ID,
			Status:       events.BotStatusOffline,
			StatusCode:   0,
			LatencyMs:    0,
			ErrorMessage: fmt.Sprintf("failed to construct request: %v", err),
			CheckedAt:    time.Now().UTC(),
		}
	}

	req.Header.Set("User-Agent", "Discord-Subscriptions-Monitor/1.0")

	resp, err := p.client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return domain.CheckResult{
			TargetID:     target.ID,
			Status:       events.BotStatusOffline,
			StatusCode:   0,
			LatencyMs:    latency,
			ErrorMessage: err.Error(),
			CheckedAt:    time.Now().UTC(),
		}
	}
	defer resp.Body.Close()

	result := domain.CheckResult{
		TargetID:   target.ID,
		StatusCode: resp.StatusCode,
		LatencyMs:  latency,
		CheckedAt:  time.Now().UTC(),
	}

	if resp.StatusCode != http.StatusOK {
		result.Status = events.BotStatusOffline
		result.ErrorMessage = fmt.Sprintf("unhealthy HTTP response code: %d", resp.StatusCode)
		return result
	}

	// Parse health payload if available
	var body domain.BotHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err == nil {
		result.DiscordPingMs = body.DiscordPingMs
		result.MemoryUsageMb = body.MemoryUsageMb

		switch body.Status {
		case "ok":
			result.Status = events.BotStatusOnline
		case "degraded":
			result.Status = events.BotStatusDegraded
		default:
			result.Status = events.BotStatusOffline
			result.ErrorMessage = fmt.Sprintf("bot reported status: %s", body.Status)
		}
	} else {
		// Valid HTTP 200 without custom body is still considered online
		result.Status = events.BotStatusOnline
	}

	return result
}
