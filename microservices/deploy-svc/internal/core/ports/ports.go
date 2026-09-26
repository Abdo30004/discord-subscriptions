package ports

import (
	"context"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
)

// BotDeployParams holds parameters needed by Kubernetes to spin up a bot pod.
type BotDeployParams struct {
	Namespace      string
	DeploymentName string
	Image          string
	BotToken       string
	BotID          string
	GuildID        string
}

// AddTokenInput holds payload for adding a pre-warmed token into the pool.
type AddTokenInput struct {
	ClientID string `json:"client_id"`
	Token    string `json:"token"`
}

// DeploymentRepository handles persistence of deployment records.
type DeploymentRepository interface {
	Create(ctx context.Context, dep *domain.Deployment) error
	GetByID(ctx context.Context, id string) (*domain.Deployment, error)
	GetByGuildID(ctx context.Context, guildID string) (*domain.Deployment, error)
	ListByGuildID(ctx context.Context, guildID string) ([]domain.Deployment, error)
	GetBySubscriptionID(ctx context.Context, subID string) (*domain.Deployment, error)
	Update(ctx context.Context, dep *domain.Deployment) error
	List(ctx context.Context, limit, offset int) ([]domain.Deployment, error)
}

// TokenPoolRepository manages the pre-warmed pool of bot tokens for 0-setup provisioning.
type TokenPoolRepository interface {
	Add(ctx context.Context, entry *domain.BotTokenPoolEntry) error
	ClaimToken(ctx context.Context, botType, guildID, subscriptionID string) (*domain.BotTokenPoolEntry, error)
	ReleaseToken(ctx context.Context, guildID string) error
	GetAvailableCount(ctx context.Context, botType string) (int, error)
	GetPoolStats(ctx context.Context) (map[string]map[string]int, error)
	List(ctx context.Context, botType string, status domain.TokenStatus, limit, offset int) ([]domain.BotTokenPoolEntry, error)
}

// K8sOrchestrator abstracts the container runtime and Kubernetes cluster management.
type K8sOrchestrator interface {
	DeployBot(ctx context.Context, params BotDeployParams) error
	StopBot(ctx context.Context, namespace, deploymentName string) error
	RestartBot(ctx context.Context, namespace, deploymentName string) error
	GetBotStatus(ctx context.Context, namespace, deploymentName string) (domain.DeploymentStatus, error)
}

// DeploymentService defines the application orchestrator business operations.
type DeploymentService interface {
	ProvisionBot(ctx context.Context, subID, userID, guildID, botType, instanceLabel, botToken, imageTag string, isZeroSetup bool) (*domain.Deployment, error)
	StopBot(ctx context.Context, deploymentID string) error
	RestartBot(ctx context.Context, deploymentID string) error
	GetDeployment(ctx context.Context, id string) (*domain.Deployment, error)
	GetDeploymentByGuild(ctx context.Context, guildID string) (*domain.Deployment, error)
	GetDeploymentsByGuild(ctx context.Context, guildID string) ([]domain.Deployment, error)
	CustomizeBot(ctx context.Context, identifier, name, avatarURL string) error
	AddPoolTokens(ctx context.Context, botType string, tokens []AddTokenInput) (int, error)
	GetPoolStats(ctx context.Context) (map[string]map[string]int, error)
	CheckPoolAvailability(ctx context.Context, botType string) (bool, int, error)
}
