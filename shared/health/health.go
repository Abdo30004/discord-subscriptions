package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// CheckFunc is a function that inspects the health of a specific dependency.
type CheckFunc func(ctx context.Context) error

// HealthResponse represents the standardized JSON payload for health probes.
type HealthResponse struct {
	Status        string            `json:"status"` // "healthy" or "unhealthy"
	Service       string            `json:"service"`
	Version       string            `json:"version,omitempty"`
	Timestamp     string            `json:"timestamp"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Checks        map[string]string `json:"checks"`
}

// Checker coordinates deep and shallow health checks for a service.
type Checker struct {
	serviceName string
	version     string
	startTime   time.Time
	mu          sync.RWMutex
	checks      map[string]CheckFunc
}

// NewChecker initializes a new health checker instance.
func NewChecker(serviceName, version string) *Checker {
	if version == "" {
		version = "1.0.0"
	}
	return &Checker{
		serviceName: serviceName,
		version:     version,
		startTime:   time.Now().UTC(),
		checks:      make(map[string]CheckFunc),
	}
}

// AddCheck registers an arbitrary dependency check function.
func (c *Checker) AddCheck(name string, check CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// AddDatabaseCheck registers a SQL database ping health check.
func (c *Checker) AddDatabaseCheck(name string, db *sql.DB) {
	c.AddCheck(name, func(ctx context.Context) error {
		if db == nil {
			return fmt.Errorf("database connection is nil")
		}
		return db.PingContext(ctx)
	})
}

// AddRabbitMQCheck registers a RabbitMQ connectivity health check.
func (c *Checker) AddRabbitMQCheck(name string, client interface{ IsConnected() bool }) {
	c.AddCheck(name, func(ctx context.Context) error {
		if client == nil || !client.IsConnected() {
			return fmt.Errorf("rabbitmq is disconnected")
		}
		return nil
	})
}

// AddVaultCheck registers a HashiCorp Vault ping health check.
func (c *Checker) AddVaultCheck(name string, pinger interface{ Ping(context.Context) error }) {
	c.AddCheck(name, func(ctx context.Context) error {
		if pinger == nil {
			return fmt.Errorf("vault client is nil")
		}
		return pinger.Ping(ctx)
	})
}

// Check executes all registered health checks with a timeout and returns the results.
func (c *Checker) Check(ctx context.Context) (HealthResponse, bool) {
	c.mu.RLock()
	checksCopy := make(map[string]CheckFunc, len(c.checks))
	for k, v := range c.checks {
		checksCopy[k] = v
	}
	c.mu.RUnlock()

	results := make(map[string]string)
	allHealthy := true

	for name, checkFn := range checksCopy {
		checkCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := checkFn(checkCtx)
		cancel()

		if err != nil {
			results[name] = fmt.Sprintf("error: %s", err.Error())
			allHealthy = false
		} else {
			results[name] = "up"
		}
	}

	status := "healthy"
	if !allHealthy {
		status = "unhealthy"
	}

	resp := HealthResponse{
		Status:        status,
		Service:       c.serviceName,
		Version:       c.version,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		UptimeSeconds: int64(time.Since(c.startTime).Seconds()),
		Checks:        results,
	}

	return resp, allHealthy
}

// HealthHandler returns an HTTP handler for deep health checks (returns 200 or 503).
func (c *Checker) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, ok := c.Check(r.Context())
		w.Header().Set("Content-Type", "application/json")
		if !ok {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// LivezHandler returns a lightweight HTTP handler for Kubernetes liveness probes.
func (c *Checker) LivezHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "alive",
			"service": c.serviceName,
		})
	}
}

// ReadyzHandler returns an HTTP handler for Kubernetes readiness probes.
func (c *Checker) ReadyzHandler() http.HandlerFunc {
	return c.HealthHandler()
}
