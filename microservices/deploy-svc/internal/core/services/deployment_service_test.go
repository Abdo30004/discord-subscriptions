package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
	"github.com/discord-subscriptions/deploy-svc/internal/core/ports"
	"github.com/discord-subscriptions/deploy-svc/internal/core/services"
	"github.com/discord-subscriptions/shared/events"
)

type mockRepo struct {
	deps map[string]*domain.Deployment
}

func (m *mockRepo) Create(ctx context.Context, dep *domain.Deployment) error {
	m.deps[dep.ID] = dep
	return nil
}
func (m *mockRepo) GetByID(ctx context.Context, id string) (*domain.Deployment, error) {
	return m.deps[id], nil
}
func (m *mockRepo) GetByGuildID(ctx context.Context, guildID string) (*domain.Deployment, error) {
	for _, d := range m.deps {
		if d.GuildID == guildID {
			return d, nil
		}
	}
	return nil, nil
}
func (m *mockRepo) GetBySubscriptionID(ctx context.Context, subID string) (*domain.Deployment, error) {
	for _, d := range m.deps {
		if d.SubscriptionID == subID {
			return d, nil
		}
	}
	return nil, nil
}
func (m *mockRepo) Update(ctx context.Context, dep *domain.Deployment) error {
	m.deps[dep.ID] = dep
	return nil
}
func (m *mockRepo) ListByGuildID(ctx context.Context, guildID string) ([]domain.Deployment, error) {
	var list []domain.Deployment
	for _, d := range m.deps {
		if d.GuildID == guildID {
			list = append(list, *d)
		}
	}
	return list, nil
}
func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]domain.Deployment, error) {
	var list []domain.Deployment
	for _, d := range m.deps {
		list = append(list, *d)
	}
	return list, nil
}

type mockTokenPoolRepo struct {
	tokens map[string]*domain.BotTokenPoolEntry
}

func (m *mockTokenPoolRepo) Add(ctx context.Context, entry *domain.BotTokenPoolEntry) error {
	m.tokens[entry.ID] = entry
	return nil
}

func (m *mockTokenPoolRepo) ClaimToken(ctx context.Context, botType, guildID, subscriptionID string) (*domain.BotTokenPoolEntry, error) {
	for _, t := range m.tokens {
		if t.BotType == botType && t.IsAvailable() {
			t.Assign(guildID, subscriptionID)
			return t, nil
		}
	}
	return nil, domain.ErrNoTokenAvailable
}

func (m *mockTokenPoolRepo) ReleaseToken(ctx context.Context, guildID string) error {
	for _, t := range m.tokens {
		if t.AssignedGuildID == guildID {
			t.Quarantine()
			return nil
		}
	}
	return nil
}

func (m *mockTokenPoolRepo) GetAvailableCount(ctx context.Context, botType string) (int, error) {
	count := 0
	for _, t := range m.tokens {
		if t.BotType == botType && t.IsAvailable() {
			count++
		}
	}
	return count, nil
}

func (m *mockTokenPoolRepo) GetPoolStats(ctx context.Context) (map[string]map[string]int, error) {
	stats := make(map[string]map[string]int)
	for _, t := range m.tokens {
		if _, ok := stats[t.BotType]; !ok {
			stats[t.BotType] = make(map[string]int)
		}
		stats[t.BotType][string(t.Status)]++
	}
	return stats, nil
}

func (m *mockTokenPoolRepo) List(ctx context.Context, botType string, status domain.TokenStatus, limit, offset int) ([]domain.BotTokenPoolEntry, error) {
	var res []domain.BotTokenPoolEntry
	for _, t := range m.tokens {
		if (botType == "" || t.BotType == botType) && (status == "" || t.Status == status) {
			res = append(res, *t)
		}
	}
	return res, nil
}

type mockK8s struct {
	deployed []ports.BotDeployParams
	stopped  []string
}

