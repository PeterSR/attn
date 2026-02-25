//go:build !linux

package proctree

// ProcessInfo holds a PID and its comm name from /proc/<pid>/status.
type ProcessInfo struct {
	PID  int
	Name string
}

// Ancestors is a no-op on non-Linux platforms.
func Ancestors(pid int) []int {
	return nil
}

// AncestorsNamed is a no-op on non-Linux platforms.
func AncestorsNamed(pid int) []ProcessInfo {
	return nil
}

// MatchKnown is a no-op on non-Linux platforms.
func MatchKnown(chain []ProcessInfo, known map[string]string) string {
	return ""
}

// IsAncestor is a no-op on non-Linux platforms.
func IsAncestor(pid, possibleAncestor int) bool {
	return false
}
