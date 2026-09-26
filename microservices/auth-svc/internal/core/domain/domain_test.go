package domain_test

import (
	"testing"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
)

func TestUser_Validate(t *testing.T) {
	valid := domain.User{
		ID:       "123456789",
		Username: "TestUser",
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid user, got %v", err)
	}

	invalidID := domain.User{Username: "NoID"}
	if err := invalidID.Validate(); err != domain.ErrInvalidUserID {
		t.Fatalf("expected ErrInvalidUserID, got %v", err)
	}

	invalidName := domain.User{ID: "123456"}
	if err := invalidName.Validate(); err != domain.ErrInvalidUsername {
		t.Fatalf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestUser_AvatarURL(t *testing.T) {
	withAvatar := domain.User{ID: "999", Avatar: "abc"}
	if withAvatar.AvatarURL() != "https://cdn.discordapp.com/avatars/999/abc.png" {
		t.Fatalf("unexpected avatar URL: %s", withAvatar.AvatarURL())
	}

	withoutAvatar := domain.User{ID: "999"}
	if withoutAvatar.AvatarURL() != "https://cdn.discordapp.com/embed/avatars/0.png" {
		t.Fatalf("unexpected default avatar URL: %s", withoutAvatar.AvatarURL())
	}
}

func TestGuild_ComputeCanManage(t *testing.T) {
	// Server Owner can always manage
	ownerGuild := domain.Guild{Owner: true, Permissions: "0"}
	if !ownerGuild.ComputeCanManage() {
		t.Fatal("expected server owner to be able to manage")
	}

	// Administrator permission (0x8 = 8)
	adminGuild := domain.Guild{Owner: false, Permissions: "8"}
	if !adminGuild.ComputeCanManage() {
		t.Fatal("expected administrator to be able to manage")
	}

	// Manage Guild permission (0x20 = 32)
	manageGuild := domain.Guild{Owner: false, Permissions: "32"}
	if !manageGuild.ComputeCanManage() {
		t.Fatal("expected manage guild permission to be able to manage")
	}

	// Regular member with only Send Messages (0x800 = 2048)
	memberGuild := domain.Guild{Owner: false, Permissions: "2048"}
	if memberGuild.ComputeCanManage() {
		t.Fatal("expected regular member not to be able to manage")
	}
}
