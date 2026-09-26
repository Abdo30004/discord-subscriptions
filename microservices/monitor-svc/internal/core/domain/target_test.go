package domain_test

import (
	"testing"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
	"github.com/discord-subscriptions/shared/events"
)

func TestMonitoringTarget_Validate(t *testing.T) {
	tests := []struct {
		name    string
		target  domain.MonitoringTarget
		wantErr bool
	}{
		{
			name: "valid target",
			target: domain.MonitoringTarget{
				GuildID:   "guild-123",
				HealthURL: "http://bot-guild-123.discord-bots.svc.cluster.local:8080/health",
			},
			wantErr: false,
		},
		{
			name: "missing guild id",
			target: domain.MonitoringTarget{
				HealthURL: "http://localhost:8080/health",
			},
			wantErr: true,
		},
		{
			name: "invalid url",
			target: domain.MonitoringTarget{
				GuildID:   "guild-123",
				HealthURL: "not-a-valid-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.target.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestMonitoringTarget_ApplyCheckResult(t *testing.T) {
	target := domain.MonitoringTarget{
		CurrentStatus: events.BotStatusOnline,
	}

	// 1 failure should not immediately mark offline (hysteresis to prevent flap)
	changed, old := target.ApplyCheckResult(domain.CheckResult{
		Status: events.BotStatusOffline,
	})
	if changed {
		t.Fatal("expected no immediate status change after only 1 failure")
	}
	if old != events.BotStatusOnline {
		t.Fatalf("expected old status online, got %s", old)
	}

	// 2nd consecutive failure triggers offline status
	changed, _ = target.ApplyCheckResult(domain.CheckResult{
		Status: events.BotStatusOffline,
	})
	if !changed {
		t.Fatal("expected status change after 2 consecutive failures")
	}
	if target.CurrentStatus != events.BotStatusOffline {
		t.Fatalf("expected status offline, got %s", target.CurrentStatus)
	}

	// Recovery to online
	changed, _ = target.ApplyCheckResult(domain.CheckResult{
		Status: events.BotStatusOnline,
	})
	if !changed || target.CurrentStatus != events.BotStatusOnline {
		t.Fatal("expected recovery to online")
	}
	if target.ConsecutiveFailures != 0 {
		t.Fatalf("expected failures counter reset to 0, got %d", target.ConsecutiveFailures)
	}
}
