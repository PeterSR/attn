//go:build linux

package screen

import (
	internaldbus "github.com/petersr/attn/internal/dbus"
)

// Get returns the current screen state by querying D-Bus screensaver
// interfaces. Tries GNOME first, then freedesktop. Returns StateUnknown
// if neither is available.
func Get() State {
	conn, err := internaldbus.SessionBus()
	if err != nil {
		return StateUnknown
	}
	defer conn.Close()

	// Try GNOME ScreenSaver.
	obj := conn.Object("org.gnome.ScreenSaver", "/org/gnome/ScreenSaver")
	var active bool
	if err := obj.Call("org.gnome.ScreenSaver.GetActive", 0).Store(&active); err == nil {
		if active {
			return StateIdle
		}
		return StateActive
	}

	// Fallback: freedesktop ScreenSaver (KDE, etc.).
	obj = conn.Object("org.freedesktop.ScreenSaver", "/org/freedesktop/ScreenSaver")
	if err := obj.Call("org.freedesktop.ScreenSaver.GetActive", 0).Store(&active); err == nil {
		if active {
			return StateIdle
		}
		return StateActive
	}

	return StateUnknown
}
