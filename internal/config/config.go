package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// When controls when a channel fires relative to screen state.
type When string

const (
	WhenNever  When = "never"  // Channel is disabled.
	WhenActive When = "active" // Fire only when screen is on, unlocked, and not focused.
	WhenIdle   When = "idle"   // Fire only when screen is off or locked.
	WhenAlways When = "always" // Fire unconditionally.
)

// Valid returns true if the When value is recognized.
func (w When) Valid() bool {
	switch w {
	case WhenNever, WhenActive, WhenIdle, WhenAlways, "":
		return true
	}
	return false
}

type Config struct {
	Context  ContextConfig  `toml:"context"`
	Desktop  DesktopConfig  `toml:"desktop"`
	Bell     BellConfig     `toml:"bell"`
	Ntfy     NtfyConfig     `toml:"ntfy"`
	Pushover PushoverConfig `toml:"pushover"`
	Webhook  WebhookConfig  `toml:"webhook"`
	Relay    RelayConfig    `toml:"relay"`
	Serve    ServeConfig    `toml:"serve"`
}

type ContextConfig struct {
	Mode string `toml:"mode"` // "auto", "none", or a fixed string
}

type DesktopConfig struct {
	Enabled *bool `toml:"enabled"` // Deprecated: use When instead.
	When    When  `toml:"when"`
}

type BellConfig struct {
	Enabled *bool `toml:"enabled"` // Deprecated: use When instead.
	When    When  `toml:"when"`
}

type NtfyConfig struct {
	Enabled bool   `toml:"enabled"` // Deprecated: use When instead.
	When    When   `toml:"when"`
	Server  string `toml:"server"`
	Topic   string `toml:"topic"`
	Token   string `toml:"token"`
}

type PushoverConfig struct {
	Enabled bool   `toml:"enabled"` // Deprecated: use When instead.
	When    When   `toml:"when"`
	Token   string `toml:"token"`
	UserKey string `toml:"user_key"`
}

type WebhookConfig struct {
	Enabled bool              `toml:"enabled"` // Deprecated: use When instead.
	When    When              `toml:"when"`
	URL     string            `toml:"url"`
	Method  string            `toml:"method"`
	Headers map[string]string `toml:"headers"`
}

type RelayConfig struct {
	When       When   `toml:"when"`
	SocketPath string `toml:"socket_path"`
}

type ServeConfig struct {
	SocketPath string         `toml:"socket_path"`
	Tunnels    []TunnelConfig `toml:"tunnels"`
}

type TunnelConfig struct {
	Name             string `toml:"name"`
	Host             string `toml:"host"`
	User             string `toml:"user"`
	RemoteSocketPath string `toml:"remote_socket_path"`
	IdentityFile     string `toml:"identity_file"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	return Config{
		Context: ContextConfig{Mode: "auto"},
		Desktop: DesktopConfig{When: WhenActive},
		Bell:    BellConfig{When: WhenNever},
		Ntfy:    NtfyConfig{Server: "https://ntfy.sh"},
		Webhook: WebhookConfig{Method: "POST"},
	}
}

// DefaultPath returns the default config file path.
func DefaultPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "attn", "config.toml")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "attn", "config.toml")
}

// Load reads the config from the given path, falling back to defaults.
// If the file does not exist, defaults are returned without error.
func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return Default(), err
	}

	// Start with zero config so we can detect which fields the TOML sets.
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	// Migrate deprecated "enabled" fields to "when".
	migrateDesktop(&cfg)
	migrateBell(&cfg)
	migrateNtfy(&cfg)
	migratePushover(&cfg)
	migrateWebhook(&cfg)

	// Apply non-channel defaults.
	if cfg.Context.Mode == "" {
		cfg.Context.Mode = "auto"
	}
	if cfg.Ntfy.Server == "" {
		cfg.Ntfy.Server = "https://ntfy.sh"
	}
	if cfg.Webhook.Method == "" {
		cfg.Webhook.Method = "POST"
	}

	// Validate When values.
	for name, w := range map[string]When{
		"desktop":  cfg.Desktop.When,
		"bell":     cfg.Bell.When,
		"ntfy":     cfg.Ntfy.When,
		"pushover": cfg.Pushover.When,
		"webhook":  cfg.Webhook.When,
		"relay":    cfg.Relay.When,
	} {
		if !w.Valid() {
			return cfg, fmt.Errorf("%s.when: invalid value %q", name, w)
		}
	}

	return cfg, nil
}

// migrateDesktop resolves When from deprecated Enabled if When is unset.
// Default: active.
func migrateDesktop(cfg *Config) {
	if cfg.Desktop.When != "" {
		return
	}
	if cfg.Desktop.Enabled != nil {
		if *cfg.Desktop.Enabled {
			cfg.Desktop.When = WhenActive
		} else {
			cfg.Desktop.When = WhenNever
		}
		return
	}
	cfg.Desktop.When = WhenActive
}

// migrateBell resolves When from deprecated Enabled if When is unset.
// Default: never.
func migrateBell(cfg *Config) {
	if cfg.Bell.When != "" {
		return
	}
	if cfg.Bell.Enabled != nil {
		if *cfg.Bell.Enabled {
			cfg.Bell.When = WhenAlways
		} else {
			cfg.Bell.When = WhenNever
		}
		return
	}
	cfg.Bell.When = WhenNever
}

// migrateNtfy resolves When from deprecated Enabled if When is unset.
func migrateNtfy(cfg *Config) {
	if cfg.Ntfy.When != "" {
		return
	}
	if cfg.Ntfy.Enabled {
		cfg.Ntfy.When = WhenAlways
	}
}

// migratePushover resolves When from deprecated Enabled if When is unset.
func migratePushover(cfg *Config) {
	if cfg.Pushover.When != "" {
		return
	}
	if cfg.Pushover.Enabled {
		cfg.Pushover.When = WhenAlways
	}
}

// migrateWebhook resolves When from deprecated Enabled if When is unset.
func migrateWebhook(cfg *Config) {
	if cfg.Webhook.When != "" {
		return
	}
	if cfg.Webhook.Enabled {
		cfg.Webhook.When = WhenAlways
	}
}

// DesktopEnabled returns whether the desktop channel is enabled.
func (c Config) DesktopEnabled() bool {
	return c.Desktop.When != WhenNever && c.Desktop.When != ""
}

// BellEnabled returns whether the bell channel is enabled.
func (c Config) BellEnabled() bool {
	return c.Bell.When != WhenNever && c.Bell.When != ""
}
