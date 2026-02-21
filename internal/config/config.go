package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Context  ContextConfig  `toml:"context"`
	Desktop  DesktopConfig  `toml:"desktop"`
	Bell     BellConfig     `toml:"bell"`
	Ntfy     NtfyConfig     `toml:"ntfy"`
	Pushover PushoverConfig `toml:"pushover"`
	Webhook  WebhookConfig  `toml:"webhook"`
	Serve    ServeConfig    `toml:"serve"`
}

type ContextConfig struct {
	Mode string `toml:"mode"` // "auto", "none", or a fixed string
}

type DesktopConfig struct {
	Enabled *bool `toml:"enabled"`
}

type BellConfig struct {
	Enabled *bool `toml:"enabled"`
}

type NtfyConfig struct {
	Enabled bool   `toml:"enabled"`
	Server  string `toml:"server"`
	Topic   string `toml:"topic"`
	Token   string `toml:"token"`
}

type PushoverConfig struct {
	Enabled bool   `toml:"enabled"`
	Token   string `toml:"token"`
	UserKey string `toml:"user_key"`
}

type WebhookConfig struct {
	Enabled bool              `toml:"enabled"`
	URL     string            `toml:"url"`
	Method  string            `toml:"method"`
	Headers map[string]string `toml:"headers"`
}

type ServeConfig struct {
	SocketPath string         `toml:"socket_path"`
	Tunnels    []TunnelConfig `toml:"tunnels"`
}

type TunnelConfig struct {
	Name         string `toml:"name"`
	Host         string `toml:"host"`
	User         string `toml:"user"`
	RemoteSocket string `toml:"remote_socket"`
	IdentityFile string `toml:"identity_file"`
}

// Default returns a Config with sensible defaults.
func Default() Config {
	t := true
	return Config{
		Context: ContextConfig{Mode: "auto"},
		Desktop: DesktopConfig{Enabled: &t},
		Bell:    BellConfig{Enabled: &t},
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
	cfg := Default()
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	// Apply defaults for fields not set after unmarshal
	if cfg.Desktop.Enabled == nil {
		t := true
		cfg.Desktop.Enabled = &t
	}
	if cfg.Bell.Enabled == nil {
		t := true
		cfg.Bell.Enabled = &t
	}
	if cfg.Ntfy.Server == "" {
		cfg.Ntfy.Server = "https://ntfy.sh"
	}
	if cfg.Webhook.Method == "" {
		cfg.Webhook.Method = "POST"
	}
	return cfg, nil
}

// DesktopEnabled returns whether the desktop channel is enabled.
func (c Config) DesktopEnabled() bool {
	if c.Desktop.Enabled == nil {
		return true
	}
	return *c.Desktop.Enabled
}

// BellEnabled returns whether the bell channel is enabled.
func (c Config) BellEnabled() bool {
	if c.Bell.Enabled == nil {
		return true
	}
	return *c.Bell.Enabled
}
