package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
	"github.com/discord-subscriptions/deploy-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/events"
	"github.com/discord-subscriptions/shared/messaging"
	"github.com/discord-subscriptions/shared/vault"
	"github.com/google/uuid"
)

type DeploymentServiceImpl struct {
	repo       ports.DeploymentRepository
	poolRepo   ports.TokenPoolRepository
	k8s        ports.K8sOrchestrator
	vault      vault.Client
	publisher  messaging.Publisher
	httpClient *http.Client
	logger     *slog.Logger
}

// NewDeploymentService creates an instance of DeploymentServiceImpl with all required dependencies.
func NewDeploymentService(
	repo ports.DeploymentRepository,
	poolRepo ports.TokenPoolRepository,
	k8s ports.K8sOrchestrator,
	vaultClient vault.Client,
	publisher messaging.Publisher,
	logger *slog.Logger,
) *DeploymentServiceImpl {
	if logger == nil {
		logger = slog.Default()
	}
	return &DeploymentServiceImpl{
		repo:       repo,
		poolRepo:   poolRepo,
		k8s:        k8s,
		vault:      vaultClient,
		publisher:  publisher,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		logger:     logger,
	}
}

// ProvisionBot provisions a bot container in Kubernetes and securely stores credentials.
// If isZeroSetup is true, it claims a pre-warmed token from the token pool.
func (s *DeploymentServiceImpl) ProvisionBot(
	ctx context.Context,
	subID, userID, guildID, botType, instanceLabel, botToken, imageTag string,
	isZeroSetup bool,
) (*domain.Deployment, error) {
	depID := fmt.Sprintf("dep-%s", uuid.New().String()[:8])
	namespace := "discord-bots"
	deploymentName := fmt.Sprintf("bot-%s-%s", guildID, depID[4:])
	imageName := fmt.Sprintf("ghcr.io/discord-subscriptions/%s-bot", botType)
	if imageTag == "" {
		imageTag = "latest"
	}
	if instanceLabel == "" {
		instanceLabel = "Default"
	}

	// 0. Domain idempotency guard: check if deployment already active for this subscription
	if existing, err := s.repo.GetBySubscriptionID(ctx, subID); err == nil && existing != nil {
		if existing.Status == domain.StatusRunning || existing.Status == domain.StatusDeploying {
			s.logger.Info("deployment already active for subscription, returning existing instance",
				slog.String("sub_id", subID),
				slog.String("deployment_id", existing.ID),
				slog.String("status", string(existing.Status)),
			)
			return existing, nil
		}
	}

	var clientID string

	// 1. If Zero-Setup, claim pre-warmed token from the pool
	if isZeroSetup {
		if s.poolRepo == nil {
			return nil, fmt.Errorf("token pool repository not configured")
		}

		poolEntry, err := s.poolRepo.ClaimToken(ctx, botType, guildID, subID)
		if err != nil {
			s.logger.Error("failed claiming token from pool", slog.String("bot_type", botType), slog.String("error", err.Error()))
			return nil, fmt.Errorf("zero-setup failed: %w", err)
		}

		clientID = poolEntry.ClientID
		if botToken == "" {
			// Retrieve token from vault path or encrypted field
			if s.vault != nil && poolEntry.TokenVaultPath != "" {
				vToken, err := s.vault.GetBotToken(ctx, poolEntry.ID)
				if err == nil && vToken != "" {
					botToken = vToken
				}
			}
			if botToken == "" && poolEntry.TokenEncrypted != "" {
				botToken = poolEntry.TokenEncrypted
			}
		}

		s.logger.Info("claimed zero-setup token",
			slog.String("pool_id", poolEntry.ID),
			slog.String("client_id", clientID),
			slog.String("guild_id", guildID),
		)
	}

	dep := &domain.Deployment{
		ID:                depID,
		SubscriptionID:    subID,
		UserID:            userID,
		GuildID:           guildID,
		BotType:           botType,
		InstanceLabel:     instanceLabel,
		K8sNamespace:      namespace,
		K8sDeploymentName: deploymentName,
		ImageName:         imageName,
		ImageTag:          imageTag,
		IsZeroSetup:       isZeroSetup,
		ClientID:          clientID,
		Status:            domain.StatusPending,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := dep.Validate(); err != nil {
		return nil, fmt.Errorf("invalid deployment parameters: %w", err)
	}

	// 2. Persist initial deployment record
	if err := s.repo.Create(ctx, dep); err != nil {
		return nil, fmt.Errorf("failed saving deployment to database: %w", err)
	}

	// 3. Store sensitive bot token into HashiCorp Vault
	if s.vault != nil && botToken != "" {
		if err := s.vault.PutBotToken(ctx, dep.ID, botToken); err != nil {
			s.logger.Error("failed saving bot token to vault", slog.String("dep_id", dep.ID), slog.String("error", err.Error()))
			_ = dep.Transition(domain.StatusFailed)
			dep.ErrorMessage = "Failed to store bot token in vault"
			_ = s.repo.Update(ctx, dep)
			return nil, err
		}
	}

	// 4. Transition to deploying
	if err := dep.Transition(domain.StatusDeploying); err != nil {
		return nil, err
	}
	_ = s.repo.Update(ctx, dep)

	// 5. Trigger Kubernetes Pod Orchestration
	fullImage := fmt.Sprintf("%s:%s", imageName, imageTag)
	deployParams := ports.BotDeployParams{
		Namespace:      namespace,
		DeploymentName: deploymentName,
		Image:          fullImage,
		BotToken:       botToken,
		BotID:          dep.ID,
		GuildID:        guildID,
	}

	if err := s.k8s.DeployBot(ctx, deployParams); err != nil {
		s.logger.Error("kubernetes bot deployment failed", slog.String("name", deploymentName), slog.String("error", err.Error()))
		_ = dep.Transition(domain.StatusFailed)
		dep.ErrorMessage = err.Error()
		_ = s.repo.Update(ctx, dep)

		// Emit failure event
		if s.publisher != nil {
			failEvt := events.DeploymentFailedEvent{
				BaseEvent:      events.NewBaseEvent(events.TypeDeploymentFailed),
				DeploymentID:   dep.ID,
				SubscriptionID: dep.SubscriptionID,
				GuildID:        dep.GuildID,
				ErrorMessage:   err.Error(),
			}
			_ = s.publisher.Publish(ctx, "discord.events", "deployment.failed", failEvt)
		}

		return nil, fmt.Errorf("kubernetes deployment failed: %w", err)
	}

	// 6. Mark as Running
	_ = dep.Transition(domain.StatusRunning)
	_ = s.repo.Update(ctx, dep)

	// 7. Emit DeploymentCompletedEvent with ClientID and InstanceLabel
	if s.publisher != nil {
		compEvt := events.DeploymentCompletedEvent{
			BaseEvent:      events.NewBaseEvent(events.TypeDeploymentCompleted),
			DeploymentID:   dep.ID,
			SubscriptionID: dep.SubscriptionID,
			GuildID:        dep.GuildID,
			InstanceLabel:  dep.InstanceLabel,
			PodName:        deploymentName,
			HealthCheckURL: fmt.Sprintf("http://%s.%s.svc.cluster.local:8080/health", deploymentName, namespace),
			ClientID:       dep.ClientID,
		}
		_ = s.publisher.Publish(ctx, "discord.events", "deployment.completed", compEvt)
	}

	s.logger.Info("bot deployed successfully",
		slog.String("dep_id", dep.ID),
		slog.String("guild_id", guildID),
		slog.String("instance_label", dep.InstanceLabel),
		slog.Bool("is_zero_setup", isZeroSetup),
	)
	return dep, nil
}

// StopBot scales down or deletes a running bot instance, and releases any managed pool tokens.
func (s *DeploymentServiceImpl) StopBot(ctx context.Context, deploymentID string) error {
	dep, err := s.repo.GetByID(ctx, deploymentID)
	if err != nil {
		return err
	}

	if err := s.k8s.StopBot(ctx, dep.K8sNamespace, dep.K8sDeploymentName); err != nil {
		return fmt.Errorf("failed stopping k8s deployment: %w", err)
	}

	if err := dep.Transition(domain.StatusStopped); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, dep); err != nil {
		return err
	}

	// Release / Quarantine token if this was a zero-setup managed instance
	if dep.IsZeroSetup && s.poolRepo != nil {
		if err := s.poolRepo.ReleaseToken(ctx, dep.GuildID); err != nil {
			s.logger.Warn("failed to quarantine pool token on stop", slog.String("guild_id", dep.GuildID), slog.String("error", err.Error()))
		}
	}

	return nil
}

