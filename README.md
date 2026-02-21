# attn

**Attention is all you need.**

A cross-platform notification tool that alerts you when long-running processes — AI agents, builds, deployments — need your attention. Single binary, zero runtime dependencies, multiple notification channels.

## Features

- **Desktop notifications** — native D-Bus on Linux (no `notify-send` required), osascript on macOS
- **Push notifications** — ntfy.sh, Pushover, generic webhooks (Slack, Discord, etc.)
- **Focus suppression** — skip notifications when the relevant window is already focused
- **Auto-context** — automatically includes repo name and git branch so you know *which* task finished
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

# With title and context
attn send -t "My Project" -c "backend:main" "Tests passed"

# Auto-context (default) — derives repo:branch from git
attn send -t "Claude Code" "Done responding"

# Skip if a window is focused
attn send --skip-if-focused "code|vscodium" "Done responding"

# Critical urgency
attn send -u critical "Build FAILED"

# Custom timeout (ms)
attn send -T 10000 "Deployment complete"
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
| `--title` | `-t` | `Notification` | Notification title |
| `--urgency` | `-u` | `normal` | `low`, `normal`, or `critical` |
| `--timeout` | `-T` | `5000` | Display timeout in milliseconds |
| `--context` | `-c` | `auto` | Context string, or `auto` to derive from git |
| `--no-context` | | `false` | Disable context entirely |
| `--skip-if-focused` | | | Regex: suppress if focused window matches |

### Global flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-C` | `~/.config/attn/config.toml` | Config file path |

## Configuration

Create `~/.config/attn/config.toml`:

```toml
[context]
mode = "auto"  # "auto", "none", or a fixed string

[desktop]
enabled = true

[bell]
enabled = true

[ntfy]
enabled = false
server = "https://ntfy.sh"
topic = "my-attn"
# token = ""  # optional access token

[pushover]
enabled = false
# token = ""
# user_key = ""

[webhook]
enabled = false
# url = "https://hooks.slack.com/services/..."
# method = "POST"
# headers = { "Content-Type" = "application/json" }
```

All channels fire concurrently. A failure in one channel does not affect others.

## Remote Notifications via SSH

Get notifications on your local machine when something finishes on a remote server.

### 1. Start the relay server locally

```bash
# Run as a systemd user service (recommended)
attn serve --install

# Or run in foreground for debugging
attn serve
```

### 2. Configure SSH tunnels

Add tunnels to your config and they'll be maintained automatically:

```toml
[serve]
# socket_path = ""  # default: $XDG_RUNTIME_DIR/attn.sock

[[serve.tunnels]]
name = "devbox"
host = "devbox.example.com"
user = "peter"
remote_socket = "/run/user/1000/attn.sock"
# identity_file = "~/.ssh/id_ed25519"  # optional

[[serve.tunnels]]
name = "gpu-server"
host = "gpu.internal"
user = "peter"
remote_socket = "/run/user/1000/attn.sock"
```

Tunnels use your system `ssh` binary, so they inherit your `~/.ssh/config`, agent, ProxyJump, etc. They auto-reconnect with exponential backoff on disconnect.

### 3. Use attn on the remote server

Install `attn` on the remote server and use it normally. It auto-detects the forwarded socket and relays notifications back to your local machine.

```bash
# On the remote server — notification appears on your local desktop
attn send -t "GPU Training" "Epoch 50 complete"
```

### Manual tunnel (alternative)

If you prefer managing SSH connections yourself:

```bash
ssh -R /run/user/1000/attn.sock:/run/user/1000/attn.sock remote-host
```

Or set `ATTN_SOCK` on the remote to point to the forwarded socket.

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
    "notification": [
      {
        "type": "command",
        "command": "attn send -t 'Claude Code' --skip-if-focused 'code|vscodium|cursor'"
      }
    ]
  }
}
```

The auto-context feature will include the repo and branch name in the notification automatically.

## Platform Support

| Platform | Desktop | Focus Detection | Relay/Tunnels | Push (ntfy, etc.) |
|----------|---------|-----------------|---------------|-------------------|
| **Linux** | D-Bus (native) | Wayland (GNOME) + X11 | Unix socket | HTTP |
| **macOS** | osascript* | osascript* | Unix socket | HTTP |
| **Windows** | PowerShell* | Not supported | Not supported | HTTP |

\* **Experimental** — macOS and Windows desktop notification and focus detection support is untested. Push channels (ntfy, Pushover, webhook) and the terminal bell work on all platforms. Contributions welcome!

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
