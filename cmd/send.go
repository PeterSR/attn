package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/petersr/attn/internal/autocontext"
	"github.com/petersr/attn/internal/channel"
	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/marker"
	"github.com/petersr/attn/internal/notification"
	"github.com/petersr/attn/internal/proctree"
	"github.com/petersr/attn/internal/render"
)

// SendCmd is the default command that sends a notification.
type SendCmd struct {
	Title   string            `short:"t" default:"Notification" help:"Notification title (supports Go templates: {{.Repo}}, {{.Branch}}, etc.)."`
	Urgency string            `short:"u" default:"normal" enum:"low,normal,critical" help:"Urgency level."`
	Timeout int               `short:"T" default:"5000" help:"Timeout in milliseconds."`
	Verbose bool              `short:"v" help:"Print channel evaluation details to stderr."`
	When    map[string]string `short:"w" help:"Override when condition per channel (e.g. desktop=always)."`
	Via     string            `help:"Send only via this channel, bypassing all others and any when conditions."`
	Message []string          `arg:"" optional:"" help:"Notification message body (supports Go templates)."`
}

func (s *SendCmd) Run(globals *CLI) error {
	cfg, err := config.Load(globals.ConfigFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "attn: warning: config load: %v\n", err)
		cfg = config.Default()
	}

	if err := applyWhenOverrides(&cfg, s.When); err != nil {
		return err
	}

	if err := applyViaOverride(&cfg, s.Via); err != nil {
		return err
	}

	// Build channel entries with When conditions (hops=0 for direct send).
	entries := buildChannelEntries(cfg, 0)
	if len(entries) == 0 {
		// Render with empty info before bailing so the user sees something useful.
		fallbackInfo := autocontext.Gather()
		fmt.Fprintf(os.Stderr, "attn: no channels configured. Message: %s — %s\n",
			render.Render(s.Title, fallbackInfo),
			render.Render(strings.Join(s.Message, " "), fallbackInfo))
		return nil
	}

	// Evaluate screen state once (hops=0 for direct send), then overlay
	// marker and global env state.
	state := channel.DetectScreenState(entries, 0)
	applyMarkerOverlay(&state, cfg, 0)

	// Gather context and resolve process label. Marker label wins; otherwise
	// fall back to the [processes] table.
	info := autocontext.Gather()
	if state.MarkerLabel != "" {
		info.Process = state.MarkerLabel
	} else if len(cfg.Processes) > 0 {
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

	if s.Verbose {
		fmt.Fprintf(os.Stderr, "attn: screen: idle=%v inProcessTree=%v detectionOK=%v\n",
			state.Idle, state.InProcessTree, state.DetectionOK)
		if state.ForceSuppress || state.ForceFire || state.MarkerVerdict != marker.VerdictFallthrough {
			fmt.Fprintf(os.Stderr, "attn: marker: %s\n", state.MarkerReason)
		}

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

// applyWhenOverrides applies ad-hoc --when overrides to the loaded config.
func applyWhenOverrides(cfg *config.Config, overrides map[string]string) error {
	for ch, val := range overrides {
		w := config.When(val)
		if !w.Valid() || w == "" {
			return fmt.Errorf("invalid when value %q for channel %q (valid: never, active, idle, always)", val, ch)
		}
		switch ch {
		case "desktop":
			cfg.Desktop.When = w
		case "bell":
			cfg.Bell.When = w
		case "ntfy":
			cfg.Ntfy.When = w
		case "pushover":
			cfg.Pushover.When = w
		case "webhook":
			cfg.Webhook.When = w
		case "relay":
			cfg.Relay.When = w
		default:
			return fmt.Errorf("unknown channel %q (valid: desktop, bell, ntfy, pushover, webhook, relay)", ch)
		}
	}
	return nil
}

// applyViaOverride forces a single channel to fire and silences all others.
// Errors if the channel name is unknown or the channel lacks required config.
func applyViaOverride(cfg *config.Config, name string) error {
	if name == "" {
		return nil
	}
	switch name {
	case "desktop":
		cfg.Desktop.When = config.WhenAlways
	case "bell":
		cfg.Bell.When = config.WhenAlways
	case "ntfy":
		if cfg.Ntfy.Topic == "" {
			return fmt.Errorf("channel %q is not configured (missing ntfy.topic)", name)
		}
		cfg.Ntfy.When = config.WhenAlways
	case "pushover":
		if cfg.Pushover.Token == "" || cfg.Pushover.UserKey == "" {
			return fmt.Errorf("channel %q is not configured (missing pushover.token or pushover.user_key)", name)
		}
		cfg.Pushover.When = config.WhenAlways
	case "webhook":
		if cfg.Webhook.URL == "" {
			return fmt.Errorf("channel %q is not configured (missing webhook.url)", name)
		}
		cfg.Webhook.When = config.WhenAlways
	case "relay":
		cfg.Relay.When = config.WhenAlways
	default:
		return fmt.Errorf("unknown channel %q (valid: desktop, bell, ntfy, pushover, webhook, relay)", name)
	}

	if name != "desktop" {
		cfg.Desktop.When = config.WhenNever
	}
	if name != "bell" {
		cfg.Bell.When = config.WhenNever
	}
	if name != "ntfy" {
		cfg.Ntfy.When = config.WhenNever
	}
	if name != "pushover" {
		cfg.Pushover.When = config.WhenNever
	}
	if name != "webhook" {
		cfg.Webhook.When = config.WhenNever
	}
	if name != "relay" {
		cfg.Relay.When = config.WhenNever
	}
	return nil
}
