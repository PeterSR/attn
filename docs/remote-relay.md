# Remote Relay

attn can relay notifications from remote servers to your local machine via SSH socket forwarding.

## Architecture

```
Remote machine                          Local machine
┌──────────────────┐                    ┌──────────────────────┐
│ attn send "Done" │                    │ attn serve           │
│   │              │                    │   │                  │
│   ▼              │                    │   ▼                  │
│ Unix socket ─────┼── SSH tunnel ──────┼─► Unix socket        │
│ (JSON-lines)     │                    │   │                  │
│                  │                    │   ▼                  │
│                  │                    │ Desktop / Bell / etc │
└──────────────────┘                    └──────────────────────┘
```

### Components

1. **Relay server** (`attn serve`): Listens on a Unix socket, receives notifications, dispatches them through local channels.
2. **Remote client**: When `attn send` detects it's in an SSH session and a relay socket exists, it sends the notification over the socket instead of firing local channels.
3. **SSH tunnel manager**: Optionally maintains `ssh -N -R` tunnels to configured remote machines.

## Wire protocol

JSON-lines over Unix socket (`internal/relay/protocol.go`):

**Request** (one JSON object per line):
```json
{"title":"Build","body":"Complete","urgency":"normal","timeout_ms":5000,"context":"myrepo:main"}
```

**Response:**
```json
{"ok":true}
```

## Setup

### 1. Start the relay server locally

```bash
# As a systemd user service (recommended)
attn serve --install

# Or foreground
attn serve
```

Default socket: `$XDG_RUNTIME_DIR/attn.sock` (typically `/run/user/1000/attn.sock`).

### 2. Configure SSH tunnels

```toml
[serve]
socket_path = "/run/user/1000/attn.sock"

[[serve.tunnels]]
name = "devbox"
host = "devbox.example.com"
user = "peter"
remote_socket = "/run/user/1000/attn.sock"
# identity_file = "~/.ssh/id_ed25519"
```

The tunnel manager spawns `ssh -N -R <remote_socket>:<local_socket> <host>` for each tunnel. It:
- Uses the system `ssh` binary (inherits `~/.ssh/config`, agent, ProxyJump)
- Auto-reconnects with exponential backoff (1s to 60s cap)
- Runs as long as the relay server is running

### 3. Use on the remote machine

Just run `attn send` normally. Socket detection order:

1. `$ATTN_SOCK` environment variable
2. `serve.socket_path` from config (only in SSH sessions)
3. `$XDG_RUNTIME_DIR/attn.sock` (only in SSH sessions)

### Manual SSH tunnel

If you prefer managing SSH connections yourself:

```bash
ssh -R /run/user/1000/attn.sock:/run/user/1000/attn.sock remote-host
```

Or set `ATTN_SOCK` on the remote to point to the forwarded socket path.

## Relay server channels

The relay server dispatches notifications through all configured channels (any channel with `when` not set to `never`). It does **not** evaluate `when` conditions at dispatch time — the assumption is that if a notification was relayed from a remote machine, it should always be delivered locally. The `when` condition evaluation happens on the originating machine.

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
| Remote client | `internal/channel/remote/client.go` |
| Wire protocol | `internal/relay/protocol.go` |
| Tunnel manager | `internal/tunnel/tunnel.go` |
| Socket detection | `cmd/send.go` — `detectRelaySocket()` |
| Server channels | `cmd/serve.go` — `buildServerChannels()` |
