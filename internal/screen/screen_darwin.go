// EXPERIMENTAL: macOS support is untested. Contributions welcome.

//go:build darwin

package screen

// Get returns StateUnknown on macOS (not yet implemented).
func Get() State {
	return StateUnknown
}
