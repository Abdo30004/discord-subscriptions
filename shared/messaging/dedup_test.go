package messaging

import (
	"testing"
	"time"
)

func TestEventDeduplicator(t *testing.T) {
	t.Run("detects duplicate and allows new IDs", func(t *testing.T) {
		dedup := NewEventDeduplicator(1 * time.Hour)
		defer dedup.Close()

		id1 := "evt-12345"
		id2 := "evt-67890"

		// First occurrence of id1
		if dedup.IsDuplicate(id1) {
			t.Errorf("expected first occurrence of %s to not be duplicate", id1)
		}

		// Second occurrence of id1
		if !dedup.IsDuplicate(id1) {
			t.Errorf("expected second occurrence of %s to be duplicate", id1)
		}

		// First occurrence of id2
		if dedup.IsDuplicate(id2) {
			t.Errorf("expected first occurrence of %s to not be duplicate", id2)
		}
	})

	t.Run("expires entries after TTL", func(t *testing.T) {
		dedup := NewEventDeduplicator(50 * time.Millisecond)
		defer dedup.Close()

		id := "evt-expiring"
		if dedup.IsDuplicate(id) {
			t.Fatalf("first check should not be duplicate")
		}
		if !dedup.IsDuplicate(id) {
			t.Fatalf("immediate check should be duplicate")
		}

		// Wait for TTL to pass
		time.Sleep(70 * time.Millisecond)

		// After expiry, should be treated as new
		if dedup.IsDuplicate(id) {
			t.Errorf("expected expired event ID to be accepted again")
		}
	})

	t.Run("ignores empty ID", func(t *testing.T) {
		dedup := NewEventDeduplicator(1 * time.Hour)
		defer dedup.Close()

		if dedup.IsDuplicate("") {
			t.Errorf("expected empty string to not be flagged duplicate")
		}
	})
}
