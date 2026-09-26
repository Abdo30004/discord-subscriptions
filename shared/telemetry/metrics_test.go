package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetrics_HTTPMiddleware(t *testing.T) {
	middleware := HTTPMiddleware("test-service")

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	// Verify metrics endpoint serves Prometheus format
	metricsReq := httptest.NewRequest("GET", "/metrics", nil)
	metricsRec := httptest.NewRecorder()

	Handler().ServeHTTP(metricsRec, metricsReq)

	if metricsRec.Code != http.StatusOK {
		t.Fatalf("expected /metrics status %d, got %d", http.StatusOK, metricsRec.Code)
	}

	body := metricsRec.Body.String()
	if body == "" {
		t.Fatalf("expected non-empty metrics output")
	}
}

func TestAMQPTraceContext(t *testing.T) {
	ctx := context.Background()
	headers := make(map[string]interface{})

	InjectAMQPTraceContext(ctx, headers)

	extractedCtx := ExtractAMQPTraceContext(ctx, headers)
	if extractedCtx == nil {
		t.Fatalf("expected non-nil extracted context")
	}
}
