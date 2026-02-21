//go:build linux

package focus

import (
	"os"

	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
	"github.com/jezek/xgbutil/icccm"
)

// focusedWindowX11 queries X11 for the focused window's WM_CLASS.
// Only works on X11 sessions (not Wayland). Returns "" on failure.
func focusedWindowX11() string {
	// Skip if we're clearly on Wayland without X11.
	if os.Getenv("WAYLAND_DISPLAY") != "" && os.Getenv("DISPLAY") == "" {
		return ""
	}

	xu, err := xgbutil.NewConn()
	if err != nil {
		return ""
	}
	defer xu.Conn().Close()

	active, err := ewmh.ActiveWindowGet(xu)
	if err != nil {
		return ""
	}

	wmClass, err := icccm.WmClassGet(xu, active)
	if err != nil {
		return ""
	}

	return wmClass.Class
}
