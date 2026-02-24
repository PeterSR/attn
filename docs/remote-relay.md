# Remote Relay

attn can relay notifications from remote servers to your local machine via SSH socket forwarding. The relay is a regular notification channel, so it participates in the `when` condition system and can be chained across multiple machines.

## Architecture

There are two sides to configure:

- **Local side** (`[serve]`): The relay server that listens on a Unix socket and dispatches received notifications through local channels. Tunnel configuration specifies which remote machines to connect to.
- **Remote side** (`[relay]`): A notification channel that sends to a Unix socket created by the SSH tunnel.

```
Remote machine (devbox)                   Local machine
┌──────────────────────┐                  ┌───────────────────────┐
│ attn send "Done"     │                  │ attn serve            │
│   │                  │                  │   │                   │
│   ▼                  │                  │   ▼                   │
│ [relay] channel      │                  │ DispatchFunc          │
│   │                  │                  │   │                   │
│   ▼                  │                  │   ▼                   │
│ Unix socket ─────────┼── SSH tunnel ────┼─► Unix socket         │
│                      │                  │   │                   │
│                      │                  │   ▼                   │
│                      │                  │ Desktop / ntfy / etc  │
└──────────────────────┘                  └───────────────────────┘
```

### Chaining (A → B → C)

Because relay is a regular channel, the relay server dispatches received notifications through the full channel pipeline — including potentially another relay channel. This enables multi-hop forwarding:

```
Machine C              Machine B              Machine A (desktop)
┌──────────┐           ┌──────────┐           ┌──────────┐
│ attn send│──relay──►│ attn serve│──relay──►│ attn serve│
│          │           │ + relay  │           │          │
└──────────┘           └──────────┘           └──────────┘
                                               Desktop ✓
```

A hop counter in the wire protocol prevents infinite loops. Notifications are dropped after 10 hops.

## Wire protocol

JSON-lines over Unix socket (`internal/relay/protocol.go`):

**Request** (one JSON object per line):
```json
{"v":1,"type":"notify","notify":{"title":"Build","body":"Complete","urgency":"normal","timeout_ms":5000},"hops":1}
```

**Response:**
```json
{"ok":true}
```

The `hops` field tracks how many relay hops a notification has taken. Each relay channel increments it before sending.

## Setup

### 1. Start the relay server locally

```bash
# As a systemd user service (recommended)
attn serve --install

# Or foreground
attn serve
```

Default socket: `$XDG_RUNTIME_DIR/attn.sock` (typically `/run/user/<uid>/attn.sock`).

### 2. Configure SSH tunnels (local config)

```toml
[serve]
# socket_path defaults to $XDG_RUNTIME_DIR/attn.sock

[[serve.tunnels]]
name = "devbox"
host = "devbox.example.com"
user = "deploy"
# remote_socket_path auto-inferred as /run/user/<remote-uid>/attn.sock
```

If `remote_socket_path` is omitted, the tunnel manager runs `ssh <host> id -u` to determine the remote user's UID and derives the path as `/run/user/<uid>/attn.sock`. You can override it explicitly if needed.

The tunnel manager spawns `ssh -N -R <remote_socket_path>:<local_socket> <host>` for each tunnel. It:
- Uses the system `ssh` binary (inherits `~/.ssh/config`, agent, ProxyJump)
- Auto-reconnects with exponential backoff (1s to 60s cap)
- Runs as long as the relay server is running

### 3. Enable relay on the remote machine (remote config)

```toml
[relay]
when = "always"
# socket_path defaults to $XDG_RUNTIME_DIR/attn.sock
```

Both `relay.socket_path` and the tunnel's remote socket path default to `/run/user/<uid>/attn.sock`. In most cases no explicit socket paths are needed — just set `when = "always"` on the remote.

### 4. Use on the remote machine

```bash
# On the remote server — notification appears on your local desktop
attn send -t "GPU Training" "Epoch 50 complete"
```

### Manual SSH tunnel

If you prefer managing SSH connections yourself:

```bash
ssh -R /run/user/2000/attn.sock:/run/user/1000/attn.sock remote-host
```

Or set `ATTN_SOCK` on the remote to override the socket path.

## Systemd service

```bash
attn serve --install     # creates ~/.config/systemd/user/attn.service, enables and starts
attn serve --uninstall   # stops, disables, and removes the service file

systemctl --user status attn   # check status
journalctl --user -u attn -f   # view logs
```

## Implementation

| Component | File |
|-----------|------|
| Relay server | `internal/relay/server.go` |
| Relay channel | `internal/channel/remote/client.go` |
| Wire protocol | `internal/relay/protocol.go` |
| Tunnel manager | `internal/tunnel/tunnel.go` |
| Channel builder | `cmd/channels.go` — `buildChannelEntries()` |
| Config | `internal/config/config.go` — `RelayConfig` |
