//go:build !linux && !darwin

package screen

// Get returns StateUnknown on unsupported platforms.
func Get() State {
	return StateUnknown
}
