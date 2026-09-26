package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/discord-subscriptions/monitor-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/events"
	"github.com/discord-subscriptions/shared/messaging"
)

type EventConsumer struct {
	service ports.MonitorService
	client  messaging.Subscriber
	logger  *slog.Logger
}

// NewEventConsumer constructs a subscriber for deployment and bot lifecycle events.
func NewEventConsumer(service ports.MonitorService, client messaging.Subscriber, logger *slog.Logger) *EventConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &EventConsumer{
		service: service,
		client:  client,
		logger:  logger,
	}
}

// Start registers subscriptions on RabbitMQ.
func (c *EventConsumer) Start(ctx context.Context) error {
	exchange := "discord.events"

	// 1. Listen for DeploymentCompletedEvent to register target
	err := c.client.Subscribe(ctx, "monitor-svc.deployment-completed", exchange, "deployment.completed", c.handleDeploymentCompleted)
	if err != nil {
		return fmt.Errorf("failed subscribing to deployment.completed: %w", err)
	}

	c.logger.Info("monitor-svc event consumer started listening for deployment events")
	return nil
}

func (c *EventConsumer) handleDeploymentCompleted(ctx context.Context, payload []byte) error {
	var evt events.DeploymentCompletedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("failed unmarshaling DeploymentCompletedEvent: %w", err)
	}

	c.logger.Info("registering newly deployed bot for health monitoring",
		slog.String("dep_id", evt.DeploymentID),
		slog.String("guild_id", evt.GuildID),
		slog.String("health_url", evt.HealthCheckURL),
	)

	_, err := c.service.RegisterTarget(ctx, evt.DeploymentID, evt.GuildID, evt.InstanceLabel, evt.HealthCheckURL, 15)
	return err
}
