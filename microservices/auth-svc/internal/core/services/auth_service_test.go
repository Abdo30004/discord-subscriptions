package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
	"github.com/discord-subscriptions/auth-svc/internal/core/services"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func (m *mockUserRepo) Upsert(ctx context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return m.users[id], nil
}

type mockDiscordClient struct {
	guilds []domain.Guild
}

func (m *mockDiscordClient) GetOAuthURL(redirectURI, state string) string {
	return "https://discord.com/oauth2/authorize?client_id=123"
}

func (m *mockDiscordClient) ExchangeCode(ctx context.Context, code, redirectURI string) (string, string, int, error) {
	return "mock_access_token", "mock_refresh_token", 3600, nil
}

func (m *mockDiscordClient) GetUserProfile(ctx context.Context, accessToken string) (*domain.User, error) {
	return &domain.User{
		ID:       "discord-12345",
		Username: "AntigravityUser",
	}, nil
}

func (m *mockDiscordClient) GetUserGuilds(ctx context.Context, accessToken string) ([]domain.Guild, error) {
	return m.guilds, nil
}

type mockTokenMgr struct{}

func (m *mockTokenMgr) GenerateToken(user domain.User) (string, time.Time, error) {
	return "signed.jwt.token", time.Now().UTC().Add(7 * 24 * time.Hour), nil
}

func (m *mockTokenMgr) ValidateToken(tokenString string) (*domain.UserClaims, error) {
	return &domain.UserClaims{
		UserID:   "discord-12345",
		Username: "AntigravityUser",
	}, nil
}

func TestAuthService_AuthenticateWithCode(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*domain.User)}
	discordMock := &mockDiscordClient{}
	tokenMock := &mockTokenMgr{}

	svc := services.NewAuthService(repo, discordMock, tokenMock, nil)

	session, err := svc.AuthenticateWithCode(context.Background(), "auth_code_xyz", "http://localhost:3000/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if session.Token != "signed.jwt.token" {
		t.Fatalf("expected signed token, got %s", session.Token)
	}

	if session.User.Username != "AntigravityUser" {
		t.Fatalf("expected username AntigravityUser, got %s", session.User.Username)
	}

	if repo.users["discord-12345"] == nil {
		t.Fatal("expected user to be saved in repository")
	}
}

func TestAuthService_GetUserManageableGuilds(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*domain.User{
			"discord-12345": {
				ID:          "discord-12345",
				Username:    "AntigravityUser",
				AccessToken: "token123",
			},
		},
	}

	discordMock := &mockDiscordClient{
		guilds: []domain.Guild{
			{ID: "g1", Name: "Admin Server", Permissions: "8"},        // Administrator -> CAN manage
			{ID: "g2", Name: "Regular Chat", Permissions: "2048"},     // Regular member -> CANNOT manage
			{ID: "g3", Name: "My Server", Owner: true, Permissions: "0"}, // Owner -> CAN manage
		},
	}

	svc := services.NewAuthService(repo, discordMock, &mockTokenMgr{}, nil)

	manageable, err := svc.GetUserManageableGuilds(context.Background(), "discord-12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(manageable) != 2 {
		t.Fatalf("expected 2 manageable guilds, got %d", len(manageable))
	}

	if manageable[0].ID != "g1" || manageable[1].ID != "g3" {
		t.Fatalf("unexpected filtered guilds: %v", manageable)
	}
}
