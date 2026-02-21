//go:build linux

package focus

import (
	"encoding/json"

	internaldbus "github.com/petersr/attn/internal/dbus"
)

// focusedWindowWayland queries the GNOME Shell "Focused Window D-Bus" extension.
// Returns the wm_class of the focused window, or "" on failure.
func focusedWindowWayland() string {
	conn, err := internaldbus.SessionBus()
	if err != nil {
		return ""
	}
	defer conn.Close()

	obj := conn.Object(
		"org.gnome.Shell",
		"/org/gnome/shell/extensions/FocusedWindow",
	)

	var result string
	err = obj.Call(
		"org.gnome.shell.extensions.FocusedWindow.Get",
		0,
	).Store(&result)
	if err != nil {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return ""
	}

	if cls, ok := data["wm_class"].(string); ok && cls != "" {
		return cls
	}
	if cls, ok := data["wm_class_instance"].(string); ok && cls != "" {
		return cls
	}

	return ""
}
