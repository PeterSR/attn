# Screen Idle Detection

attn can detect whether the screen is locked or turned off, and route notifications accordingly. This enables patterns like "desktop notification when I'm at my desk, push to phone when I'm away."

## How it works

On Linux, attn queries the D-Bus screensaver interface to determine screen state:

1. **GNOME**: `org.gnome.ScreenSaver.GetActive()` — returns `true` when the screen is locked or the screensaver is active
2. **Freedesktop (KDE, etc.)**: `org.freedesktop.ScreenSaver.GetActive()` — fallback for non-GNOME desktops

GNOME is tried first, then freedesktop. Both use the existing D-Bus session bus helper (`internal/dbus/session.go`), which includes fallback logic for non-interactive contexts.

## Screen states

| State | Meaning |
|-------|---------|
| `StateActive` | Screen is on and unlocked |
| `StateIdle` | Screen is off or locked |
| `StateUnknown` | Cannot determine (D-Bus unavailable, unsupported DE, SSH session) |

## Channel behavior by state

| `when` value | Screen active | Screen idle | State unknown |
|-------------|--------------|-------------|---------------|
| `active` | Fires (unless focused) | Does not fire | **Fires** (fail-open) |
| `idle` | Does not fire | Fires | **Does not fire** (fail-closed) |
| `always` | Fires | Fires | Fires |
| `never` | Does not fire | Does not fire | Does not fire |

### Design rationale

- **`active` fails open**: If we can't tell whether the screen is locked, assume it's active and send the desktop notification. The worst case is an unnecessary notification.
- **`idle` fails closed**: If we can't tell whether the screen is locked, don't send push notifications to the phone. The worst case is a missed push that would have been delivered on the desktop anyway.

## Typical setup

```toml
[desktop]
when = "active"  # default

[ntfy]
when = "idle"
topic = "my-notifications"
```

This ensures exactly one notification path fires:
- Screen on → desktop notification
- Screen off → ntfy push to phone

## Platform support

| Platform | Detection | Method |
|----------|-----------|--------|
| Linux (GNOME) | Supported | D-Bus `org.gnome.ScreenSaver` |
| Linux (KDE) | Supported | D-Bus `org.freedesktop.ScreenSaver` |
| Linux (other) | May work | Falls back to freedesktop interface |
| macOS | Not yet | Returns `StateUnknown` |
| Windows | Not yet | Returns `StateUnknown` |

## Implementation

| Component | File |
|-----------|------|
| State type | `internal/screen/screen.go` |
| Linux detection | `internal/screen/screen_linux.go` |
| D-Bus session bus | `internal/dbus/session.go` |
| Dispatch filtering | `internal/channel/channel.go` — `ShouldFire()` |
