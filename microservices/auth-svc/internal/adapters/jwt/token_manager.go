package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired authorization token")
)

type JWTClaims struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar"`
	GlobalName string `json:"global_name,omitempty"`
	jwt.RegisteredClaims
}

type Manager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

// NewManager creates a JWT signer and validator.
func NewManager(secretKey string, duration time.Duration) *Manager {
	if duration <= 0 {
		duration = 7 * 24 * time.Hour // Default 7 days session
	}
	return &Manager{
		secretKey:     []byte(secretKey),
		tokenDuration: duration,
	}
}

// GenerateToken creates a signed JWT for the given user.
func (m *Manager) GenerateToken(user domain.User) (string, time.Time, error) {
	expiresAt := time.Now().UTC().Add(m.tokenDuration)

	claims := JWTClaims{
		UserID:     user.ID,
		Username:   user.Username,
		Avatar:     user.Avatar,
		GlobalName: user.GlobalName,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			Issuer:    "discord-subscriptions-auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign jwt: %w", err)
	}

	return signedToken, expiresAt, nil
}

// ValidateToken parses and verifies a JWT string, returning user claims.
func (m *Manager) ValidateToken(tokenString string) (*domain.UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &domain.UserClaims{
		UserID:     claims.UserID,
		Username:   claims.Username,
		Avatar:     claims.Avatar,
		GlobalName: claims.GlobalName,
	}, nil
}
