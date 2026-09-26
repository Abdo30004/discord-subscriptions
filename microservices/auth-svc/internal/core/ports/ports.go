package ports

import (
	"context"
	"time"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
)

// UserRepository handles user entity persistence in PostgreSQL.
type UserRepository interface {
	Upsert(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

// DiscordClient abstracts communication with Discord OAuth2 and REST APIs.
type DiscordClient interface {
	GetOAuthURL(redirectURI, state string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (accessToken, refreshToken string, expiresIn int, err error)
	GetUserProfile(ctx context.Context, accessToken string) (*domain.User, error)
	GetUserGuilds(ctx context.Context, accessToken string) ([]domain.Guild, error)
}

// TokenManager signs and verifies JWT tokens.
type TokenManager interface {
	GenerateToken(user domain.User) (token string, expiresAt time.Time, err error)
	ValidateToken(tokenString string) (*domain.UserClaims, error)
}

// AuthService defines high-level authentication use cases.
type AuthService interface {
	GetOAuthURL(redirectURI string) string
	AuthenticateWithCode(ctx context.Context, code, redirectURI string) (*domain.AuthSession, error)
	GetUserSession(ctx context.Context, token string) (*domain.User, error)
	GetUserManageableGuilds(ctx context.Context, userID string) ([]domain.Guild, error)
	DevLogin(ctx context.Context, mockUserID, mockUsername string) (*domain.AuthSession, error)
}
