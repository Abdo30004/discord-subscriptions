package events_test

import (
	"testing"
	"time"

	"github.com/discord-subscriptions/shared/events"
)

func TestSubscriptionActivatedEvent(t *testing.T) {
	now := time.Now().UTC()
	subID := "sub-1234"
	userID := "user-discord-987"
	guildID := "guild-discord-456"
	botType := "music"
	planID := "pro-monthly"
	isDedicated := true
	provider := "paypal"

	event := events.NewSubscriptionActivatedEvent(
		subID, userID, guildID, botType, planID, "Main Stage", isDedicated, false, now.Add(30*24*time.Hour), provider,
	)

	if event.GetID() == "" {
		t.Fatal("expected event ID to be generated")
	}

	if event.GetType() != events.TypeSubscriptionActivated {
		t.Fatalf("expected event type %s, got %s", events.TypeSubscriptionActivated, event.GetType())
	}

	if event.SubscriptionID != subID {
		t.Fatalf("expected subscription ID %s, got %s", subID, event.SubscriptionID)
	}

	if !event.IsDedicated {
		t.Fatal("expected isDedicated to be true")
	}
}

func TestDeploymentRequestedEvent(t *testing.T) {
	event := events.NewDeploymentRequestedEvent(
		"dep-1", "sub-1", "user-1", "guild-1", "music", "Main Stage", "secret/data/bots/dep-1", "v1.0.0", true,
	)

	if event.GetType() != events.TypeDeploymentRequested {
		t.Fatalf("expected %s, got %s", events.TypeDeploymentRequested, event.GetType())
	}

	if event.VaultSecretPath != "secret/data/bots/dep-1" {
		t.Fatalf("unexpected vault secret path: %s", event.VaultSecretPath)
	}
}
