package vault

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrSecretNotFound = errors.New("secret not found in vault")
)

// Client defines methods for interacting with HashiCorp Vault KV v2 engine.
type Client interface {
	PutBotToken(ctx context.Context, botID, token string) error
	GetBotToken(ctx context.Context, botID string) (string, error)
	PutUserTokens(ctx context.Context, userID, accessToken, refreshToken string, expiresAt time.Time) error
	GetUserTokens(ctx context.Context, userID string) (accessToken, refreshToken string, expiresAt time.Time, err error)
}

// VaultClient provides HTTP access to Vault KV v2 secrets engine.
type VaultClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewVaultClient creates a configured Vault KV v2 client.
func NewVaultClient(addr, token string) *VaultClient {
	return &VaultClient{
		baseURL: strings.TrimRight(addr, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Ping checks if Vault is responsive via the /v1/sys/health endpoint.
func (v *VaultClient) Ping(ctx context.Context) error {
	if v == nil {
		return fmt.Errorf("vault client is nil")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/sys/health", v.baseURL), nil)
	if err != nil {
		return err
	}
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("vault health probe failed: %w", err)
	}
	defer resp.Body.Close()

	// Status 200 (initialized & unsealed) or 429 (unsealed standby) are operational
	if resp.StatusCode >= 500 {
		return fmt.Errorf("vault returned error status: %d", resp.StatusCode)
	}
	return nil
}

type secretWritePayload struct {
	Data map[string]string `json:"data"`
}

type secretReadResponse struct {
	Data struct {
		Data map[string]string `json:"data"`
	} `json:"data"`
}

// PutBotToken securely stores a bot's Discord token at secret/data/bots/<botID>.
func (v *VaultClient) PutBotToken(ctx context.Context, botID, token string) error {
	url := fmt.Sprintf("%s/v1/secret/data/bots/%s", v.baseURL, botID)

	payload := secretWritePayload{
		Data: map[string]string{
			"token": token,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal vault payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create vault request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault responded with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetBotToken retrieves a bot's Discord token from secret/data/bots/<botID>.
func (v *VaultClient) GetBotToken(ctx context.Context, botID string) (string, error) {
	url := fmt.Sprintf("%s/v1/secret/data/bots/%s", v.baseURL, botID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create vault request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", ErrSecretNotFound
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vault responded with status %d: %s", resp.StatusCode, string(respBody))
	}

	var vaultResp secretReadResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return "", fmt.Errorf("failed to decode vault response: %w", err)
	}

	token, ok := vaultResp.Data.Data["token"]
	if !ok || token == "" {
		return "", errors.New("bot token missing in secret data")
	}

	return token, nil
}

// PutUserTokens securely stores a user's Discord OAuth access and refresh tokens at secret/data/users/<userID>.
func (v *VaultClient) PutUserTokens(ctx context.Context, userID, accessToken, refreshToken string, expiresAt time.Time) error {
	url := fmt.Sprintf("%s/v1/secret/data/users/%s", v.baseURL, userID)

	payload := secretWritePayload{
		Data: map[string]string{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt.Format(time.RFC3339),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal vault payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create vault request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("vault responded with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetUserTokens retrieves a user's Discord OAuth tokens from secret/data/users/<userID>.
func (v *VaultClient) GetUserTokens(ctx context.Context, userID string) (string, string, time.Time, error) {
	url := fmt.Sprintf("%s/v1/secret/data/users/%s", v.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to create vault request: %w", err)
	}

	req.Header.Set("X-Vault-Token", v.token)

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to execute vault request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", "", time.Time{}, ErrSecretNotFound
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", "", time.Time{}, fmt.Errorf("vault responded with status %d: %s", resp.StatusCode, string(respBody))
	}

	var vaultResp secretReadResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return "", "", time.Time{}, fmt.Errorf("failed to decode vault response: %w", err)
	}

	accessToken := vaultResp.Data.Data["access_token"]
	refreshToken := vaultResp.Data.Data["refresh_token"]
	var expiresAt time.Time
	if expStr, ok := vaultResp.Data.Data["expires_at"]; ok && expStr != "" {
		expiresAt, _ = time.Parse(time.RFC3339, expStr)
	}

	if accessToken == "" {
		return "", "", time.Time{}, errors.New("access token missing in user secret data")
	}

	return accessToken, refreshToken, expiresAt, nil
}
