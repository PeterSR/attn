//go:build linux

package proctree

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Ancestors returns the chain of PIDs from pid up to PID 1 (inclusive).
// Returns nil if the process tree cannot be read.
func Ancestors(pid int) []int {
	var chain []int
	seen := make(map[int]bool)
	for pid >= 1 {
		if seen[pid] {
			break
		}
		ppid, err := parentPID(pid)
		if err != nil {
			break // Process doesn't exist or is not readable.
		}
		seen[pid] = true
		chain = append(chain, pid)
		if pid == 1 {
			break
		}
		if ppid <= 0 {
			break
		}
		pid = ppid
	}
	return chain
}

// IsAncestor returns true if possibleAncestor is a direct ancestor
// of pid (i.e., possibleAncestor appears in pid's ancestor chain).
func IsAncestor(pid, possibleAncestor int) bool {
	for _, a := range Ancestors(pid) {
		if a == possibleAncestor {
			return true
		}
	}
	return false
}

func parentPID(pid int) (int, error) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/status")
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PPid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return strconv.Atoi(fields[1])
			}
		}
	}
	return 0, fmt.Errorf("PPid not found for pid %d", pid)
}
