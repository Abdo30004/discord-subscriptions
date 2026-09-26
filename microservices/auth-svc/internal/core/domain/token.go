package domain

import (
	"time"
)

// UserClaims defines the payload embedded inside the signed JWT.
type UserClaims struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	Avatar     string `json:"avatar"`
	GlobalName string `json:"global_name,omitempty"`
}

// AuthSession represents a successfully authenticated user session.
type AuthSession struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}
