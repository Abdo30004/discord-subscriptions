package messaging

import (
	"sync"
	"time"
)

// EventDeduplicator provides thread-safe in-memory deduplication for event IDs with TTL expiration.
type EventDeduplicator struct {
	mu      sync.RWMutex
	seen    map[string]time.Time
	ttl     time.Duration
	stopCh  chan struct{}
	stopped bool
}

// NewEventDeduplicator creates a deduplicator with the given retention TTL and starts background cleanup.
func NewEventDeduplicator(ttl time.Duration) *EventDeduplicator {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	d := &EventDeduplicator{
		seen:   make(map[string]time.Time),
		ttl:    ttl,
		stopCh: make(chan struct{}),
	}

	go d.cleanupLoop()
	return d
}

// IsDuplicate returns true if the eventID has already been processed within the TTL window.
// If the eventID is new, it records the ID and returns false.
func (d *EventDeduplicator) IsDuplicate(eventID string) bool {
	if eventID == "" {
		return false
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	if exp, exists := d.seen[eventID]; exists {
		if now.Before(exp) {
			return true
		}
	}

	d.seen[eventID] = now.Add(d.ttl)
	return false
}

// Close stops the background cleanup routine.
func (d *EventDeduplicator) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.stopped {
		d.stopped = true
		close(d.stopCh)
	}
}

func (d *EventDeduplicator) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-d.stopCh:
			return
		case <-ticker.C:
			d.cleanup()
		}
	}
}

func (d *EventDeduplicator) cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now().UTC()
	for id, exp := range d.seen {
		if now.After(exp) {
			delete(d.seen, id)
		}
	}
}
