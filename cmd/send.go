package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/petersr/attn/internal/autocontext"
	"github.com/petersr/attn/internal/channel"
	"github.com/petersr/attn/internal/channel/bell"
	"github.com/petersr/attn/internal/channel/desktop"
	"github.com/petersr/attn/internal/channel/ntfy"
	"github.com/petersr/attn/internal/channel/pushover"
	"github.com/petersr/attn/internal/channel/remote"
	"github.com/petersr/attn/internal/channel/webhook"
	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/focus"
	"github.com/petersr/attn/internal/notification"
)

// SendCmd is the default command that sends a notification.
type SendCmd struct {
	Title          string   `short:"t" default:"Notification" help:"Notification title."`
	Urgency        string   `short:"u" default:"normal" enum:"low,normal,critical" help:"Urgency level."`
	Timeout        int      `short:"T" default:"5000" help:"Timeout in milliseconds."`
	Context        string   `short:"c" default:"auto" help:"Context identifier (repo/branch/agent). Use 'auto' to derive from git/PWD."`
	NoContext      bool     `help:"Disable context entirely."`
	SkipIfFocused  string   `help:"Regex: suppress notification if focused window matches."`
	Message        []string `arg:"" optional:"" help:"Notification message body."`
}

func (s *SendCmd) Run(globals *CLI) error {
	cfg, err := config.Load(globals.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: warning: config load: %v\n", err)
		cfg = config.Default()
	}

	// Focus suppression.
	if focus.ShouldSuppress(s.SkipIfFocused) {
		return nil
	}

	// Resolve context.
	ctx := s.resolveContext(cfg)

	// Build notification.
	n := notification.Notification{
		Title:     s.Title,
		Body:      strings.Join(s.Message, " "),
		Urgency:   notification.Urgency(s.Urgency),
		TimeoutMS: s.Timeout,
		Context:   ctx,
	}

	// Check if we should relay through a Unix socket.
	if sock := detectRelaySocket(cfg); sock != "" {
		rc := remote.New(sock)
		if err := rc.Send(context.Background(), n); err != nil {
			fmt.Fprintf(os.Stderr, "attn: relay failed: %v\n", err)
			// Fall through to local channels.
		} else {
			return nil
		}
	}

	// Build enabled channels.
	channels := buildChannels(cfg)
	if len(channels) == 0 {
		fmt.Fprintf(os.Stderr, "attn: no channels enabled. Message: %s — %s\n", n.Title, n.FormatBody())
		return nil
	}

	if err := channel.Dispatch(context.Background(), channels, n); err != nil {
		fmt.Fprintf(os.Stderr, "attn: %v\n", err)
	}
	return nil
}

func (s *SendCmd) resolveContext(cfg config.Config) string {
	if s.NoContext {
		return ""
	}
	if s.Context != "auto" {
		return s.Context
	}
	if cfg.Context.Mode == "none" {
		return ""
	}
	if cfg.Context.Mode != "" && cfg.Context.Mode != "auto" {
		return cfg.Context.Mode
	}
	return autocontext.Derive()
}

func buildChannels(cfg config.Config) []channel.Channel {
	var channels []channel.Channel

	if cfg.DesktopEnabled() {
		channels = append(channels, desktop.New())
	}
	if cfg.BellEnabled() {
		channels = append(channels, bell.New())
	}
	if cfg.Ntfy.Enabled && cfg.Ntfy.Topic != "" {
		channels = append(channels, ntfy.New(cfg.Ntfy.Server, cfg.Ntfy.Topic, cfg.Ntfy.Token))
	}
	if cfg.Pushover.Enabled && cfg.Pushover.Token != "" && cfg.Pushover.UserKey != "" {
		channels = append(channels, pushover.New(cfg.Pushover.Token, cfg.Pushover.UserKey))
	}
	if cfg.Webhook.Enabled && cfg.Webhook.URL != "" {
		channels = append(channels, webhook.New(cfg.Webhook.URL, cfg.Webhook.Method, cfg.Webhook.Headers))
	}

	return channels
}

func detectRelaySocket(cfg config.Config) string {
	// 1. Explicit env var.
	if sock := os.Getenv("ATTN_SOCK"); sock != "" {
		if _, err := os.Stat(sock); err == nil {
			return sock
		}
	}
	// 2. Config file.
	if cfg.Serve.SocketPath != "" {
		// Only use if we're in an SSH session (don't relay locally).
		if os.Getenv("SSH_CLIENT") != "" {
			if _, err := os.Stat(cfg.Serve.SocketPath); err == nil {
				return cfg.Serve.SocketPath
			}
		}
	}
	// 3. Default path in SSH session.
	if os.Getenv("SSH_CLIENT") != "" {
		sock := defaultSocketPath()
		if _, err := os.Stat(sock); err == nil {
			return sock
		}
	}
	return ""
}

func defaultSocketPath() string {
	xdg := os.Getenv("XDG_RUNTIME_DIR")
	if xdg == "" {
		xdg = fmt.Sprintf("/run/user/%d", os.Getuid())
	}
	return filepath.Join(xdg, "attn.sock")
}
