package services_test

import (
	"context"
	"errors"
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
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *mockUserRepo) ListAdmins(ctx context.Context) ([]domain.User, error) {
	var list []domain.User
	for _, u := range m.users {
		if u.IsAdmin {
			list = append(list, *u)
		}
	}
	return list, nil
}

func (m *mockUserRepo) SearchUsers(ctx context.Context, query string, limit int) ([]domain.User, error) {
	var list []domain.User
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, nil
}

func (m *mockUserRepo) SetAdmin(ctx context.Context, discordID string, isAdmin bool, promotedBy string) error {
	u, ok := m.users[discordID]
	if !ok {
		return errors.New("user not found")
	}
	u.IsAdmin = isAdmin
	u.AdminPromotedBy = promotedBy
	return nil
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

	svc := services.NewAuthService(repo, discordMock, tokenMock, []string{"super-admin-999"}, nil)

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

	svc := services.NewAuthService(repo, discordMock, &mockTokenMgr{}, []string{"super-admin-999"}, nil)

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

func TestAuthService_SuperAdminAndPromoteFlow(t *testing.T) {
	superAdminID := "super-111"
	normalUserID := "user-222"
	repo := &mockUserRepo{
		users: map[string]*domain.User{
			superAdminID: {
				ID:       superAdminID,
				Username: "SuperOwner",
			},
			normalUserID: {
				ID:       normalUserID,
				Username: "ModeratorBob",
			},
		},
	}

	svc := services.NewAuthService(repo, &mockDiscordClient{}, &mockTokenMgr{}, []string{superAdminID}, nil)

	// Verify super admin check
	if !svc.IsSuperAdmin(superAdminID) {
		t.Fatalf("expected %s to be recognized as super admin", superAdminID)
	}
	if svc.IsSuperAdmin(normalUserID) {
		t.Fatalf("expected %s NOT to be recognized as super admin", normalUserID)
	}

	// Normal user attempts to promote someone -> should be rejected
	err := svc.PromoteAdmin(context.Background(), normalUserID, normalUserID)
	if err == nil {
		t.Fatalf("expected unauthorized error when non-super admin attempts promotion")
	}

	// Super admin promotes normal user -> should succeed
	err = svc.PromoteAdmin(context.Background(), superAdminID, normalUserID)
	if err != nil {
		t.Fatalf("super admin promotion failed: %v", err)
	}

	if !repo.users[normalUserID].IsAdmin {
		t.Fatalf("expected user %s to be marked as admin", normalUserID)
	}

	// Verify ListAdmins
	admins, err := svc.ListAdmins(context.Background(), superAdminID)
	if err != nil {
		t.Fatalf("failed listing admins: %v", err)
	}
	if len(admins) < 2 {
		t.Fatalf("expected at least 2 admins in list (super + promoted), got %d", len(admins))
	}

	// Super admin revokes normal user
	err = svc.RevokeAdmin(context.Background(), superAdminID, normalUserID)
	if err != nil {
		t.Fatalf("super admin revoking failed: %v", err)
	}
	if repo.users[normalUserID].IsAdmin {
		t.Fatalf("expected user %s admin privileges to be revoked", normalUserID)
	}

	// Attempting to revoke super admin should fail
	err = svc.RevokeAdmin(context.Background(), superAdminID, superAdminID)
	if err == nil {
		t.Fatalf("expected error when trying to revoke super admin")
	}
}
