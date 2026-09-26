package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
	"github.com/discord-subscriptions/auth-svc/internal/core/ports"
	"github.com/google/uuid"
)

type AuthServiceImpl struct {
	userRepo   ports.UserRepository
	discord    ports.DiscordClient
	tokenMgr   ports.TokenManager
	logger     *slog.Logger
}

func NewAuthService(
	userRepo ports.UserRepository,
	discord ports.DiscordClient,
	tokenMgr ports.TokenManager,
	logger *slog.Logger,
) *AuthServiceImpl {
	if logger == nil {
		logger = slog.Default()
	}
	return &AuthServiceImpl{
		userRepo: userRepo,
		discord:  discord,
		tokenMgr: tokenMgr,
		logger:   logger,
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

	now := time.Now().UTC()
	discordUser.AccessToken = accessToken
	discordUser.RefreshToken = refreshToken
	discordUser.TokenExpiresAt = now.Add(time.Duration(expiresIn) * time.Second)
	discordUser.UpdatedAt = now
	discordUser.CreatedAt = now

	// 3. Upsert into database
	if err := s.userRepo.Upsert(ctx, discordUser); err != nil {
		return nil, fmt.Errorf("failed saving user profile: %w", err)
	}

	// 4. Generate JWT
	jwtToken, expiresAt, err := s.tokenMgr.GenerateToken(*discordUser)
	if err != nil {
		return nil, fmt.Errorf("failed generating auth token: %w", err)
	}

	s.logger.Info("user logged in successfully",
		slog.String("user_id", discordUser.ID),
		slog.String("username", discordUser.Username),
	)

	return &domain.AuthSession{
		Token:     jwtToken,
		ExpiresAt: expiresAt,
		User:      *discordUser,
	}, nil
}

// GetUserSession validates a JWT and retrieves the user record.
func (s *AuthServiceImpl) GetUserSession(ctx context.Context, token string) (*domain.User, error) {
	claims, err := s.tokenMgr.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	return s.userRepo.GetByID(ctx, claims.UserID)
}

// GetUserManageableGuilds retrieves only the guilds where the user has bot-installation permissions.
func (s *AuthServiceImpl) GetUserManageableGuilds(ctx context.Context, userID string) ([]domain.Guild, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	allGuilds, err := s.discord.GetUserGuilds(ctx, user.AccessToken)
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

// DevLogin generates an instant mock session for local development without Discord credentials.
func (s *AuthServiceImpl) DevLogin(ctx context.Context, mockUserID, mockUsername string) (*domain.AuthSession, error) {
	if mockUserID == "" {
		mockUserID = "123456789012345678"
	}
	if mockUsername == "" {
		mockUsername = "DevAdmin"
	}

	now := time.Now().UTC()
	user := domain.User{
		ID:         mockUserID,
		Username:   mockUsername,
		GlobalName: "Developer Administrator",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	_ = s.userRepo.Upsert(ctx, &user)

	token, expiresAt, err := s.tokenMgr.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthSession{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}