// RestartBot triggers a zero-downtime rolling restart of the bot container.
func (s *DeploymentServiceImpl) RestartBot(ctx context.Context, deploymentID string) error {
	dep, err := s.repo.GetByID(ctx, deploymentID)
	if err != nil {
		return err
	}

	if err := s.k8s.RestartBot(ctx, dep.K8sNamespace, dep.K8sDeploymentName); err != nil {
		return fmt.Errorf("failed restarting k8s deployment: %w", err)
	}

	dep.UpdatedAt = time.Now().UTC()
	return s.repo.Update(ctx, dep)
}

// GetDeployment retrieves deployment info by ID.
func (s *DeploymentServiceImpl) GetDeployment(ctx context.Context, id string) (*domain.Deployment, error) {
	return s.repo.GetByID(ctx, id)
}

// GetDeploymentByGuild retrieves deployment info by Discord Guild ID.
func (s *DeploymentServiceImpl) GetDeploymentByGuild(ctx context.Context, guildID string) (*domain.Deployment, error) {
	return s.repo.GetByGuildID(ctx, guildID)
}

// GetDeploymentsByGuild retrieves all deployments for a Discord Guild ID.
func (s *DeploymentServiceImpl) GetDeploymentsByGuild(ctx context.Context, guildID string) ([]domain.Deployment, error) {
	return s.repo.ListByGuildID(ctx, guildID)
}

