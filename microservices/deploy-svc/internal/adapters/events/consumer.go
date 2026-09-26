package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/discord-subscriptions/deploy-svc/internal/core/ports"
	"github.com/discord-subscriptions/shared/events"
	"github.com/discord-subscriptions/shared/messaging"
)

type EventConsumer struct {
	subscriber ports.DeploymentService
	client     messaging.Subscriber
	logger     *slog.Logger
}

// NewEventConsumer creates an event listener for deployment and subscription events.
func NewEventConsumer(service ports.DeploymentService, client messaging.Subscriber, logger *slog.Logger) *EventConsumer {
	if logger == nil {
		logger = slog.Default()
	}
	return &EventConsumer{
		subscriber: service,
		client:     client,
		logger:     logger,
	}
}

// Start registers subscriptions for deployment requests and cancellations.
func (c *EventConsumer) Start(ctx context.Context) error {
	exchange := "discord.events"

	// 1. Listen for DeploymentRequestedEvent
	err := c.client.Subscribe(ctx, "deploy-svc.deployment-requested", exchange, "deployment.requested", c.handleDeploymentRequested)
	if err != nil {
		return fmt.Errorf("failed subscribing to deployment.requested: %w", err)
	}

	// 2. Listen for SubscriptionCancelledEvent
	err = c.client.Subscribe(ctx, "deploy-svc.subscription-cancelled", exchange, "subscription.cancelled", c.handleSubscriptionCancelled)
	if err != nil {
		return fmt.Errorf("failed subscribing to subscription.cancelled: %w", err)
	}

	c.logger.Info("event consumer started listening for deployment and cancellation events")
	return nil
}

func (c *EventConsumer) handleDeploymentRequested(ctx context.Context, payload []byte) error {
	var evt events.DeploymentRequestedEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("failed unmarshaling DeploymentRequestedEvent: %w", err)
	}

	c.logger.Info("received deployment requested event",
		slog.String("sub_id", evt.SubscriptionID),
		slog.String("guild_id", evt.GuildID),
		slog.String("bot_type", evt.BotType),
	)

	_, err := c.subscriber.ProvisionBot(
		ctx,
		evt.SubscriptionID,
		evt.UserID,
		evt.GuildID,
		evt.BotType,
		evt.InstanceLabel,
		"", // token is retrieved from vault or injected later
		evt.ImageTag,
		evt.IsZeroSetup,
	)
	return err
}

func (c *EventConsumer) handleSubscriptionCancelled(ctx context.Context, payload []byte) error {
	var evt events.SubscriptionCancelledEvent
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("failed unmarshaling SubscriptionCancelledEvent: %w", err)
	}

	c.logger.Info("received subscription cancelled event, stopping bot",
		slog.String("sub_id", evt.SubscriptionID),
		slog.String("guild_id", evt.GuildID),
	)

	dep, err := c.subscriber.GetDeploymentByGuild(ctx, evt.GuildID)
	if err != nil {
		c.logger.Warn("no active deployment found for cancelled guild", slog.String("guild_id", evt.GuildID))
		return nil
	}

	return c.subscriber.StopBot(ctx, dep.ID)
}
