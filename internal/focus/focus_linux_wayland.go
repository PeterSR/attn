//go:build linux

package focus

import (
	"encoding/json"

	internaldbus "github.com/petersr/attn/internal/dbus"
)

// focusedWindowInfoWayland queries the GNOME Shell "Focused Window D-Bus" extension.
// Returns the wm_class and PID of the focused window. Returns zero-value on failure.
func focusedWindowInfoWayland() FocusInfo {
	conn, err := internaldbus.SessionBus()
	if err != nil {
		return FocusInfo{}
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
		return FocusInfo{}
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		return FocusInfo{}
	}

	var info FocusInfo

	if cls, ok := data["wm_class"].(string); ok && cls != "" {
		info.Class = cls
	} else if cls, ok := data["wm_class_instance"].(string); ok && cls != "" {
		info.Class = cls
	}

	if pid, ok := data["pid"].(float64); ok && pid > 0 {
		info.PID = int(pid)
	}

	return info
}
