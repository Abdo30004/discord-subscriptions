package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChecker_Healthy(t *testing.T) {
	checker := NewChecker("test-service", "1.0.0")
	checker.AddCheck("custom_ok", func(ctx context.Context) error {
		return nil
	})

	resp, ok := checker.Check(context.Background())
	if !ok {
		t.Fatalf("expected healthy, got unhealthy")
	}
	if resp.Status != "healthy" {
		t.Fatalf("expected status healthy, got %s", resp.Status)
	}
	if resp.Checks["custom_ok"] != "up" {
		t.Fatalf("expected custom_ok up, got %s", resp.Checks["custom_ok"])
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	checker.HealthHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}

func TestChecker_Unhealthy(t *testing.T) {
	checker := NewChecker("test-service", "1.0.0")
	checker.AddCheck("failing_dep", func(ctx context.Context) error {
		return errors.New("connection failed")
	})

	resp, ok := checker.Check(context.Background())
	if ok {
		t.Fatalf("expected unhealthy, got healthy")
	}
	if resp.Status != "unhealthy" {
		t.Fatalf("expected status unhealthy, got %s", resp.Status)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	checker.HealthHandler()(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 Service Unavailable, got %d", rr.Code)
	}

	var jsonResp HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &jsonResp); err != nil {
		t.Fatalf("failed decoding json response: %v", err)
	}
	if jsonResp.Status != "unhealthy" {
		t.Fatalf("expected unhealthy in body, got %s", jsonResp.Status)
	}
}

func TestChecker_Livez(t *testing.T) {
	checker := NewChecker("test-service", "1.0.0")
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rr := httptest.NewRecorder()
	checker.LivezHandler()(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
}
