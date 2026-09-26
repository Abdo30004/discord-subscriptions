package events

const (
	TypeDeploymentRequested  EventType = "deployment.requested"
	TypeDeploymentCompleted  EventType = "deployment.completed"
	TypeDeploymentFailed     EventType = "deployment.failed"
	TypeDeploymentTerminated EventType = "deployment.terminated"
)

// DeploymentRequestedEvent is received by deploy-svc to trigger container orchestration.
type DeploymentRequestedEvent struct {
	BaseEvent
	DeploymentID    string `json:"deployment_id"`
	SubscriptionID  string `json:"subscription_id"`
	UserID          string `json:"user_id"`
	GuildID         string `json:"guild_id"`
	BotType         string `json:"bot_type"`
	InstanceLabel   string `json:"instance_label"`
	VaultSecretPath string `json:"vault_secret_path"` // Path where Bot Token is stored
	ImageTag        string `json:"image_tag"`
	IsZeroSetup     bool   `json:"is_zero_setup"`
}

// NewDeploymentRequestedEvent creates a ready-to-publish DeploymentRequestedEvent.
func NewDeploymentRequestedEvent(deploymentID, subID, userID, guildID, botType, instanceLabel, vaultSecretPath, imageTag string, isZeroSetup bool) DeploymentRequestedEvent {
	return DeploymentRequestedEvent{
		BaseEvent:       NewBaseEvent(TypeDeploymentRequested),
		DeploymentID:    deploymentID,
		SubscriptionID:  subID,
		UserID:          userID,
		GuildID:         guildID,
		BotType:         botType,
		InstanceLabel:   instanceLabel,
		VaultSecretPath: vaultSecretPath,
		ImageTag:        imageTag,
		IsZeroSetup:     isZeroSetup,
	}
}

// DeploymentCompletedEvent is emitted when a bot instance successfully starts and registers.
type DeploymentCompletedEvent struct {
	BaseEvent
	DeploymentID   string `json:"deployment_id"`
	SubscriptionID string `json:"subscription_id"`
	GuildID        string `json:"guild_id"`
	InstanceLabel  string `json:"instance_label"`
	PodName        string `json:"pod_name"`
	HealthCheckURL string `json:"health_check_url"`
	ClientID       string `json:"client_id,omitempty"`
}

// DeploymentFailedEvent is emitted when pod creation or initial health check fails.
type DeploymentFailedEvent struct {
	BaseEvent
	DeploymentID   string `json:"deployment_id"`
	SubscriptionID string `json:"subscription_id"`
	GuildID        string `json:"guild_id"`
	ErrorMessage   string `json:"error_message"`
}
