# Configuration Reference

attn is configured via a TOML file at `~/.config/attn/config.toml`. If the file doesn't exist, sensible defaults are used.

## Full example

```toml
[format]
prefix = "[{{.Repo}}:{{.Branch}}] "  # prepended to every message body (default: empty)

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

[relay]
when = "always"
# socket_path = ""  # default: $XDG_RUNTIME_DIR/attn.sock

[serve]
socket_path = "/run/user/1000/attn.sock"

[[serve.tunnels]]
name = "devbox"
host = "devbox.example.com"
user = "peter"
remote_socket_path = "/run/user/1000/attn.sock"
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
| relay | `never` |

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

## Format

```toml
[format]
prefix = "[{{.Repo}}:{{.Branch}}] "
```

The `prefix` template is rendered and prepended to every notification body. Default is empty (no prefix).

### Template variables

Title (`--title`), message body, and `format.prefix` all support Go [`text/template`](https://pkg.go.dev/text/template) syntax:

| Variable | Description |
|----------|-------------|
| `{{.Dir}}` | Basename of current working directory |
| `{{.Path}}` | Full CWD path |
| `{{.Repo}}` | Git repo name (basename of git toplevel, 200ms timeout) |
| `{{.Branch}}` | Git branch name |
| `{{env "VAR"}}` | Environment variable lookup |

If a template fails to parse or execute, the literal string is used unchanged and a warning is printed to stderr.

## Relay (channel)

The relay channel sends notifications to a local Unix socket, where a relay server (`attn serve`) can receive and re-dispatch them. This is used on remote machines to forward notifications back to your local workstation.

```toml
[relay]
when = "always"                              # required to enable
socket_path = "/run/user/2000/attn.sock"     # default: $XDG_RUNTIME_DIR/attn.sock
```

The `socket_path` must match the tunnel's remote socket path. In most cases, both default to `/run/user/<uid>/attn.sock` and no explicit configuration is needed.

Relay supports chaining (A → B → C). A hop counter prevents infinite loops — notifications are dropped after 10 hops.

See [Remote Relay](remote-relay.md) for the full architecture.

## Serve (relay server)

```toml
[serve]
socket_path = "/run/user/1000/attn.sock"  # default: $XDG_RUNTIME_DIR/attn.sock

[[serve.tunnels]]
name = "devbox"           # display name for logs
host = "devbox.example.com"
user = "peter"
# remote_socket_path = ""  # optional: auto-inferred as /run/user/<remote-uid>/attn.sock
# identity_file = "~/.ssh/id_ed25519"  # optional
```

If `remote_socket_path` is omitted, the tunnel manager runs `ssh <host> id -u` to determine the remote user's UID and derives the path as `/run/user/<uid>/attn.sock`.

See [Remote Relay](remote-relay.md) for details on the relay architecture.
