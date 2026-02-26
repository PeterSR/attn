package cmd

import (
	"testing"

	"github.com/petersr/attn/internal/config"
)

func TestApplyWhenOverrides(t *testing.T) {
	t.Run("valid overrides", func(t *testing.T) {
		cfg := config.Default()
		err := applyWhenOverrides(&cfg, map[string]string{
			"bell":    "always",
			"desktop": "never",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Bell.When != config.WhenAlways {
			t.Errorf("bell.when = %q, want %q", cfg.Bell.When, config.WhenAlways)
		}
		if cfg.Desktop.When != config.WhenNever {
			t.Errorf("desktop.when = %q, want %q", cfg.Desktop.When, config.WhenNever)
		}
	})

	t.Run("all channels", func(t *testing.T) {
		cfg := config.Default()
		err := applyWhenOverrides(&cfg, map[string]string{
			"desktop":  "idle",
			"bell":     "active",
			"ntfy":     "always",
			"pushover": "never",
			"webhook":  "idle",
			"relay":    "always",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Desktop.When != config.WhenIdle {
			t.Errorf("desktop.when = %q, want %q", cfg.Desktop.When, config.WhenIdle)
		}
		if cfg.Bell.When != config.WhenActive {
			t.Errorf("bell.when = %q, want %q", cfg.Bell.When, config.WhenActive)
		}
		if cfg.Ntfy.When != config.WhenAlways {
			t.Errorf("ntfy.when = %q, want %q", cfg.Ntfy.When, config.WhenAlways)
		}
		if cfg.Pushover.When != config.WhenNever {
			t.Errorf("pushover.when = %q, want %q", cfg.Pushover.When, config.WhenNever)
		}
		if cfg.Webhook.When != config.WhenIdle {
			t.Errorf("webhook.when = %q, want %q", cfg.Webhook.When, config.WhenIdle)
		}
		if cfg.Relay.When != config.WhenAlways {
			t.Errorf("relay.when = %q, want %q", cfg.Relay.When, config.WhenAlways)
		}
	})

	t.Run("unknown channel", func(t *testing.T) {
		cfg := config.Default()
		err := applyWhenOverrides(&cfg, map[string]string{"bogus": "always"})
		if err == nil {
			t.Fatal("expected error for unknown channel")
		}
	})

	t.Run("invalid when value", func(t *testing.T) {
		cfg := config.Default()
		err := applyWhenOverrides(&cfg, map[string]string{"desktop": "bogus"})
		if err == nil {
			t.Fatal("expected error for invalid when value")
		}
	})

	t.Run("nil map", func(t *testing.T) {
		cfg := config.Default()
		err := applyWhenOverrides(&cfg, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
