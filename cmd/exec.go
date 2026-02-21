package cmd

import "os/exec"

// newCommand creates an exec.Cmd. Thin wrapper for testability.
func newCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}
