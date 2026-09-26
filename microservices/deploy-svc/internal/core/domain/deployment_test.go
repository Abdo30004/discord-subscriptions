package domain_test

import (
	"testing"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
)

func TestDeployment_Validate(t *testing.T) {
	tests := []struct {
		name    string
		dep     domain.Deployment
		wantErr bool
	}{
		{
			name: "valid deployment",
			dep: domain.Deployment{
				GuildID:        "guild-123",
				SubscriptionID: "sub-456",
				BotType:        "music",
			},
			wantErr: false,
		},
		{
			name: "missing guild id",
			dep: domain.Deployment{
				SubscriptionID: "sub-456",
				BotType:        "music",
			},
			wantErr: true,
		},
		{
			name: "missing subscription id",
			dep: domain.Deployment{
				GuildID: "guild-123",
				BotType: "music",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dep.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestDeployment_StateMachine(t *testing.T) {
	dep := domain.Deployment{
		Status: domain.StatusPending,
	}

	// Valid transition: Pending -> Deploying
	if err := dep.Transition(domain.StatusDeploying); err != nil {
		t.Fatalf("expected valid transition from pending to deploying: %v", err)
	}

	// Valid transition: Deploying -> Running
	if err := dep.Transition(domain.StatusRunning); err != nil {
		t.Fatalf("expected valid transition from deploying to running: %v", err)
	}

	// Invalid transition: Running -> Pending
	if err := dep.Transition(domain.StatusPending); err == nil {
		t.Fatal("expected invalid transition error when transitioning from running to pending")
	}

	// Valid transition: Running -> Stopped
	if err := dep.Transition(domain.StatusStopped); err != nil {
		t.Fatalf("expected valid transition from running to stopped: %v", err)
	}
}
