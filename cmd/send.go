package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/petersr/attn/internal/autocontext"
	"github.com/petersr/attn/internal/channel"
	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/notification"
)

// SendCmd is the default command that sends a notification.
type SendCmd struct {
	Title     string   `short:"t" default:"Notification" help:"Notification title."`
	Urgency   string   `short:"u" default:"normal" enum:"low,normal,critical" help:"Urgency level."`
	Timeout   int      `short:"T" default:"5000" help:"Timeout in milliseconds."`
	Context   string   `short:"c" default:"auto" help:"Context identifier (repo/branch/agent). Use 'auto' to derive from git/PWD."`
	NoContext bool     `help:"Disable context entirely."`
	Message   []string `arg:"" optional:"" help:"Notification message body."`
}

func (s *SendCmd) Run(globals *CLI) error {
	cfg, err := config.Load(globals.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: warning: config load: %v\n", err)
		cfg = config.Default()
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

	// Build channel entries with When conditions (hops=0 for direct send).
	entries := buildChannelEntries(cfg, 0)
	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "attn: no channels configured. Message: %s — %s\n", n.Title, n.FormatBody())
		return nil
	}

	// Evaluate screen state once.
	state := channel.DetectScreenState(entries)

	if err := channel.DispatchFiltered(context.Background(), entries, state, n); err != nil {
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
