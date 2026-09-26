package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/discord-subscriptions/auth-svc/internal/core/domain"
)

type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	logger       *slog.Logger
	isMockMode   bool
}

// NewClient initializes a Discord API adapter.
func NewClient(clientID, clientSecret string, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}

	isMock := clientID == "" || clientID == "your_discord_client_id"
	if isMock {
		logger.Warn("Discord credentials not set or placeholder; running Discord client in simulation mode")
	}

	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		logger:       logger,
		isMockMode:   isMock,
	}
}

// GetOAuthURL constructs the Discord OAuth2 authorization URL.
func (c *Client) GetOAuthURL(redirectURI, state string) string {
	baseURL := "https://discord.com/api/oauth2/authorize"
	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "identify email guilds")
	if state != "" {
		params.Set("state", state)
	}
	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// ExchangeCode trades an authorization code for Discord access and refresh tokens.
func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (string, string, int, error) {
	if c.isMockMode {
		return "mock_discord_access_token", "mock_discord_refresh_token", 604800, nil
	}

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://discord.com/api/v10/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create oauth token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("discord token exchange error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", 0, fmt.Errorf("discord oauth token failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", "", 0, fmt.Errorf("failed to decode discord token response: %w", err)
	}

	return tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn, nil
}

// RefreshToken exchanges a Discord refresh token for a fresh access token and refresh token.
func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (string, string, int, error) {
	if c.isMockMode {
		return "mock_discord_access_token_refreshed", "mock_discord_refresh_token_refreshed", 604800, nil
	}

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://discord.com/api/v10/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to create oauth refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", 0, fmt.Errorf("discord token refresh error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", 0, fmt.Errorf("discord oauth refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", "", 0, fmt.Errorf("failed to decode discord refresh response: %w", err)
	}

	return tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn, nil
}

// GetUserProfile fetches the authenticated user's Discord profile.
func (c *Client) GetUserProfile(ctx context.Context, accessToken string) (*domain.User, error) {
	if c.isMockMode {
		return &domain.User{
			ID:         "123456789012345678",
			Username:   "DevAdmin",
			GlobalName: "Developer Admin",
			Avatar:     "",
			Email:      "dev@example.com",
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/v10/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed fetching discord profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord profile returned status %d", resp.StatusCode)
	}

	var u struct {
		ID         string `json:"id"`
		Username   string `json:"username"`
		GlobalName string `json:"global_name"`
		Avatar     string `json:"avatar"`
		Email      string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return nil, fmt.Errorf("failed decoding discord profile: %w", err)
	}

	return &domain.User{
		ID:         u.ID,
		Username:   u.Username,
		GlobalName: u.GlobalName,
		Avatar:     u.Avatar,
		Email:      u.Email,
	}, nil
}

// GetUserGuilds fetches guilds the user is a member of.
func (c *Client) GetUserGuilds(ctx context.Context, accessToken string) ([]domain.Guild, error) {
	if c.isMockMode {
		return []domain.Guild{
			{
				ID:          "987654321098765432",
				Name:        "Dev Community Server",
				Owner:       true,
				Permissions: "8", // Administrator
				CanManage:   true,
			},
			{
				ID:          "555555555555555555",
				Name:        "Gaming Clan",
				Owner:       false,
				Permissions: "32", // Manage Guild
				CanManage:   true,
			},
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://discord.com/api/v10/users/@me/guilds", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed fetching user guilds: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discord guilds returned status %d", resp.StatusCode)
	}

	var rawGuilds []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Icon        string `json:"icon"`
		Owner       bool   `json:"owner"`
		Permissions string `json:"permissions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawGuilds); err != nil {
		return nil, fmt.Errorf("failed decoding user guilds: %w", err)
	}

	guilds := make([]domain.Guild, 0, len(rawGuilds))
	for _, g := range rawGuilds {
		guild := domain.Guild{
			ID:          g.ID,
			Name:        g.Name,
			Icon:        g.Icon,
			Owner:       g.Owner,
			Permissions: g.Permissions,
		}
		guild.ComputeCanManage()
		guilds = append(guilds, guild)
	}

	return guilds, nil
}
