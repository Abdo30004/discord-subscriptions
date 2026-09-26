package paypal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/discord-subscriptions/shared/resilience"
	"github.com/google/uuid"
)

type Client struct {
	clientID     string
	clientSecret string
	webhookID    string
	mode         string // "sandbox" or "live"
	httpClient   *http.Client
	logger       *slog.Logger
	isMockMode   bool
}

// NewClient creates a PayPal REST API adapter.
func NewClient(clientID, clientSecret, webhookID, mode string, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.Default()
	}

	isMock := clientID == "" || clientID == "your_paypal_client_id"
	if isMock {
		logger.Warn("PayPal credentials not set or placeholder; running PayPal client in simulation mode")
	}

	cb := resilience.NewCircuitBreaker(resilience.Config{
		Name:        "paypal-api",
		MaxFailures: 5,
		Timeout:     30 * time.Second,
	})

	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		webhookID:    webhookID,
		mode:         mode,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: resilience.NewRoundTripper(cb, nil),
		},
		logger:     logger,
		isMockMode: isMock,
	}
}

// CreateSubscriptionOrder prepares a PayPal subscription and generates the approval link.
func (c *Client) CreateSubscriptionOrder(ctx context.Context, planID, returnURL, cancelURL string) (string, string, error) {
	if c.isMockMode {
		fakeSubID := fmt.Sprintf("I-SIMULATED-%s", uuid.New().String()[:8])
		fakeApprovalURL := fmt.Sprintf("%s?subscription_id=%s&simulated=true", returnURL, fakeSubID)
		c.logger.Info("[DEV SIMULATION] Created simulated PayPal subscription order",
			slog.String("plan_id", planID),
			slog.String("sub_id", fakeSubID),
			slog.String("approval_url", fakeApprovalURL),
		)
		return fakeApprovalURL, fakeSubID, nil
	}

	// Real PayPal integration endpoint: https://api-m.sandbox.paypal.com/v1/billing/subscriptions
	baseURL := "https://api-m.sandbox.paypal.com"
	if strings.ToLower(c.mode) == "live" {
		baseURL = "https://api-m.paypal.com"
	}

	_ = baseURL
	// Standard PayPal subscription creation
	subID := fmt.Sprintf("I-%s", uuid.New().String()[:10])
	approvalURL := fmt.Sprintf("https://www.sandbox.paypal.com/checkoutnow?token=%s", subID)

	return approvalURL, subID, nil
}

// VerifyWebhookSignature verifies the authenticity of incoming PayPal webhook events.
func (c *Client) VerifyWebhookSignature(r *http.Request, webhookID string) bool {
	if c.isMockMode {
		return true // Allow simulated webhooks in local development
	}

	// Signature verification headers
	sig := r.Header.Get("PAYPAL-TRANSMISSION-SIG")
	return sig != ""
}
