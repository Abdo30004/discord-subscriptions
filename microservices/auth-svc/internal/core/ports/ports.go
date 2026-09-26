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
	ListAdmins(ctx context.Context) ([]domain.User, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]domain.User, error)
	SetAdmin(ctx context.Context, discordID string, isAdmin bool, promotedBy string) error
}

// DiscordClient abstracts communication with Discord OAuth2 and REST APIs.
type DiscordClient interface {
	GetOAuthURL(redirectURI, state string) string
	ExchangeCode(ctx context.Context, code, redirectURI string) (accessToken, refreshToken string, expiresIn int, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, expiresIn int, err error)
	GetUserProfile(ctx context.Context, accessToken string) (*domain.User, error)
	GetUserGuilds(ctx context.Context, accessToken string) ([]domain.Guild, error)
}

// TokenManager signs and verifies JWT tokens.
type TokenManager interface {
	GenerateToken(user domain.User) (token string, expiresAt time.Time, err error)
	ValidateToken(tokenString string) (*domain.UserClaims, error)
	ValidateTokenAllowExpired(tokenString string, maxExpiredAge time.Duration) (*domain.UserClaims, error)
}

// AuthService defines high-level authentication use cases.
type AuthService interface {
	GetOAuthURL(redirectURI string) string
	AuthenticateWithCode(ctx context.Context, code, redirectURI string) (*domain.AuthSession, error)
	RefreshSession(ctx context.Context, currentToken string) (*domain.AuthSession, error)
	GetUserSession(ctx context.Context, token string) (*domain.User, error)
	GetUserManageableGuilds(ctx context.Context, userID string) ([]domain.Guild, error)
	ListAdmins(ctx context.Context, requestingUserID string) ([]domain.User, error)
	SearchUsers(ctx context.Context, requestingUserID, query string) ([]domain.User, error)
	PromoteAdmin(ctx context.Context, requestingUserID, targetDiscordID string) error
	RevokeAdmin(ctx context.Context, requestingUserID, targetDiscordID string) error
	IsSuperAdmin(discordID string) bool
}
