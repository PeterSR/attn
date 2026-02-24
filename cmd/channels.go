package cmd

import (
	"github.com/petersr/attn/internal/channel"
	"github.com/petersr/attn/internal/channel/bell"
	"github.com/petersr/attn/internal/channel/desktop"
	"github.com/petersr/attn/internal/channel/ntfy"
	"github.com/petersr/attn/internal/channel/pushover"
	"github.com/petersr/attn/internal/channel/remote"
	"github.com/petersr/attn/internal/channel/webhook"
	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/relay"
)

// buildChannelEntries creates channel entries from config, including relay
// if configured and within the hop limit.
func buildChannelEntries(cfg config.Config, hops int) []channel.Entry {
	var entries []channel.Entry

	if cfg.Desktop.When != config.WhenNever {
		entries = append(entries, channel.Entry{
			Channel: desktop.New(),
			When:    channel.When(cfg.Desktop.When),
		})
	}
	if cfg.Bell.When != config.WhenNever {
		entries = append(entries, channel.Entry{
			Channel: bell.New(),
			When:    channel.When(cfg.Bell.When),
		})
	}
	if cfg.Ntfy.When != config.WhenNever && cfg.Ntfy.Topic != "" {
		entries = append(entries, channel.Entry{
			Channel: ntfy.New(cfg.Ntfy.Server, cfg.Ntfy.Topic, cfg.Ntfy.Token),
			When:    channel.When(cfg.Ntfy.When),
		})
	}
	if cfg.Pushover.When != config.WhenNever && cfg.Pushover.Token != "" && cfg.Pushover.UserKey != "" {
		entries = append(entries, channel.Entry{
			Channel: pushover.New(cfg.Pushover.Token, cfg.Pushover.UserKey),
			When:    channel.When(cfg.Pushover.When),
		})
	}
	if cfg.Webhook.When != config.WhenNever && cfg.Webhook.URL != "" {
		entries = append(entries, channel.Entry{
			Channel: webhook.New(cfg.Webhook.URL, cfg.Webhook.Method, cfg.Webhook.Headers),
			When:    channel.When(cfg.Webhook.When),
		})
	}

	// Relay channel: only include if configured and within hop limit.
	if cfg.Relay.When != config.WhenNever && hops < relay.MaxHops {
		socketPath := cfg.Relay.SocketPath
		if socketPath == "" {
			socketPath = relay.DefaultSocketPath()
		}
		entries = append(entries, channel.Entry{
			Channel: remote.New(socketPath, hops),
			When:    channel.When(cfg.Relay.When),
		})
	}

	return entries
}
