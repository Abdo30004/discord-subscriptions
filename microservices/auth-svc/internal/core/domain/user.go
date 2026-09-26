package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidUserID   = errors.New("invalid or empty user ID")
	ErrInvalidUsername = errors.New("username cannot be empty")
)

// User represents an authenticated Discord user.
type User struct {
	ID             string    `json:"id"` // Discord Snowflake
	Username       string    `json:"username"`
	GlobalName     string    `json:"global_name,omitempty"`
	Avatar         string     `json:"avatar,omitempty"`
	Email          string     `json:"email,omitempty"`
	IsAdmin        bool       `json:"is_admin"`
	IsSuperAdmin   bool       `json:"is_super_admin"`
	AdminPromotedAt *time.Time `json:"admin_promoted_at,omitempty"`
	AdminPromotedBy string     `json:"admin_promoted_by,omitempty"`
	HasOAuthToken   bool       `json:"has_oauth_token"`
	AccessToken    string     `json:"-"`
	RefreshToken   string     `json:"-"`
	TokenExpiresAt time.Time  `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Validate checks user domain invariants.
func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidUserID
	}
	if u.Username == "" {
		return ErrInvalidUsername
	}
	return nil
}

// AvatarURL returns the Discord CDN URL for the user's avatar.
func (u *User) AvatarURL() string {
	if u.Avatar == "" {
		return "https://cdn.discordapp.com/embed/avatars/0.png"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", u.ID, u.Avatar)
}
