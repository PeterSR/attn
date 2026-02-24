//go:build !linux

package proctree

// Ancestors is a no-op on non-Linux platforms.
func Ancestors(pid int) []int {
	return nil
}

// IsAncestor is a no-op on non-Linux platforms.
func IsAncestor(pid, possibleAncestor int) bool {
	return false
}
