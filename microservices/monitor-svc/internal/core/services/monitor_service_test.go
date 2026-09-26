package services_test

import (
	"context"
	"testing"

	"github.com/discord-subscriptions/monitor-svc/internal/core/domain"
	"github.com/discord-subscriptions/monitor-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/events"
)

type mockRepo struct {
	targets map[string]*domain.MonitoringTarget
	logs    []domain.CheckResult
}

func (m *mockRepo) CreateTarget(ctx context.Context, target *domain.MonitoringTarget) error {
	m.targets[target.ID] = target
	return nil
}
func (m *mockRepo) GetTargetByID(ctx context.Context, id string) (*domain.MonitoringTarget, error) {
	return m.targets[id], nil
}
func (m *mockRepo) GetTargetByGuildID(ctx context.Context, guildID string) (*domain.MonitoringTarget, error) {
	for _, t := range m.targets {
		if t.GuildID == guildID {
			return t, nil
		}
	}
	return nil, nil
}
func (m *mockRepo) ListTargetsByGuildID(ctx context.Context, guildID string) ([]domain.MonitoringTarget, error) {
	var list []domain.MonitoringTarget
	for _, t := range m.targets {
		if t.GuildID == guildID {
			list = append(list, *t)
		}
	}
	return list, nil
}
func (m *mockRepo) GetActiveTargets(ctx context.Context) ([]domain.MonitoringTarget, error) {
	var list []domain.MonitoringTarget
	for _, t := range m.targets {
		if t.IsActive {
			list = append(list, *t)
		}
	}
	return list, nil
}
func (m *mockRepo) UpdateTarget(ctx context.Context, target *domain.MonitoringTarget) error {
	m.targets[target.ID] = target
	return nil
}
func (m *mockRepo) SaveCheckLog(ctx context.Context, log *domain.CheckResult) error {
	m.logs = append(m.logs, *log)
	return nil
}
func (m *mockRepo) GetRecentLogs(ctx context.Context, targetID string, limit int) ([]domain.CheckResult, error) {
	return m.logs, nil
}

type mockPinger struct {
	result domain.CheckResult
}

func (m *mockPinger) Ping(ctx context.Context, target *domain.MonitoringTarget) domain.CheckResult {
	res := m.result
	res.TargetID = target.ID
	return res
}

type mockPublisher struct {
	published []events.Event
}

func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, event events.Event) error {
	m.published = append(m.published, event)
	return nil
}

func TestMonitorService_RegisterAndCheck(t *testing.T) {
	repo := &mockRepo{targets: make(map[string]*domain.MonitoringTarget)}
	pinger := &mockPinger{
		result: domain.CheckResult{
			Status:     events.BotStatusOnline,
			StatusCode: 200,
			LatencyMs:  45,
		},
	}
	pub := &mockPublisher{}

	svc := services.NewMonitorService(repo, pinger, pub, nil)

	// 1. Register target
	tgt, err := svc.RegisterTarget(context.Background(), "bot-1", "guild-1", "Main Stage", "http://localhost:8080/health", 15)
	if err != nil {
		t.Fatalf("unexpected error registering target: %v", err)
	}

	if tgt.CurrentStatus != events.BotStatusOnline {
		t.Fatalf("expected initial status online, got %s", tgt.CurrentStatus)
	}

	if tgt.InstanceLabel != "Main Stage" {
		t.Fatalf("expected instance label 'Main Stage', got %s", tgt.InstanceLabel)
	}

	// 2. Simulate 2 consecutive failure pings to trigger offline state transition
	pinger.result.Status = events.BotStatusOffline
	pinger.result.StatusCode = 503

	_ = svc.CheckTarget(context.Background(), tgt) // 1st failure
	_ = svc.CheckTarget(context.Background(), tgt) // 2nd failure -> offline

	if tgt.CurrentStatus != events.BotStatusOffline {
		t.Fatalf("expected target to become offline, got %s", tgt.CurrentStatus)
	}

	if len(pub.published) != 1 {
		t.Fatalf("expected 1 status change event published, got %d", len(pub.published))
	}

	if pub.published[0].GetType() != events.TypeBotStatusChanged {
		t.Fatalf("expected %s, got %s", events.TypeBotStatusChanged, pub.published[0].GetType())
	}
}

func TestMonitorService_MultiTarget_Guild(t *testing.T) {
	repo := &mockRepo{targets: make(map[string]*domain.MonitoringTarget)}
	pinger := &mockPinger{
		result: domain.CheckResult{Status: events.BotStatusOnline, StatusCode: 200},
	}
	svc := services.NewMonitorService(repo, pinger, &mockPublisher{}, nil)

	_, err := svc.RegisterTarget(context.Background(), "bot-music", "guild-multi", "Lounge Music", "http://music.svc:8080/health", 15)
	if err != nil {
		t.Fatalf("register music target failed: %v", err)
	}

	_, err = svc.RegisterTarget(context.Background(), "bot-mod", "guild-multi", "Aegis Guard", "http://mod.svc:8080/health", 15)
	if err != nil {
		t.Fatalf("register mod target failed: %v", err)
	}

	targets, err := svc.GetGuildTargets(context.Background(), "guild-multi")
	if err != nil {
		t.Fatalf("GetGuildTargets failed: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets for guild-multi, got %d", len(targets))
	}
}
