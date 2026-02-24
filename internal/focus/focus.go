package focus

import (
	"os"

	"github.com/petersr/attn/internal/proctree"
)

// FocusInfo contains information about the currently focused window.
type FocusInfo struct {
	Class string // WM_CLASS or application name.
	PID   int    // Process ID of the focused window (0 if unavailable).
}

// IsInProcessTree returns true if the focused window's process is a
// direct ancestor of the current process. This indicates the user is
// looking at the terminal that spawned attn.
func IsInProcessTree() bool {
	info := FocusedWindowInfo()
	if info.PID == 0 {
		return false
	}
	return proctree.IsAncestor(os.Getpid(), info.PID)
}
