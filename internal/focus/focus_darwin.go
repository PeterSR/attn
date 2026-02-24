// EXPERIMENTAL: macOS support is untested. Contributions welcome.

//go:build darwin

package focus

import (
	"os/exec"
	"strings"
)

// FocusedWindowInfo returns the name of the frontmost application on macOS.
// PID is not available via osascript, so only Class is set.
func FocusedWindowInfo() FocusInfo {
	return FocusInfo{Class: FocusedWindow()}
}

// FocusedWindow returns the name of the frontmost application on macOS.
func FocusedWindow() string {
	cmd := exec.Command("osascript", "-e",
		`tell application "System Events" to get name of first process whose frontmost is true`)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
