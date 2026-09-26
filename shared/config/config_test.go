package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/discord-subscriptions/shared/config"
)

func TestConfigFallbacks(t *testing.T) {
	if val := config.GetString("NON_EXISTENT_VAR_123", "default_val"); val != "default_val" {
		t.Fatalf("expected default_val, got %s", val)
	}

	if val := config.GetInt("NON_EXISTENT_INT_123", 42); val != 42 {
		t.Fatalf("expected 42, got %d", val)
	}

	if val := config.GetBool("NON_EXISTENT_BOOL_123", true); !val {
		t.Fatal("expected true")
	}

	if val := config.GetDuration("NON_EXISTENT_DUR_123", 5*time.Second); val != 5*time.Second {
		t.Fatalf("expected 5s, got %v", val)
	}
}

func TestConfigOverrides(t *testing.T) {
	_ = os.Setenv("TEST_KEY_STR", "custom_val")
	_ = os.Setenv("TEST_KEY_INT", "100")
	_ = os.Setenv("TEST_KEY_BOOL", "true")
	defer func() {
		_ = os.Unsetenv("TEST_KEY_STR")
		_ = os.Unsetenv("TEST_KEY_INT")
		_ = os.Unsetenv("TEST_KEY_BOOL")
	}()

	if val := config.GetString("TEST_KEY_STR", "fallback"); val != "custom_val" {
		t.Fatalf("expected custom_val, got %s", val)
	}
	if val := config.GetInt("TEST_KEY_INT", 1); val != 100 {
		t.Fatalf("expected 100, got %d", val)
	}
	if val := config.GetBool("TEST_KEY_BOOL", false); !val {
		t.Fatal("expected true")
	}
}
