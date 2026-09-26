package messaging

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestRabbitMQClient_InitialConnectFailureAndResilience(t *testing.T) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Attempt connecting to an unreachable port to verify resilient startup
	client, err := NewRabbitMQClient("amqp://guest:guest@127.0.0.1:59999/", discardLogger)
	if err != nil {
		t.Fatalf("expected NewRabbitMQClient to not return fatal error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}

	// Should not be connected
	if client.IsConnected() {
		t.Error("expected client to not be connected to invalid address")
	}

	// Subscribing while disconnected should register subscription without crashing or returning error
	called := false
	handler := func(ctx context.Context, body []byte) error {
		called = true
		return nil
	}

	err = client.Subscribe(context.Background(), "test-queue", "test-exchange", "test-key", handler)
	if err != nil {
		t.Errorf("expected Subscribe while disconnected to return nil, got: %v", err)
	}

	// Verify subscription was saved in client
	client.mu.RLock()
	subCount := len(client.subscriptions)
	client.mu.RUnlock()
	if subCount != 1 {
		t.Errorf("expected 1 queued subscription, got %d", subCount)
	}

	// Publishing while disconnected should return ErrNotConnected
	err = client.Publish(context.Background(), "test-exchange", "test-key", nil)
	if err == nil {
		t.Error("expected error publishing while disconnected, got nil")
	}

	// Cleanup
	doneCh := make(chan struct{})
	go func() {
		_ = client.Close()
		close(doneCh)
	}()

	select {
	case <-doneCh:
		// Clean exit
	case <-time.After(2 * time.Second):
		t.Fatal("client.Close() timed out")
	}

	if !called {
		// Handler shouldn't have been called
	}
}
