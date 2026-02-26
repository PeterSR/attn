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
	"github.com/petersr/attn/internal/proctree"
	"github.com/petersr/attn/internal/render"
)

// SendCmd is the default command that sends a notification.
type SendCmd struct {
	Title   string   `short:"t" default:"Notification" help:"Notification title (supports Go templates: {{.Repo}}, {{.Branch}}, etc.)."`
	Urgency string   `short:"u" default:"normal" enum:"low,normal,critical" help:"Urgency level."`
	Timeout int      `short:"T" default:"5000" help:"Timeout in milliseconds."`
	Verbose bool     `short:"v" help:"Print channel evaluation details to stderr."`
	Message []string `arg:"" optional:"" help:"Notification message body (supports Go templates)."`
}

func (s *SendCmd) Run(globals *CLI) error {
	cfg, err := config.Load(globals.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: warning: config load: %v\n", err)
		cfg = config.Default()
	}

	// Gather context and resolve process label.
	info := autocontext.Gather()
	if len(cfg.Processes) > 0 {
		chain := proctree.AncestorsNamed(os.Getpid())
		info.Process = proctree.MatchKnown(chain, cfg.Processes)
	}
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

	if s.Verbose {
		fmt.Fprintf(os.Stderr, "attn: screen: idle=%v inProcessTree=%v detectionOK=%v\n",
			state.Idle, state.InProcessTree, state.DetectionOK)

		results, err := channel.DispatchFilteredVerbose(context.Background(), entries, state, n)
		for _, r := range results {
			if r.Fired {
				if r.Err != nil {
					fmt.Fprintf(os.Stderr, "attn: %s(when=%s): error: %v\n", r.Name, r.When, r.Err)
				} else {
					fmt.Fprintf(os.Stderr, "attn: %s(when=%s): sent\n", r.Name, r.When)
				}
			} else {
				fmt.Fprintf(os.Stderr, "attn: %s(when=%s): skipped (%s)\n", r.Name, r.When, r.Reason)
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "attn: %v\n", err)
		}
	} else {
		if err := channel.DispatchFiltered(context.Background(), entries, state, n); err != nil {
			fmt.Fprintf(os.Stderr, "attn: %v\n", err)
		}
	}
	return nil
}
