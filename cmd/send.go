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
	"github.com/petersr/attn/internal/render"
)

// SendCmd is the default command that sends a notification.
type SendCmd struct {
	Title   string   `short:"t" default:"Notification" help:"Notification title (supports Go templates: {{.Repo}}, {{.Branch}}, etc.)."`
	Urgency string   `short:"u" default:"normal" enum:"low,normal,critical" help:"Urgency level."`
	Timeout int      `short:"T" default:"5000" help:"Timeout in milliseconds."`
	Message []string `arg:"" optional:"" help:"Notification message body (supports Go templates)."`
}

func (s *SendCmd) Run(globals *CLI) error {
	cfg, err := config.Load(globals.Config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: warning: config load: %v\n", err)
		cfg = config.Default()
	}

	// Gather context and render templates.
	info := autocontext.Gather()
	title := render.Render(s.Title, info)
	body := render.Render(strings.Join(s.Message, " "), info)
	prefix := render.Render(cfg.Format.Prefix, info)

	// Build notification.
	n := notification.Notification{
		Title:     title,
		Body:      prefix + body,
		Urgency:   notification.Urgency(s.Urgency),
		TimeoutMS: s.Timeout,
	}

	// Build channel entries with When conditions (hops=0 for direct send).
	entries := buildChannelEntries(cfg, 0)
	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "attn: no channels configured. Message: %s — %s\n", n.Title, n.Body)
		return nil
	}

	// Evaluate screen state once (hops=0 for direct send).
	state := channel.DetectScreenState(entries, 0)

	if err := channel.DispatchFiltered(context.Background(), entries, state, n); err != nil {
		fmt.Fprintf(os.Stderr, "attn: %v\n", err)
	}
	return nil
}
