package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
	"github.com/discord-subscriptions/auth-svc/internal/core/ports"
	sharedErrors "github.com/discord-subscriptions/shared/errors"
	"github.com/discord-subscriptions/shared/vault"
	"github.com/google/uuid"
)

type AuthServiceImpl struct {
	userRepo      ports.UserRepository
	discord       ports.DiscordClient
	tokenMgr      ports.TokenManager
	vaultClient   vault.Client
	superAdminIDs map[string]struct{}
	logger        *slog.Logger
}

func NewAuthService(
	userRepo ports.UserRepository,
	discord ports.DiscordClient,
	tokenMgr ports.TokenManager,
	vaultClient vault.Client,
	superAdminIDs []string,
	logger *slog.Logger,
) *AuthServiceImpl {
	if logger == nil {
		logger = slog.Default()
	}
	superMap := make(map[string]struct{}, len(superAdminIDs))
	for _, id := range superAdminIDs {
		superMap[id] = struct{}{}
	}
	return &AuthServiceImpl{
		userRepo:      userRepo,
		discord:       discord,
		tokenMgr:      tokenMgr,
		vaultClient:   vaultClient,
		superAdminIDs: superMap,
		logger:        logger,
	}
}

// IsSuperAdmin checks if the given Discord ID is configured as Super Admin in the environment.
func (s *AuthServiceImpl) IsSuperAdmin(discordID string) bool {
	if discordID == "" {
		return false
	}
	_, ok := s.superAdminIDs[discordID]
	return ok
}

func (s *AuthServiceImpl) hydrateUserRoles(user *domain.User) {
	if user == nil {
		return
	}
	if s.IsSuperAdmin(user.ID) {
		user.IsSuperAdmin = true
		user.IsAdmin = true
	}
}

// GetOAuthURL returns the Discord authorization URL for login.
func (s *AuthServiceImpl) GetOAuthURL(redirectURI string) string {
	state := uuid.New().String()
	return s.discord.GetOAuthURL(redirectURI, state)
}

// AuthenticateWithCode exchanges Discord authorization code, saves user, and issues a JWT.
func (s *AuthServiceImpl) AuthenticateWithCode(ctx context.Context, code, redirectURI string) (*domain.AuthSession, error) {
	// 1. Exchange code for Discord tokens
	accessToken, refreshToken, expiresIn, err := s.discord.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("oauth exchange failed: %w", err)
	}

	// 2. Fetch Discord Profile
	discordUser, err := s.discord.GetUserProfile(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed fetching user profile: %w", err)
	}

	// Preserve existing is_admin status if already in database
	existing, _ := s.userRepo.GetByID(ctx, discordUser.ID)
	if existing != nil && existing.IsAdmin {
		discordUser.IsAdmin = true
		discordUser.AdminPromotedAt = existing.AdminPromotedAt
		discordUser.AdminPromotedBy = existing.AdminPromotedBy
	}

	// Hydrate super admin status from environment
	s.hydrateUserRoles(discordUser)

	now := time.Now().UTC()
	tokenExpiresAt := now.Add(time.Duration(expiresIn) * time.Second)
	discordUser.AccessToken = accessToken
	discordUser.RefreshToken = refreshToken
	discordUser.TokenExpiresAt = tokenExpiresAt
	discordUser.UpdatedAt = now
	discordUser.CreatedAt = now

	// 3. Securely store OAuth tokens in HashiCorp Vault KV v2 (never in plaintext PostgreSQL)
	if s.vaultClient != nil {
		if err := s.vaultClient.PutUserTokens(ctx, discordUser.ID, accessToken, refreshToken, tokenExpiresAt); err != nil {
			s.logger.Error("failed storing user oauth tokens in vault", slog.String("error", err.Error()), slog.String("user_id", discordUser.ID))
		} else {
			discordUser.HasOAuthToken = true
		}
	}

	// 4. Upsert user into database
	if err := s.userRepo.Upsert(ctx, discordUser); err != nil {
		return nil, fmt.Errorf("failed saving user profile: %w", err)
	}

	// 5. Generate JWT
	jwtToken, expiresAt, err := s.tokenMgr.GenerateToken(*discordUser)
	if err != nil {
		return nil, fmt.Errorf("failed generating auth token: %w", err)
	}

	s.logger.Info("user logged in successfully via discord oauth2",
		slog.String("user_id", discordUser.ID),
		slog.String("username", discordUser.Username),
		slog.Bool("is_admin", discordUser.IsAdmin),
		slog.Bool("is_super_admin", discordUser.IsSuperAdmin),
		slog.Bool("vault_stored", discordUser.HasOAuthToken),
	)

	return &domain.AuthSession{
		Token:     jwtToken,
		ExpiresAt: expiresAt,
		User:      *discordUser,
	}, nil
}

