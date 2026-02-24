//go:build !linux && !darwin

package focus

// FocusedWindowInfo is a no-op on unsupported platforms.
func FocusedWindowInfo() FocusInfo {
	return FocusInfo{}
}

// FocusedWindow is a no-op on unsupported platforms.
func FocusedWindow() string {
	return ""
}
