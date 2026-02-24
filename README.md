# attn

**Attention is all you need.**

A cross-platform notification tool that alerts you when long-running processes — AI agents, builds, deployments — need your attention. Single binary, zero runtime dependencies, multiple notification channels.

## Features

- **Smart focus detection** — automatically suppresses notifications when you're looking at the terminal that triggered them (process-tree matching via X11 `_NET_WM_PID` / Wayland)
- **Screen-aware routing** — send desktop notifications when active, push to your phone when the screen is locked
- **Desktop notifications** — native D-Bus on Linux (no `notify-send` required), osascript on macOS
- **Push notifications** — ntfy.sh, Pushover, generic webhooks (Slack, Discord, etc.)
- **Templatable messages** — use Go templates in title, body, and a configurable prefix (`{{.Repo}}`, `{{.Branch}}`, `{{env "VAR"}}`, etc.)
- **Remote relay** — get notifications from remote servers via SSH socket forwarding
- **Managed SSH tunnels** — auto-maintain tunnels to your remote machines
- **Single binary** — zero runtime dependencies, cross-compiled for Linux, macOS, and Windows

## Install

### From releases

Download the latest binary from [Releases](https://github.com/petersr/attn/releases) and place it in your `$PATH`.

### From source

```bash
go install github.com/petersr/attn@latest
```

## Usage

```bash
# Basic notification
attn send "Build complete"

# With templated title
attn send -t "{{.Repo}}" "Build complete"

# Critical urgency
attn send -u critical "Build FAILED"

# Custom timeout (ms)
attn send -T 10000 "Deployment complete"

# Environment variable in message
attn send -t '{{env "USER"}}' "Task done"
```

### Commands

```
attn send [flags] [message...]   Send a notification
attn serve [flags]               Start the relay server
attn version                     Print version
```

### Send flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | `Notification` | Notification title (supports Go templates) |
| `--urgency` | `-u` | `normal` | `low`, `normal`, or `critical` |
| `--timeout` | `-T` | `5000` | Display timeout in milliseconds |

### Global flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-C` | `~/.config/attn/config.toml` | Config file path |

## Configuration

Create `~/.config/attn/config.toml`:

```toml
[format]
prefix = "[{{.Repo}}:{{.Branch}}] "  # prepended to every message body (default: empty)

[desktop]
when = "active"  # default: fire when screen is on and you're not looking at the source terminal

[bell]
when = "never"  # "never", "active", "idle", or "always"

[ntfy]
when = "idle"  # fire when screen is locked — pushes to your phone
server = "https://ntfy.sh"
topic = "my-attn"
# token = ""  # optional access token

[pushover]
when = "never"
# token = ""
# user_key = ""

[webhook]
when = "never"
# url = "https://hooks.slack.com/services/..."
# method = "POST"
# headers = { "Content-Type" = "application/json" }

[relay]
when = "never"  # enable on remote machines to relay notifications back
# socket_path = ""  # default: $XDG_RUNTIME_DIR/attn.sock
```

### Channel conditions (`when`)

| Value | Behavior |
|-------|----------|
| `never` | Channel is disabled (default for all channels except desktop) |
| `active` | Fire when screen is on, unlocked, and the focused window is **not** in attn's process tree |
| `idle` | Fire when screen is off or locked |
| `always` | Fire unconditionally |

Relay is a regular channel — it participates in the `when` condition system and supports chaining across multiple machines. See [Remote Relay](docs/remote-relay.md).

### Template variables

Title, message body, and `format.prefix` all support Go `text/template` syntax:

| Variable | Description |
|----------|-------------|
| `{{.Dir}}` | Basename of current working directory |
| `{{.Path}}` | Full CWD path |
| `{{.Repo}}` | Git repo name (basename of git toplevel) |
| `{{.Branch}}` | Git branch name |
| `{{env "VAR"}}` | Environment variable lookup |

Example: with `prefix = "[{{.Repo}}:{{.Branch}}] "` in your config, running `attn send "done"` in a git repo produces the body `[myrepo:main] done`.

**Fail-safe behavior:**
- `active` channels **fail open** — if screen state can't be detected, they fire anyway
- `idle` channels **fail closed** — if screen state can't be detected, they don't fire

All channels fire concurrently. A failure in one channel does not affect others.

### How focus detection works

When a channel is set to `when = "active"`, attn walks its own process tree up through `/proc` to find the terminal that spawned it:

```
attn → bash (hook) → claude → bash (shell) → warp (PID 1234)
```

It then gets the focused window's PID (via X11 `_NET_WM_PID` or Wayland D-Bus) and checks if that PID is a direct ancestor in attn's chain. If it is (e.g., the focused window is the Warp terminal that spawned attn), the notification is suppressed — you're already looking at it.

This means **no configuration is needed** for focus suppression. It works automatically when called from any context (Claude Code hooks, shell scripts, CI runners).

## Remote Notifications via SSH

Get notifications on your local machine when something finishes on a remote server.

### 1. Start the relay server locally

```bash
# Run as a systemd user service (recommended)
attn serve --install

# Or run in foreground for debugging
attn serve
```

### 2. Configure SSH tunnels (local config)

Add tunnels to your config and they'll be maintained automatically:

```toml
[serve]
# socket_path = ""  # default: $XDG_RUNTIME_DIR/attn.sock

[[serve.tunnels]]
name = "devbox"
host = "devbox.example.com"
user = "deploy"
# remote_socket_path auto-inferred via ssh id -u

[[serve.tunnels]]
name = "gpu-server"
host = "gpu.internal"
user = "peter"
# remote_socket_path = "/run/user/1000/attn.sock"  # or override explicitly
```

Tunnels use your system `ssh` binary, so they inherit your `~/.ssh/config`, agent, ProxyJump, etc. They auto-reconnect with exponential backoff on disconnect. If `remote_socket_path` is omitted, the remote user's UID is queried via `ssh id -u` and the path is derived as `/run/user/<uid>/attn.sock`.

### 3. Enable relay on the remote machine (remote config)

Install `attn` on the remote server and enable the relay channel:

```toml
[relay]
when = "always"
# socket_path defaults to $XDG_RUNTIME_DIR/attn.sock
```

In most cases no explicit socket paths are needed — both sides default to `/run/user/<uid>/attn.sock`.

```bash
# On the remote server — notification appears on your local desktop
attn send -t "GPU Training" "Epoch 50 complete"
```

### Manual tunnel (alternative)

If you prefer managing SSH connections yourself:

```bash
ssh -R /run/user/2000/attn.sock:/run/user/1000/attn.sock remote-host
```

### Systemd service management

```bash
attn serve --install     # install and enable
attn serve --uninstall   # disable and remove
systemctl --user status attn  # check status
journalctl --user -u attn -f  # view logs
```

## Use with AI Agents

### Claude Code

Add to your [Claude Code hooks](https://docs.anthropic.com/en/docs/claude-code/hooks) configuration (`.claude/settings.json`):

```json
{
  "hooks": {
    "Notification": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "attn send -t 'Claude Code'"
          }
        ]
      }
    ]
  }
}
```

With `when = "active"` (the default for desktop), notifications are automatically suppressed when you're looking at the terminal running Claude Code. No flags needed.

To also get push notifications when your screen is locked, add ntfy to your config:

```toml
[ntfy]
when = "idle"
topic = "my-claude-notifications"
```

## Platform Support

| Platform | Desktop | Focus Detection | Screen Idle | Relay/Tunnels | Push (ntfy, etc.) |
|----------|---------|-----------------|-------------|---------------|-------------------|
| **Linux** | D-Bus (native) | Wayland (GNOME) + X11 | D-Bus ScreenSaver | Unix socket | HTTP |
| **macOS** | osascript* | osascript* | Not yet | Unix socket | HTTP |
| **Windows** | PowerShell* | Not supported | Not supported | Not supported | HTTP |

\* **Experimental** — macOS and Windows desktop notification support is untested. Push channels (ntfy, Pushover, webhook) and the terminal bell work on all platforms. Contributions welcome!

## Documentation

See the [docs/](docs/) directory for detailed documentation:

- [Configuration Reference](docs/configuration.md) — full config format, `when` semantics, migration guide
- [Channels](docs/channels.md) — each notification channel explained
- [Focus Detection](docs/focus-detection.md) — how process-tree matching works
- [Screen Idle Detection](docs/screen-idle.md) — how screen lock/idle detection works
- [Remote Relay](docs/remote-relay.md) — relay architecture, SSH tunnels, systemd setup

## Building

```bash
go build -o attn .

# Cross-compile
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o attn-linux-arm64 .

# With version info
go build -ldflags "-s -w -X main.version=v1.0.0" -o attn .
```

## License

[MIT](LICENSE)
