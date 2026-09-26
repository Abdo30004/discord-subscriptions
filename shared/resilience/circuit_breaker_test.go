package resilience

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:            "test-cb",
		MaxFailures:     3,
		Timeout:         50 * time.Millisecond,
		HalfOpenTrials:  2,
	})

	ctx := context.Background()

	// Initial state is CLOSED
	if cb.State() != StateClosed {
		t.Fatalf("expected state %s, got %s", StateClosed, cb.State())
	}

	testErr := errors.New("simulated network failure")

	// Trigger 2 failures (below threshold 3)
	for i := 0; i < 2; i++ {
		_ = cb.Execute(ctx, func() error { return testErr })
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected state %s, got %s", StateClosed, cb.State())
	}

	// 3rd failure trips the breaker to OPEN
	_ = cb.Execute(ctx, func() error { return testErr })
	if cb.State() != StateOpen {
		t.Fatalf("expected state %s, got %s", StateOpen, cb.State())
	}

	// Call while OPEN returns ErrCircuitOpen immediately without running op
	ran := false
	err := cb.Execute(ctx, func() error {
		ran = true
		return nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if ran {
		t.Fatalf("expected operation not to execute while circuit is OPEN")
	}

	// Wait for Timeout to transition to HALF_OPEN
	time.Sleep(60 * time.Millisecond)
	if cb.State() != StateHalfOpen {
		t.Fatalf("expected state %s, got %s", StateHalfOpen, cb.State())
	}

	// Trial 1: success -> still HALF_OPEN
	err = cb.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cb.State() != StateHalfOpen {
		t.Fatalf("expected state %s, got %s", StateHalfOpen, cb.State())
	}

	// Trial 2: success -> reaches HalfOpenTrials (2) -> recovers to CLOSED
	err = cb.Execute(ctx, func() error { return nil })
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cb.State() != StateClosed {
		t.Fatalf("expected state %s, got %s", StateClosed, cb.State())
	}
}

func TestCircuitBreaker_RoundTripper(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		Name:        "http-cb",
		MaxFailures: 2,
		Timeout:     50 * time.Millisecond,
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: NewRoundTripper(cb, nil),
	}

	// Trigger 2 500 errors to trip circuit
	for i := 0; i < 2; i++ {
		_, _ = client.Get(server.URL)
	}

	if cb.State() != StateOpen {
		t.Fatalf("expected circuit OPEN, got %s", cb.State())
	}

	// Next call fails fast with ErrCircuitOpen
	_, err := client.Get(server.URL)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}
