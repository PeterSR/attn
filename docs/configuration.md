# Configuration Reference

attn is configured via a TOML file at `~/.config/attn/config.toml`. If the file doesn't exist, sensible defaults are used.

The config file path is determined by (in order): `--config` / `-C` flag, `ATTN_CONFIG_PATH` env var, or the default `~/.config/attn/config.toml`.

## CLI configuration

You can read and write config values from the command line:

```bash
attn config set <key> <value>   # Set a value
attn config get <key>           # Get the effective value (including defaults)
attn config path                # Print the config file path
```

### Supported keys

| Key | Validation |
|-----|-----------|
| `format.prefix` | any string |
| `desktop.when` | `never`, `active`, `idle`, `always` |
| `bell.when` | same |
| `ntfy.when`, `ntfy.server`, `ntfy.topic`, `ntfy.token` | `when` validated, rest any string |
| `pushover.when`, `pushover.token`, `pushover.user_key` | same |
| `webhook.when`, `webhook.url`, `webhook.method` | same |
| `relay.when`, `relay.socket_path` | same |
| `processes.<name>` | any string (friendly label for a process comm name) |

### Examples

```bash
# Quick ntfy setup
attn config set ntfy.topic my-notifications
attn config set ntfy.when idle

# Check effective values
attn config get ntfy.server     # → https://ntfy.sh (default)
attn config get desktop.when    # → active (default)

# Add process labels (use 'attn proctree' to discover comm names)
attn config set processes.code "VS Code"
attn config set processes.warp "Warp"
attn config get processes.code  # → VS Code

# Or use interactive mode to browse and label processes
attn proctree -i
```

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

[processes]
code = "VS Code"
warp = "Warp"
alacritty = "Alacritty"

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
| `{{.Process}}` | Friendly label of the parent process (from `[processes]` config) |
| `{{env "VAR"}}` | Environment variable lookup |

If a template fails to parse or execute, the literal string is used unchanged and a warning is printed to stderr.

## Process detection

The `[processes]` section maps Linux process comm names to friendly labels. When `attn send` runs, it walks its own process ancestor chain (via `/proc`) and matches against the configured names. The first match (closest ancestor, skipping self) populates the `{{.Process}}` template variable.

```toml
[processes]
code = "VS Code"
warp = "Warp"
alacritty = "Alacritty"
kitty = "kitty"
```

Use `attn proctree` to discover the comm names in your ancestor chain, or `attn proctree -i` to browse and assign labels interactively:

```bash
$ attn proctree
PID       NAME             LABEL
48291     attn
48290     bash
48115     node
47903     code             VS Code
1         systemd
```

The comm name is the `Name:` field from `/proc/<pid>/status` — typically the executable basename, truncated to 15 characters.

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
