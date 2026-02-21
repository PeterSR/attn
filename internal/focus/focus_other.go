//go:build !linux && !darwin

package focus

// FocusedWindow is a no-op on unsupported platforms.
// Returns "" so focus suppression is never triggered.
func FocusedWindow() string {
	return ""
}