func (m *mockK8s) DeployBot(ctx context.Context, params ports.BotDeployParams) error {
	m.deployed = append(m.deployed, params)
	return nil
}
func (m *mockK8s) StopBot(ctx context.Context, namespace, deploymentName string) error {
	m.stopped = append(m.stopped, deploymentName)
	return nil
}
func (m *mockK8s) RestartBot(ctx context.Context, namespace, deploymentName string) error {
	return nil
}
func (m *mockK8s) GetBotStatus(ctx context.Context, namespace, deploymentName string) (domain.DeploymentStatus, error) {
	return domain.StatusRunning, nil
}

type mockVault struct {
	tokens map[string]string
}

func (m *mockVault) PutBotToken(ctx context.Context, botID, token string) error {
	m.tokens[botID] = token
	return nil
}
func (m *mockVault) GetBotToken(ctx context.Context, botID string) (string, error) {
	return m.tokens[botID], nil
}

type mockPublisher struct {
	published []events.Event
}

func (m *mockPublisher) Publish(ctx context.Context, exchange, routingKey string, event events.Event) error {
	m.published = append(m.published, event)
	return nil
}

func TestDeploymentService_ProvisionBot_SelfSetup(t *testing.T) {
	repo := &mockRepo{deps: make(map[string]*domain.Deployment)}
	poolRepo := &mockTokenPoolRepo{tokens: make(map[string]*domain.BotTokenPoolEntry)}
	k8sMock := &mockK8s{}
	vaultMock := &mockVault{tokens: make(map[string]string)}
	pubMock := &mockPublisher{}

	svc := services.NewDeploymentService(repo, poolRepo, k8sMock, vaultMock, pubMock, nil)

	dep, err := svc.ProvisionBot(
		context.Background(),
		"sub-abc",
		"user-123",
		"guild-999",
		"music",
		"Main Stage",
		"secret_discord_token_xyz",
		"v1.0.0",
		false,
	)

	if err != nil {
		t.Fatalf("unexpected error provisioning bot: %v", err)
	}

	if dep.Status != domain.StatusRunning {
		t.Fatalf("expected status running, got %s", dep.Status)
	}

	if dep.InstanceLabel != "Main Stage" {
		t.Fatalf("expected instance label 'Main Stage', got %s", dep.InstanceLabel)
	}

	if len(k8sMock.deployed) != 1 {
		t.Fatalf("expected 1 bot deployed to k8s, got %d", len(k8sMock.deployed))
	}

	if vaultMock.tokens[dep.ID] != "secret_discord_token_xyz" {
		t.Fatal("expected token to be securely saved in Vault")
	}

	if len(pubMock.published) != 1 {
		t.Fatalf("expected 1 completed event published, got %d", len(pubMock.published))
	}

	if pubMock.published[0].GetType() != events.TypeDeploymentCompleted {
		t.Fatalf("expected %s, got %s", events.TypeDeploymentCompleted, pubMock.published[0].GetType())
	}
}

func TestDeploymentService_ProvisionBot_ZeroSetup(t *testing.T) {
	repo := &mockRepo{deps: make(map[string]*domain.Deployment)}
	poolRepo := &mockTokenPoolRepo{
		tokens: map[string]*domain.BotTokenPoolEntry{
			"pool-music-1": {
				ID:             "pool-music-1",
				BotType:        "music",
				ClientID:       "131234567890123456",
				TokenEncrypted: "managed_token_music_abc",
				Status:         domain.TokenStatusAvailable,
				CreatedAt:      time.Now().UTC(),
				UpdatedAt:      time.Now().UTC(),
			},
		},
	}
	k8sMock := &mockK8s{}
	vaultMock := &mockVault{tokens: make(map[string]string)}
	pubMock := &mockPublisher{}

	svc := services.NewDeploymentService(repo, poolRepo, k8sMock, vaultMock, pubMock, nil)

	dep, err := svc.ProvisionBot(
		context.Background(),
		"sub-zero-1",
		"user-555",
		"guild-888",
		"music",
		"VIP Bot",
		"",
		"v1.0.0",
		true,
	)

	if err != nil {
		t.Fatalf("unexpected error in zero-setup provisioning: %v", err)
	}

	if !dep.IsZeroSetup {
		t.Fatal("expected IsZeroSetup to be true")
	}

	if dep.ClientID != "131234567890123456" {
		t.Fatalf("expected client ID 131234567890123456, got %s", dep.ClientID)
	}

	// Verify token pool entry is marked assigned
	if poolRepo.tokens["pool-music-1"].Status != domain.TokenStatusAssigned {
		t.Fatalf("expected pool token to be assigned, got %s", poolRepo.tokens["pool-music-1"].Status)
	}

	// Test StopBot releases/quarantines pool token
	err = svc.StopBot(context.Background(), dep.ID)
	if err != nil {
		t.Fatalf("unexpected error stopping bot: %v", err)
	}

	if poolRepo.tokens["pool-music-1"].Status != domain.TokenStatusQuarantined {
		t.Fatalf("expected pool token to be quarantined, got %s", poolRepo.tokens["pool-music-1"].Status)
	}
}

