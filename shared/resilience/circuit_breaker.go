package resilience

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// State represents the current operational status of the CircuitBreaker.
type State string

const (
	StateClosed   State = "CLOSED"
	StateHalfOpen State = "HALF_OPEN"
	StateOpen     State = "OPEN"
)

var (
	ErrCircuitOpen = errors.New("circuit breaker is open: request denied to prevent cascade failure")
)

// Config specifies operating parameters for the CircuitBreaker.
type Config struct {
	Name            string
	MaxFailures     int           // Consecutive failures required to open the circuit (default: 5)
	Timeout         time.Duration // Time to wait in OPEN state before trying HALF_OPEN (default: 30s)
	HalfOpenTrials  int           // Consecutive successful trial requests needed to close circuit (default: 2)
}

// CircuitBreaker guards against cascading failures when communicating with external dependencies.
type CircuitBreaker struct {
	mu             sync.RWMutex
	name           string
	maxFailures    int
	timeout        time.Duration
	halfOpenTrials int

	state          State
	failures       int
	halfOpenSuccess int
	openedAt       time.Time
}

// NewCircuitBreaker creates a circuit breaker with sensible defaults.
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	if cfg.Name == "" {
		cfg.Name = "default"
	}
	if cfg.MaxFailures <= 0 {
		cfg.MaxFailures = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.HalfOpenTrials <= 0 {
		cfg.HalfOpenTrials = 2
	}

	return &CircuitBreaker{
		name:           cfg.Name,
		maxFailures:    cfg.MaxFailures,
		timeout:        cfg.Timeout,
		halfOpenTrials: cfg.HalfOpenTrials,
		state:          StateClosed,
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.checkTransition()
	return cb.state
}

// checkTransition updates OPEN -> HALF_OPEN if timeout has elapsed. Must be called under lock.
func (cb *CircuitBreaker) checkTransition() {
	if cb.state == StateOpen && time.Since(cb.openedAt) >= cb.timeout {
		cb.state = StateHalfOpen
		cb.halfOpenSuccess = 0
	}
}

// Execute runs the operation protected by the circuit breaker.
func (cb *CircuitBreaker) Execute(ctx context.Context, op func() error) error {
	cb.mu.Lock()
	cb.checkTransition()

	switch cb.state {
	case StateOpen:
		cb.mu.Unlock()
		return fmt.Errorf("%w (%s)", ErrCircuitOpen, cb.name)

	case StateHalfOpen:
		cb.mu.Unlock()
		err := op()
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if err != nil {
			// Failed trial, reopen immediately
			cb.state = StateOpen
			cb.openedAt = time.Now()
			cb.failures++
			return err
		}
		cb.halfOpenSuccess++
		if cb.halfOpenSuccess >= cb.halfOpenTrials {
			// Restored healthy state
			cb.state = StateClosed
			cb.failures = 0
			cb.halfOpenSuccess = 0
		}
		return nil

	default: // StateClosed
		cb.mu.Unlock()
		err := op()
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if err != nil {
			cb.failures++
			if cb.failures >= cb.maxFailures {
				cb.state = StateOpen
				cb.openedAt = time.Now()
			}
			return err
		}
		cb.failures = 0
		return nil
	}
}

// ExecuteWithFallback runs the operation and invokes the fallback if the circuit is open.
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, op func() error, fallback func(err error) error) error {
	err := cb.Execute(ctx, op)
	if err != nil && errors.Is(err, ErrCircuitOpen) && fallback != nil {
		return fallback(err)
	}
	return err
}

// Reset resets the circuit breaker to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.halfOpenSuccess = 0
}

// RoundTripper wraps an http.RoundTripper with circuit breaker protection.
type RoundTripper struct {
	cb        *CircuitBreaker
	transport http.RoundTripper
}

// NewRoundTripper returns an http.RoundTripper protected by the given circuit breaker.
func NewRoundTripper(cb *CircuitBreaker, transport http.RoundTripper) *RoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &RoundTripper{
		cb:        cb,
		transport: transport,
	}
}

// RoundTrip executes the HTTP request under the circuit breaker.
func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	err := rt.cb.Execute(req.Context(), func() error {
		var tripErr error
		resp, tripErr = rt.transport.RoundTrip(req)
		if tripErr != nil {
			return tripErr
		}
		// Consider 5xx server errors as failure signals for upstream dependency
		if resp.StatusCode >= 500 {
			return fmt.Errorf("upstream dependency returned server error: %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}
