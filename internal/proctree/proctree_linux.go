//go:build linux

package proctree

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ProcessInfo holds a PID and its comm name from /proc/<pid>/status.
type ProcessInfo struct {
	PID  int
	Name string
}

// Ancestors returns the chain of PIDs from pid up to PID 1 (inclusive).
// Returns nil if the process tree cannot be read.
func Ancestors(pid int) []int {
	var chain []int
	seen := make(map[int]bool)
	for pid >= 1 {
		if seen[pid] {
			break
		}
		name, ppid, err := readStatus(pid)
		_ = name
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

// AncestorsNamed returns the ancestor chain like Ancestors, but includes
// the process name from /proc/<pid>/status for each entry.
func AncestorsNamed(pid int) []ProcessInfo {
	var chain []ProcessInfo
	seen := make(map[int]bool)
	for pid >= 1 {
		if seen[pid] {
			break
		}
		name, ppid, err := readStatus(pid)
		if err != nil {
			break
		}
		seen[pid] = true
		chain = append(chain, ProcessInfo{PID: pid, Name: name})
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

// MatchKnown walks the process chain (skipping the first entry, which is self)
// and returns the friendly label of the first ancestor whose Name matches a key
// in known. Returns "" if no match is found.
func MatchKnown(chain []ProcessInfo, known map[string]string) string {
	for i, p := range chain {
		if i == 0 {
			continue // Skip self.
		}
		if label, ok := known[p.Name]; ok {
			return label
		}
	}
	return ""
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

// readStatus reads /proc/<pid>/status and returns the process name and parent PID.
func readStatus(pid int) (string, int, error) {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/status")
	if err != nil {
		return "", 0, err
	}
	var name string
	ppid := -1
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Name:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				name = fields[1]
			}
		}
		if strings.HasPrefix(line, "PPid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				ppid, err = strconv.Atoi(fields[1])
				if err != nil {
					return "", 0, err
				}
			}
		}
	}
	if ppid < 0 {
		return "", 0, fmt.Errorf("PPid not found for pid %d", pid)
	}
	return name, ppid, nil
}
