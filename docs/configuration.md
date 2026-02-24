# Configuration Reference

attn is configured via a TOML file at `~/.config/attn/config.toml`. If the file doesn't exist, sensible defaults are used.

## Full example

```toml
[context]
mode = "auto"  # "auto", "none", or a fixed string

[desktop]
when = "active"

[bell]
when = "always"

[ntfy]
when = "idle"
server = "https://ntfy.sh"
topic = "my-notifications"
token = "tk_mytoken"

[pushover]
when = "idle"
token = "app-token"
user_key = "user-key"

[webhook]
when = "always"
url = "https://hooks.slack.com/services/T.../B.../xxx"
method = "POST"
headers = { "Content-Type" = "application/json" }

[serve]
socket_path = "/run/user/1000/attn.sock"

[[serve.tunnels]]
name = "devbox"
host = "devbox.example.com"
user = "peter"
remote_socket = "/run/user/1000/attn.sock"
```

## The `when` field

Every channel has a `when` field that controls when it fires:

| Value | Description | Fail behavior |
|-------|-------------|---------------|
| `never` | Channel is disabled | — |
| `active` | Fire when screen is on and you're not looking at the source terminal | Fail-open: fires if detection unavailable |
| `idle` | Fire when screen is off or locked | Fail-closed: does not fire if detection unavailable |
| `always` | Fire unconditionally | — |

### Defaults

| Channel | Default `when` |
|---------|---------------|
| desktop | `active` |
| bell | `never` |
| ntfy | `never` |
| pushover | `never` |
| webhook | `never` |

### Common setups

**Desktop only (default, zero config needed):**
No config file required. Desktop notifications fire when you're not already looking at the terminal.

**Desktop + phone when away:**
```toml
[ntfy]
when = "idle"
topic = "my-notifications"
```

**Always push to phone, skip desktop:**
```toml
[desktop]
when = "never"

[ntfy]
when = "always"
topic = "my-notifications"
```

**Bell in every terminal, desktop when not focused:**
```toml
[desktop]
when = "active"

[bell]
when = "always"
```

## Migration from `enabled`

The old `enabled = true/false` field is still recognized for backward compatibility. If both `when` and `enabled` are present, `when` takes precedence.

Migration rules when only `enabled` is set:

| Channel | `enabled = true` becomes | `enabled = false` becomes |
|---------|--------------------------|---------------------------|
| desktop | `when = "active"` | `when = "never"` |
| bell | `when = "always"` | `when = "never"` |
| ntfy | `when = "always"` | `when = "never"` |
| pushover | `when = "always"` | `when = "never"` |
| webhook | `when = "always"` | `when = "never"` |

## Context

```toml
[context]
mode = "auto"  # default
```

- `auto` — derives `repo:branch` from git (200ms timeout), falls back to directory name
- `none` — no context appended
- Any other string — used as-is

Context can also be overridden per-invocation with `--context` or disabled with `--no-context`.

## Serve (relay server)

```toml
[serve]
socket_path = "/run/user/1000/attn.sock"  # default: $XDG_RUNTIME_DIR/attn.sock

[[serve.tunnels]]
name = "devbox"           # display name for logs
host = "devbox.example.com"
user = "peter"
remote_socket = "/run/user/1000/attn.sock"
identity_file = "~/.ssh/id_ed25519"  # optional
```

See [Remote Relay](remote-relay.md) for details on the relay architecture.
