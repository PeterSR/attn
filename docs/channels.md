# Channels

attn sends notifications through multiple channels concurrently. Each channel can be independently configured with a `when` condition. A failure in one channel does not affect others.

## Desktop

Native desktop notifications using the system notification daemon.

| Platform | Implementation |
|----------|---------------|
| Linux | Direct D-Bus call to `org.freedesktop.Notifications.Notify` via godbus. No dependency on `notify-send`. |
| macOS | `osascript` AppleScript (experimental) |
| Windows | BurntToast PowerShell module with .NET fallback (experimental) |

**Config:**
```toml
[desktop]
when = "active"  # default
```

Desktop maps urgency levels to the notification daemon's urgency hints (low=0, normal=1, critical=2).

## Bell

Sends the terminal bell character (`\a`) to stderr. Works in any terminal emulator.

**Config:**
```toml
[bell]
when = "always"
```

The bell channel is useful as a complement to desktop notifications — it works even when the terminal is on a different workspace or monitor. Many terminal emulators can be configured to flash or bounce when they receive a bell.

## ntfy

Push notifications via [ntfy.sh](https://ntfy.sh) (or a self-hosted ntfy server). Ideal for phone notifications when you're away from your desk.

**Config:**
```toml
[ntfy]
when = "idle"
server = "https://ntfy.sh"  # default
topic = "my-notifications"   # required
token = "tk_mytoken"         # optional, for access-controlled topics
```

Maps urgency to ntfy priority: low→2 (low), normal→3 (default), critical→5 (max).

## Pushover

Push notifications via [Pushover](https://pushover.net/).

**Config:**
```toml
[pushover]
when = "idle"
token = "app-token"    # required: your Pushover application token
user_key = "user-key"  # required: your Pushover user key
```

Maps urgency to Pushover priority: low→-1, normal→0, critical→1.

## Webhook

Generic HTTP webhook. Works with Slack, Discord, or any service that accepts HTTP requests.

**Config:**
```toml
[webhook]
when = "always"
url = "https://hooks.slack.com/services/..."  # required
method = "POST"                                # default
headers = { "Content-Type" = "application/json" }
```

Sends a JSON payload:
```json
{
  "title": "Notification",
  "body": "Build complete",
  "urgency": "normal"
}
```

## Relay

Sends notifications to a Unix socket where a relay server (`attn serve`) is listening. Used on remote machines to forward notifications back to your local workstation via SSH tunnels.

**Config:**
```toml
[relay]
when = "always"
# socket_path = ""  # default: $XDG_RUNTIME_DIR/attn.sock
```

The relay channel participates in the `when` condition system like any other channel. It supports chaining — a relay server can re-dispatch received notifications through its own relay channel to forward them further. A hop counter (max 10) prevents infinite loops.

See [Remote Relay](remote-relay.md) for the full architecture.

## Channel dispatch

All enabled channels fire concurrently with a 10-second per-channel timeout. Screen state (active/idle) and process-tree focus are evaluated once before dispatch, not per-channel.
