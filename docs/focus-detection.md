# Focus Detection

When a channel is set to `when = "active"`, attn decides whether to fire by walking its own process ancestor chain and checking who owns the focused window. The default rule is "suppress if the focused window is one of my ancestors." This catches the simple case (Claude in Warp → suppress when looking at Warp). For nested setups, **proctree markers** let you attach custom rules to specific ancestors. Two **global env-var lists** layer on top as a transient mute / override.

This page covers the default rule, the marker system, and the precedence between the two.

## How it works

attn uses **process-tree matching** rather than window-class regex. This is strictly better because it identifies the *specific* terminal instance, not just "any terminal of this type."

### Step 1: Walk attn's own ancestor chain

When invoked (e.g., from a Claude Code hook), attn walks `/proc/<pid>/status` up through its parent processes:

```
attn (PID 5432)
  → bash (PID 5430)        # hook subprocess
    → claude (PID 5100)    # Claude Code CLI
      → bash (PID 4800)    # user's shell
        → warp (PID 4500)  # terminal emulator
```

### Step 2: Get the focused window's PID

attn queries the display server for the currently focused window's PID:

- **X11**: Reads `_NET_WM_PID` property via EWMH on the active window (from `_NET_ACTIVE_WINDOW`)
- **Wayland (GNOME)**: Queries the GNOME Shell "Focused Window" D-Bus extension, which returns a JSON object that may include a `pid` field

### Step 3: Check if the focused PID is a direct ancestor

attn checks whether the focused window's PID appears in its own ancestor chain. If the focused window is Warp (PID 4500) and attn's chain includes PID 4500, the user is looking at the terminal that spawned attn. The notification is suppressed.

If the user is looking at Firefox (PID 7000), that PID doesn't appear in attn's ancestor chain, and the notification fires.

**Why direct ancestry, not shared ancestors?** All user processes share `systemd --user` as a common ancestor. A shared-ancestor check would incorrectly suppress notifications whenever *any* user-owned window is focused. Direct ancestry correctly answers: "is the focused window the terminal that launched me?"

## Edge cases

### Terminal multiplexers (tmux, screen)

If attn runs inside tmux, the chain goes `attn → bash → tmux-client → terminal`. The tmux client is a child of the terminal, so the ancestor chains still share the terminal PID. Focus detection works correctly.

### Wayland without the GNOME extension

The PID field depends on the GNOME Shell "Focused Window" D-Bus extension being installed. Without it, or on non-GNOME Wayland compositors, PID detection returns 0 (focus suppression is skipped, fail-open).

On Wayland sessions, X11 is **not** used as a fallback for focus detection because XWayland only tracks XWayland windows. Switching to a native Wayland window wouldn't update the X11 active window, producing stale/incorrect results.

### PID unavailable

Some applications (especially legacy X11 apps) may not set `_NET_WM_PID`. When the focused window's PID is 0, process-tree matching is skipped and the channel fires (fail-open).

### Non-Linux platforms

Process-tree walking requires `/proc`, which is Linux-specific. On macOS and other platforms, `IsInProcessTree()` always returns false, so `when = "active"` channels always fire when the screen is on (no focus suppression, but idle detection still works where available).

## Proctree markers

The default ancestor-check works for the simple case but breaks for nested setups. Example: a dev tool runs Claude inside browser-served web terminals. The Claude process still has `warp-terminal` as a GUI ancestor (the dev tool was launched from Warp), but the user actually interacts with Claude through the browser. The default rule suppresses the notification whenever Warp is focused — wrong, because the user isn't looking at Warp at all.

Markers let you attach a rule to a specific ancestor process. When the walker hits a matching ancestor, the marker's `type` decides what attn does:

