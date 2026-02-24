//go:build linux

package focus

import "os"

// FocusedWindowInfo returns detailed info about the focused window.
// Tries Wayland/GNOME D-Bus first, then X11. Returns zero-value on failure.
//
// On Wayland sessions, X11 is not used as a fallback because XWayland only
// tracks XWayland windows — switching to a native Wayland window won't update
// the X11 active window, producing stale/incorrect results.
func FocusedWindowInfo() FocusInfo {
	wayland := os.Getenv("WAYLAND_DISPLAY") != ""

	if info := focusedWindowInfoWayland(); info.PID != 0 || info.Class != "" {
		return info
	}

	// Only fall back to X11 on a pure X11 session.
	if !wayland {
		if info := focusedWindowInfoX11(); info.PID != 0 || info.Class != "" {
			return info
		}
	}

	return FocusInfo{}
}

// FocusedWindow returns the WM class of the currently focused window.
// Tries Wayland/GNOME D-Bus first, then X11. Returns "" on failure.
func FocusedWindow() string {
	return FocusedWindowInfo().Class
}
