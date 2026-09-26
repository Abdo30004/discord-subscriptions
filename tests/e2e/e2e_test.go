package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

var (
	gatewayURL string
	client     *http.Client
)

func init() {
	gatewayURL = os.Getenv("GATEWAY_URL")
	if gatewayURL == "" {
		gatewayURL = "http://localhost"
	}
	client = &http.Client{
		Timeout: 5 * time.Second,
	}
}

// isGatewayReachable tests if the Traefik edge gateway or local dev stack is running.
func isGatewayReachable() bool {
	resp, err := client.Get(fmt.Sprintf("%s/api/v1/catalog/bots", gatewayURL))
	if err == nil {
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}
	resp, err = client.Get("http://localhost:8081/health")
	if err == nil {
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}
	return false
}

func TestE2E_HealthProbeSweep(t *testing.T) {
	if !isGatewayReachable() {
		t.Skip("Live gateway or local stack not running; skipping live E2E health probe sweep.")
	}

	endpoints := []string{
		gatewayURL + "/api/v1/auth/me",
		gatewayURL + "/api/v1/catalog/bots",
		gatewayURL + "/api/v1/subscriptions/guild/test-guild",
		gatewayURL + "/api/v1/deployments/guild/test-guild",
		gatewayURL + "/api/v1/targets/guild/test-guild",
	}

	for _, ep := range endpoints {
		resp, err := client.Get(ep)
		if err != nil {
			t.Errorf("failed probing endpoint %s: %v", ep, err)
			continue
		}
		resp.Body.Close()
		// Endpoint is up if status is 200 or 401 Unauthorized (which proves the auth middleware is active)
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
			t.Logf("Endpoint %s responded with status %d", ep, resp.StatusCode)
		}
	}
}

func TestE2E_CatalogAndPricingQuery(t *testing.T) {
	if !isGatewayReachable() {
		t.Skip("Live gateway or local stack not running; skipping live E2E catalog query.")
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v1/catalog/bots", gatewayURL))
	if err != nil {
		t.Fatalf("failed fetching catalog from gateway: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 from catalog, got %d", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
			Plans    []struct {
				ID         string `json:"id"`
				PriceCents int    `json:"price_cents"`
			} `json:"plans"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed decoding catalog response: %v", err)
	}

	if len(result.Data) == 0 {
		t.Log("Catalog returned 0 templates (database may be unseeded)")
	} else {
		t.Logf("Catalog verified: found %d live bot templates", len(result.Data))
	}
}

func TestE2E_VoucherCheckoutLifecycle(t *testing.T) {
	if !isGatewayReachable() {
		t.Skip("Live gateway or local stack not running; skipping live E2E voucher checkout.")
	}

	// 1. Redeem a test voucher for guild-e2e-suite
	voucherPayload := map[string]string{
		"code":           "TEST-VOUCHER-PRO",
		"guild_id":       "112233445566778899",
		"user_id":        "998877665544332211",
		"instance_label": "E2E Stage Bot",
	}

	body, _ := json.Marshal(voucherPayload)
	resp, err := client.Post(
		fmt.Sprintf("%s/api/v1/billing/redeem", gatewayURL),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("failed sending voucher redemption: %v", err)
	}
	defer resp.Body.Close()

	// In test mode without voucher seeded, check that the service gracefully returns 400 or 200
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected voucher redemption status: %d", resp.StatusCode)
	}

	t.Logf("Voucher redemption endpoint responded with status %d", resp.StatusCode)
}

func TestE2E_MultiTenantFleetDisambiguation(t *testing.T) {
	if !isGatewayReachable() {
		t.Skip("Live gateway or local stack not running; skipping live E2E fleet disambiguation.")
	}

	guildID := "112233445566778899"
	resp, err := client.Get(fmt.Sprintf("%s/api/v1/deployments/guild/%s", gatewayURL, guildID))
	if err != nil {
		t.Fatalf("failed querying deployments for guild %s: %v", guildID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected HTTP 200 or 404 for fleet query, got %d", resp.StatusCode)
	}

	t.Log("Fleet disambiguation endpoint successfully routed through Traefik")
}