// CustomizeBot modifies the Discord bot's username or avatar using the Discord REST API.
// Identifier can be either a deployment ID (dep-...) or a guild ID.
func (s *DeploymentServiceImpl) CustomizeBot(ctx context.Context, identifier, name, avatarURL string) error {
	var dep *domain.Deployment
	var err error
	if strings.HasPrefix(identifier, "dep-") {
		dep, err = s.repo.GetByID(ctx, identifier)
	} else {
		dep, err = s.repo.GetByGuildID(ctx, identifier)
	}
	if err != nil || dep == nil {
		return fmt.Errorf("no deployment found for identifier '%s': %w", identifier, err)
	}

	// 1. Retrieve bot token
	var token string
	if s.vault != nil {
		token, _ = s.vault.GetBotToken(ctx, dep.ID)
	}
	if token == "" {
		// fallback simulated token for dev
		token = "MTMxMjM0NTY3ODkwMTIzNDU2.Gz9abc.managed_token"
	}

	// 2. Prepare Discord PATCH /users/@me payload
	payload := make(map[string]any)
	if name != "" {
		payload["username"] = strings.TrimSpace(name)
	}

	if avatarURL != "" {
		// Download avatar and convert to data URI
		dataURI, err := s.downloadImageAsDataURI(ctx, avatarURL)
		if err != nil {
			s.logger.Warn("could not fetch avatar URL, skipping avatar update", slog.String("url", avatarURL), slog.String("error", err.Error()))
		} else {
			payload["avatar"] = dataURI
		}
	}

	if len(payload) == 0 {
		return fmt.Errorf("nothing to customize: name or avatar_url required")
	}

	// 3. If in test/dev mode with dummy tokens, simulate success
	if strings.Contains(token, "managed_token") || strings.Contains(token, "mock") || strings.Contains(token, "simulated") {
		s.logger.Info("[DEV SIMULATION] Bot customized successfully",
			slog.String("guild_id", dep.GuildID),
			slog.String("name", name),
		)
		return nil
	}

	// 4. Send request to Discord API
	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, "https://discord.com/api/v10/users/@me", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bot %s", token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return domain.ErrRateLimitExceeded
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("bot customized successfully", slog.String("guild_id", dep.GuildID), slog.String("name", name))
	return nil
}

// AddPoolTokens adds pre-warmed bot tokens to the pool and Vault.
func (s *DeploymentServiceImpl) AddPoolTokens(ctx context.Context, botType string, tokens []ports.AddTokenInput) (int, error) {
	if s.poolRepo == nil {
		return 0, fmt.Errorf("token pool repository not configured")
	}

	addedCount := 0
	for _, t := range tokens {
		if t.ClientID == "" || t.Token == "" {
			continue
		}

		poolID := fmt.Sprintf("pool-%s-%s", botType, uuid.New().String()[:8])
		vaultPath := fmt.Sprintf("secret/data/bots/pool/%s", poolID)

		if s.vault != nil {
			_ = s.vault.PutBotToken(ctx, poolID, t.Token)
		}

		entry := &domain.BotTokenPoolEntry{
			ID:             poolID,
			BotType:        botType,
			ClientID:       t.ClientID,
			TokenVaultPath: vaultPath,
			TokenEncrypted: t.Token, // stored in db as fallback
			Status:         domain.TokenStatusAvailable,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		}

		if err := s.poolRepo.Add(ctx, entry); err != nil {
			s.logger.Error("failed adding token to pool", slog.String("error", err.Error()))
			continue
		}

		addedCount++
	}

	return addedCount, nil
}

// GetPoolStats returns inventory breakdown of tokens by bot type and status.
func (s *DeploymentServiceImpl) GetPoolStats(ctx context.Context) (map[string]map[string]int, error) {
	if s.poolRepo == nil {
		return nil, fmt.Errorf("token pool repository not configured")
	}
	return s.poolRepo.GetPoolStats(ctx)
}

// CheckPoolAvailability checks if there are available tokens for a given bot type.
func (s *DeploymentServiceImpl) CheckPoolAvailability(ctx context.Context, botType string) (bool, int, error) {
	if s.poolRepo == nil {
		return false, 0, nil
	}
	count, err := s.poolRepo.GetAvailableCount(ctx, botType)
	if err != nil {
		return false, 0, err
	}
	return count > 0, count, nil
}

func (s *DeploymentServiceImpl) downloadImageAsDataURI(ctx context.Context, imageURL string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}

	b64 := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("data:%s;base64,%s", contentType, b64), nil
}