// RefreshSession validates an existing (or recently expired) JWT and issues a refreshed session,
// transparently refreshing Discord OAuth tokens via HashiCorp Vault if needed.
func (s *AuthServiceImpl) RefreshSession(ctx context.Context, currentToken string) (*domain.AuthSession, error) {
	if currentToken == "" {
		return nil, sharedErrors.ErrUnauthorized
	}

	// Allow tokens expired within the past 7 days to be refreshed
	claims, err := s.tokenMgr.ValidateTokenAllowExpired(currentToken, 7*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("invalid session token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if Discord tokens in Vault need refreshing
	if s.vaultClient != nil {
		_, refreshToken, expiresAt, err := s.vaultClient.GetUserTokens(ctx, user.ID)
		if err == nil && refreshToken != "" {
			// If Discord token is expiring soon (within 30 mins) or already expired, refresh it
			if time.Now().UTC().Add(30 * time.Minute).After(expiresAt) {
				newAccess, newRefresh, expiresIn, refreshErr := s.discord.RefreshToken(ctx, refreshToken)
				if refreshErr == nil {
					newExp := time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)
					_ = s.vaultClient.PutUserTokens(ctx, user.ID, newAccess, newRefresh, newExp)
					user.HasOAuthToken = true
				} else {
					s.logger.Warn("discord oauth refresh failed during session refresh",
						slog.String("user_id", user.ID),
						slog.String("error", refreshErr.Error()),
					)
				}
			}
		}
	}

	s.hydrateUserRoles(user)

	newToken, expiresAt, err := s.tokenMgr.GenerateToken(*user)
	if err != nil {
		return nil, fmt.Errorf("failed generating refreshed token: %w", err)
	}

	s.logger.Info("user session refreshed successfully",
		slog.String("user_id", user.ID),
		slog.String("username", user.Username),
	)

	return &domain.AuthSession{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User:      *user,
	}, nil
}

// GetUserSession validates a JWT and retrieves the user record.
func (s *AuthServiceImpl) GetUserSession(ctx context.Context, token string) (*domain.User, error) {
	claims, err := s.tokenMgr.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	s.hydrateUserRoles(user)
	return user, nil
}

// GetUserManageableGuilds retrieves only the guilds where the user has bot-installation permissions.
func (s *AuthServiceImpl) GetUserManageableGuilds(ctx context.Context, userID string) ([]domain.Guild, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	accessToken := user.AccessToken
	if accessToken == "" && s.vaultClient != nil {
		vaultAccess, _, _, err := s.vaultClient.GetUserTokens(ctx, user.ID)
		if err == nil && vaultAccess != "" {
			accessToken = vaultAccess
		}
	}
	if accessToken == "" {
		return nil, fmt.Errorf("no valid discord access token found for user %s", userID)
	}

	allGuilds, err := s.discord.GetUserGuilds(ctx, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed fetching user guilds from discord: %w", err)
	}

	// Filter down to only servers the user has permission to manage
	manageable := make([]domain.Guild, 0, len(allGuilds))
	for _, g := range allGuilds {
		if g.ComputeCanManage() {
			manageable = append(manageable, g)
		}
	}

	return manageable, nil
}

// ListAdmins returns all platform administrators.
func (s *AuthServiceImpl) ListAdmins(ctx context.Context, requestingUserID string) ([]domain.User, error) {
	caller, err := s.userRepo.GetByID(ctx, requestingUserID)
	if err != nil {
		return nil, sharedErrors.ErrNotFound
	}
	s.hydrateUserRoles(caller)
	if !caller.IsAdmin && !caller.IsSuperAdmin {
		return nil, sharedErrors.ErrUnauthorized
	}

	dbAdmins, err := s.userRepo.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}

	adminMap := make(map[string]domain.User)
	for _, a := range dbAdmins {
		s.hydrateUserRoles(&a)
		adminMap[a.ID] = a
	}

	// Ensure any Super Admin configured via environment is present in the list
	for superID := range s.superAdminIDs {
		if _, exists := adminMap[superID]; !exists {
			if u, err := s.userRepo.GetByID(ctx, superID); err == nil && u != nil {
				s.hydrateUserRoles(u)
				adminMap[superID] = *u
			} else {
				adminMap[superID] = domain.User{
					ID:           superID,
					Username:     "Super Admin (Configured via ENV)",
					IsAdmin:      true,
					IsSuperAdmin: true,
				}
			}
		}
	}

	result := make([]domain.User, 0, len(adminMap))
	for _, a := range adminMap {
		result = append(result, a)
	}
	return result, nil
}

// SearchUsers allows a Super Admin to find registered users by Discord ID or Username.
func (s *AuthServiceImpl) SearchUsers(ctx context.Context, requestingUserID, query string) ([]domain.User, error) {
	if !s.IsSuperAdmin(requestingUserID) {
		return nil, sharedErrors.ErrUnauthorized
	}

	users, err := s.userRepo.SearchUsers(ctx, query, 20)
	if err != nil {
		return nil, err
	}
	for i := range users {
		s.hydrateUserRoles(&users[i])
	}
	return users, nil
}

// PromoteAdmin promotes a registered Discord user to platform administrator.
func (s *AuthServiceImpl) PromoteAdmin(ctx context.Context, requestingUserID, targetDiscordID string) error {
	if !s.IsSuperAdmin(requestingUserID) {
		return sharedErrors.ErrUnauthorized
	}

	target, err := s.userRepo.GetByID(ctx, targetDiscordID)
	if err != nil {
		return fmt.Errorf("user %s must sign in with Discord at least once before being promoted", targetDiscordID)
	}

	if s.IsSuperAdmin(target.ID) {
		return fmt.Errorf("user %s is already a Super Admin via environment configuration", targetDiscordID)
	}

	return s.userRepo.SetAdmin(ctx, targetDiscordID, true, requestingUserID)
}

// RevokeAdmin removes admin privileges from an appointed administrator.
func (s *AuthServiceImpl) RevokeAdmin(ctx context.Context, requestingUserID, targetDiscordID string) error {
	if !s.IsSuperAdmin(requestingUserID) {
		return sharedErrors.ErrUnauthorized
	}

	if s.IsSuperAdmin(targetDiscordID) {
		return fmt.Errorf("cannot revoke a Super Admin configured via environment variable")
	}

	return s.userRepo.SetAdmin(ctx, targetDiscordID, false, requestingUserID)
}
