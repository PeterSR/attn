//go:build linux

package focus

// FocusedWindow returns the WM class of the currently focused window.
// Tries Wayland/GNOME D-Bus first, then X11. Returns "" on failure.
func FocusedWindow() string {
	if cls := focusedWindowWayland(); cls != "" {
		return cls
	}
	if cls := focusedWindowX11(); cls != "" {
		return cls
	}
	return ""
}
