//go:build linux

package focus

import (
	"os"

	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
	"github.com/jezek/xgbutil/icccm"
)

// focusedWindowInfoX11 queries X11 for the focused window's WM_CLASS and PID.
// Only works on X11 sessions (not pure Wayland). Returns zero-value on failure.
func focusedWindowInfoX11() FocusInfo {
	if os.Getenv("WAYLAND_DISPLAY") != "" && os.Getenv("DISPLAY") == "" {
		return FocusInfo{}
	}

	xu, err := xgbutil.NewConn()
	if err != nil {
		return FocusInfo{}
	}
	defer xu.Conn().Close()

	active, err := ewmh.ActiveWindowGet(xu)
	if err != nil {
		return FocusInfo{}
	}

	var info FocusInfo

	if wmClass, err := icccm.WmClassGet(xu, active); err == nil {
		info.Class = wmClass.Class
	}

	if pid, err := ewmh.WmPidGet(xu, active); err == nil {
		info.PID = int(pid)
	}

	return info
}
