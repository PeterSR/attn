# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
go build -o attn .                          # Build binary
go build -ldflags "-s -w" -o attn .         # Build stripped (smaller)
go vet ./...                                # Static analysis
go test -race ./...                         # Run tests with race detector
golangci-lint run ./...                     # Lint (same as CI, v1.64.8)
```

**After any significant change, run the full CI pipeline locally before considering the work done:**
```bash
make ci
```

Cross-compile (all targets use CGO_ENABLED=0):
```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o attn-linux-arm64 .
```

Version injection at build time:
```bash
go build -ldflags "-s -w -X main.version=v1.0.0" -o attn .
```

## Architecture

**CLI framework**: `alecthomas/kong` with struct-based command definitions in `cmd/`. Three commands: `send`, `serve`, `version`. No implicit default command — `send` must be explicit.

**Channel system** (`internal/channel/`): Core abstraction. All notification backends implement:
```go
type Channel interface {
    Name() string
    Send(ctx context.Context, n notification.Notification) error
}
```
`Dispatch()` fires all enabled channels concurrently with 10s per-channel timeout. Failures are collected, not fatal.

Channels: `desktop` (D-Bus), `bell` (\a), `ntfy`, `pushover`, `webhook`, `remote` (Unix socket relay).

**Platform-specific code**: Uses Go build tag file suffixes (`_linux.go`, `_darwin.go`, `_windows.go`, `_other.go`). Linux is the primary platform with native D-Bus and X11 support. macOS and Windows files are experimental and marked as such in their header comments.

**Desktop notifications on Linux** (`internal/channel/desktop/desktop_linux.go`): Direct D-Bus call to `org.freedesktop.Notifications.Notify` via `godbus/dbus/v5`. No dependency on `notify-send`.

**D-Bus session bus** (`internal/dbus/session.go`): Helper with fallback logic — tries standard `DBUS_SESSION_BUS_ADDRESS` first, then constructs address from `$XDG_RUNTIME_DIR/bus`. Critical for non-interactive contexts (systemd services, cron, agent hooks).

**Focus detection** (`internal/focus/`): Linux router tries Wayland (GNOME Shell D-Bus extension) then X11 (`jezek/xgbutil` EWMH/ICCCM). Returns empty string on failure — suppression only triggers on positive match.

**Remote relay**: `internal/relay/server.go` listens on Unix socket, `internal/channel/remote/client.go` sends to it. Wire format is JSON-lines (`internal/relay/protocol.go`). Auto-detected on remote machines via `$ATTN_SOCK` env var or `$SSH_CLIENT` + socket existence.

**SSH tunnel manager** (`internal/tunnel/tunnel.go`): Spawns `ssh -N -R` processes per configured tunnel. Uses system `ssh` binary (inherits user's config/agent/ProxyJump). Auto-reconnects with exponential backoff (1s to 60s cap).

**Auto-context** (`internal/autocontext/`): Derives `repo:branch` from git with 200ms timeout. Falls back to directory basename. Git not installed is not an error.

**Config** (`internal/config/`): TOML at `~/.config/attn/config.toml`. Desktop and bell enabled by default (use `*bool` for tri-state). Missing config file is not an error.

## Release

GoReleaser on tag push (`v*`). Produces 6 binaries: linux/darwin × amd64/arm64, windows × amd64/arm64. All statically linked, no CGO.
