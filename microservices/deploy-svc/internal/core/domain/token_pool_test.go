package domain_test

import (
	"testing"

	"github.com/discord-subscriptions/deploy-svc/internal/core/domain"
)

func TestBotTokenPoolEntry_Lifecycle(t *testing.T) {
	entry := &domain.BotTokenPoolEntry{
		ID:       "pool-1",
		BotType:  "music",
		ClientID: "client-123",
		Status:   domain.TokenStatusAvailable,
	}

	if !entry.IsAvailable() {
		t.Fatal("expected entry to be available")
	}

	entry.Assign("guild-99", "sub-88")
	if entry.IsAvailable() {
		t.Fatal("expected entry not to be available after assign")
	}
	if entry.Status != domain.TokenStatusAssigned {
		t.Fatalf("expected assigned status, got %s", entry.Status)
	}
	if entry.AssignedGuildID != "guild-99" {
		t.Fatalf("expected guild-99, got %s", entry.AssignedGuildID)
	}

	entry.Quarantine()
	if entry.Status != domain.TokenStatusQuarantined {
		t.Fatalf("expected quarantined, got %s", entry.Status)
	}

	entry.Release()
	if !entry.IsAvailable() {
		t.Fatal("expected entry to be available after release")
	}
	if entry.AssignedGuildID != "" {
		t.Fatalf("expected empty guild ID, got %s", entry.AssignedGuildID)
	}
}
