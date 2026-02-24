# Focus Detection

When a channel is set to `when = "active"`, attn automatically detects whether the user is looking at the terminal that triggered the notification. If they are, the notification is suppressed.

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

## Implementation

| Component | File |
|-----------|------|
| Process tree walking | `internal/proctree/proctree_linux.go` |
| X11 PID retrieval | `internal/focus/focus_linux_x11.go` — `ewmh.WmPidGet()` |
| Wayland PID retrieval | `internal/focus/focus_linux_wayland.go` — JSON `pid` field |
| Ancestor check | `internal/proctree/proctree_linux.go` — `IsAncestor()` |
| Integration | `internal/focus/focus.go` — `IsInProcessTree()` |