func TestDeploymentService_CustomizeBot(t *testing.T) {
	repo := &mockRepo{
		deps: map[string]*domain.Deployment{
			"dep-1": {
				ID:      "dep-1",
				GuildID: "guild-777",
				Status:  domain.StatusRunning,
			},
		},
	}
	poolRepo := &mockTokenPoolRepo{tokens: make(map[string]*domain.BotTokenPoolEntry)}
	vaultMock := &mockVault{tokens: map[string]string{"dep-1": "simulated_discord_token"}}

	svc := services.NewDeploymentService(repo, poolRepo, &mockK8s{}, vaultMock, &mockPublisher{}, nil)

	// In simulated/dev environment, CustomizeBot should succeed cleanly with guild ID or dep ID
	err := svc.CustomizeBot(context.Background(), "guild-777", "NewBotName", "")
	if err != nil {
		t.Fatalf("unexpected error customizing bot by guild: %v", err)
	}

	err = svc.CustomizeBot(context.Background(), "dep-1", "NewBotName2", "")
	if err != nil {
		t.Fatalf("unexpected error customizing bot by dep ID: %v", err)
	}
}

func TestDeploymentService_MultiBot_Isolation(t *testing.T) {
	repo := &mockRepo{deps: make(map[string]*domain.Deployment)}
	poolRepo := &mockTokenPoolRepo{tokens: make(map[string]*domain.BotTokenPoolEntry)}
	k8sMock := &mockK8s{}
	vaultMock := &mockVault{tokens: make(map[string]string)}
	pubMock := &mockPublisher{}

	svc := services.NewDeploymentService(repo, poolRepo, k8sMock, vaultMock, pubMock, nil)

	// Deploy Bot 1 in guild-fleet
	dep1, err := svc.ProvisionBot(
		context.Background(),
		"sub-1",
		"user-owner",
		"guild-fleet",
		"music",
		"Lounge Beats",
		"token-music-1",
		"v1.0",
		false,
	)
	if err != nil {
		t.Fatalf("failed provisioning bot 1: %v", err)
	}

	// Deploy Bot 2 in SAME guild-fleet
	dep2, err := svc.ProvisionBot(
		context.Background(),
		"sub-2",
		"user-owner",
		"guild-fleet",
		"moderation",
		"Security Watch",
		"token-mod-2",
		"v1.0",
		false,
	)
	if err != nil {
		t.Fatalf("failed provisioning bot 2: %v", err)
	}

	// Check K8s deployment names do NOT collide
	if dep1.K8sDeploymentName == dep2.K8sDeploymentName {
		t.Fatalf("k8s deployment name collision: both are %s", dep1.K8sDeploymentName)
	}

	// Check ListByGuildID returns both instances
	list, err := svc.GetDeploymentsByGuild(context.Background(), "guild-fleet")
	if err != nil {
		t.Fatalf("GetDeploymentsByGuild failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 deployments for guild-fleet, got %d", len(list))
	}
}