| Type | Meaning |
|------|---------|
| `delegate` | Another notifier owns this subtree. Suppress active channels. (Idle channels still fire — they're your AFK fallback.) |
| `focus_check` | This ancestor is a UI surface. Suppress if its window is the focused one (i.e., fall through to the existing in-process-tree check). |

Markers are matched by `name` (exact match against `/proc/<pid>/status` `Name:`) plus optional **distinguishers** that are AND-composed:

- `match_env` — list of env var names that must all be set on attn's own process. Presence-only, values are ignored. Useful because env vars are inherited through `fork`/`exec`, so a marker can detect "I was launched from inside *this* tool" without inspecting the ancestor's `/proc/<pid>/environ`.
- `cmdline_contains` — substring that must appear in the matched ancestor's `/proc/<pid>/cmdline`. Useful for telling apart instances of a generic process like `node` or `python`.

The walker iterates ancestors bottom-up (innermost first) and within each ancestor evaluates markers in declaration order. **First match wins.** The first chain entry (attn itself) is skipped, so a marker named after attn's own comm name will never accidentally match.

### Example: delegate to a custom notifier

```toml
[[proctree.marker]]
name      = "node"
type      = "delegate"
label     = "WebTerm"
match_env = ["WEBTERM_ID"]
```

Translation: "If any ancestor is named `node` AND `WEBTERM_ID` is set in my own env, treat this as someone else's job and stay silent on active channels. Use `WebTerm` as `{{.Process}}`."

### Example: extra UI surface that should suppress when focused

```toml
[[proctree.marker]]
name = "tmux: server"
type = "focus_check"
```

Translation: "If any ancestor is `tmux: server`, treat it as a focusable UI surface — fall through to the normal focused-window check."

### Marker `label` and `[processes]`

Both `label` and the `[processes]` table populate `{{.Process}}`. **Marker label wins** when a marker matches; otherwise `[processes]` is consulted as before. The `[processes]` table is unchanged and continues to work as a pure rendering shorthand.

## Global env-var overrides

Two global lists let env-var presence flip the decision regardless of markers:

```toml
[suppress]
if_env = ["IN_MEETING", "DND"]    # any of these set → suppress every channel

[force]
if_env = ["ATTN_FORCE"]           # any of these set → fire every channel (except never)
```

These describe the host's mood — DND, in a meeting, "ignore my usual rules just this once." Unlike markers, they apply on the **relay side too** — a notification arriving from a remote machine still respects your local DND. The marker walk itself is skipped on the relay side because the local ancestor chain has nothing to do with the originating notification.

## Precedence

When `ShouldFire` evaluates a channel, the precedence is:

1. **`when = "never"`** — always false. Nothing overrides this.
2. **`[suppress]` env var set** — false. Beats everything else.
3. **`[force]` env var set** — true (unless `when = "never"`). Beats markers.
4. **Marker verdict** (active channels only):
   - `delegate` matched → false
   - `focus_check` matched → fall through to step 5
   - no marker matched → fall through to step 5
5. **Existing logic** — `WhenActive` checks idle / in-process-tree; `WhenIdle` checks idle; `WhenAlways` fires; etc.

Markers do **not** affect `when = "idle"` channels. Idle channels are the AFK fallback (push to phone when screen is locked), and silencing them based on local proctree state would defeat the purpose. If you need to mute everything globally, use `[suppress]`.

## Verbose output

`attn send -v` prints the marker decision to stderr when one applied:

```
attn: screen: idle=false inProcessTree=false detectionOK=true
attn: marker: delegate node(pid=4321) env=WEBTERM_ID
attn: desktop(when=active): skipped (marker: delegate node(pid=4321) env=WEBTERM_ID)
attn: ntfy(when=idle): skipped (screen active)
```

## Implementation

| Component | File |
|-----------|------|
| Process tree walking | `internal/proctree/proctree_linux.go` |
| X11 PID retrieval | `internal/focus/focus_linux_x11.go` — `ewmh.WmPidGet()` |
| Wayland PID retrieval | `internal/focus/focus_linux_wayland.go` — JSON `pid` field |
| Ancestor check | `internal/proctree/proctree_linux.go` — `IsAncestor()` |
| Default focus integration | `internal/focus/focus.go` — `IsInProcessTree()` |
| Marker walker (pure) | `internal/marker/marker.go` — `Walk()` |
| Precedence orchestrator | `internal/marker/evaluate.go` — `Evaluate()` |
| Runtime overlay | `cmd/markers.go` — `applyMarkerOverlay()` |
| `ShouldFire` integration | `internal/channel/channel.go` |
