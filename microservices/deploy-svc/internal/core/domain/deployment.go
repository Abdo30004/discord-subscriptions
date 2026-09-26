package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidGuildID        = errors.New("guild ID is required")
	ErrInvalidSubscriptionID = errors.New("subscription ID is required")
	ErrInvalidBotType        = errors.New("bot type is required")
	ErrInvalidStateTransition= errors.New("illegal state transition")
)

type DeploymentStatus string

const (
	StatusPending   DeploymentStatus = "pending"
	StatusDeploying DeploymentStatus = "deploying"
	StatusRunning   DeploymentStatus = "running"
	StatusFailed    DeploymentStatus = "failed"
	StatusStopped   DeploymentStatus = "stopped"
)

// Deployment represents an instance of a Discord bot managed in Kubernetes.
type Deployment struct {
	ID                string           `json:"id"`
	SubscriptionID    string           `json:"subscription_id"`
	UserID            string           `json:"user_id"`
	GuildID           string           `json:"guild_id"`
	BotType           string           `json:"bot_type"`
	InstanceLabel     string           `json:"instance_label"`
	K8sNamespace      string           `json:"k8s_namespace"`
	K8sDeploymentName string           `json:"k8s_deployment_name"`
	ImageName         string           `json:"image_name"`
	ImageTag          string           `json:"image_tag"`
	Status            DeploymentStatus `json:"status"`
	IsZeroSetup       bool             `json:"is_zero_setup"`
	ClientID          string           `json:"client_id,omitempty"`
	ErrorMessage      string           `json:"error_message,omitempty"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// Validate verifies domain invariants.
func (d *Deployment) Validate() error {
	if d.GuildID == "" {
		return ErrInvalidGuildID
	}
	if d.SubscriptionID == "" {
		return ErrInvalidSubscriptionID
	}
	if d.BotType == "" {
		return ErrInvalidBotType
	}
	return nil
}

// CanTransitionTo enforces the state machine for deployment lifecycles.
func (d *Deployment) CanTransitionTo(target DeploymentStatus) bool {
	switch d.Status {
	case StatusPending:
		return target == StatusDeploying || target == StatusFailed
	case StatusDeploying:
		return target == StatusRunning || target == StatusFailed
	case StatusRunning:
		return target == StatusStopped || target == StatusDeploying || target == StatusFailed
	case StatusFailed:
		return target == StatusDeploying || target == StatusStopped
	case StatusStopped:
		return target == StatusDeploying
	default:
		return false
	}
}

// Transition moves the deployment to a new status if permitted.
func (d *Deployment) Transition(target DeploymentStatus) error {
	if !d.CanTransitionTo(target) {
		return ErrInvalidStateTransition
	}
	d.Status = target
	d.UpdatedAt = time.Now().UTC()
	return nil
}
