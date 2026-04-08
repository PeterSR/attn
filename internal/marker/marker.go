// Package marker implements proctree marker rules. A marker attaches a
// suppression policy to an ancestor process: when an ancestor matches a
// marker by name (and any optional distinguishers), its Type controls how
// attn should treat notifications spawned beneath it. The walker is a pure
// function over a process chain so it can be tested without touching /proc.
package marker

import (
	"fmt"
	"strings"

	"github.com/petersr/attn/internal/proctree"
)

// Type is the action a matched marker takes.
type Type string

const (
	// TypeFocusCheck means: this ancestor is a UI surface — suppress only
	// when its window is the focused one. The walker only reports the
	// verdict; the actual focus comparison is done by the caller via
	// the existing in-process-tree check.
	TypeFocusCheck Type = "focus_check"

	// TypeDelegate means: another notifier owns this subtree, so attn
	// should stay silent for active channels.
	TypeDelegate Type = "delegate"
)

// Valid reports whether t is a known marker type.
func (t Type) Valid() bool {
	return t == TypeFocusCheck || t == TypeDelegate
}

// Marker is a single rule. Name is required; the other fields are optional
// distinguishers that are AND-composed when present.
type Marker struct {
	// Name is the exact comm name to match against ProcessInfo.Name.
	Name string
	// Type is the action to take on match.
	Type Type
	// Label optionally populates {{.Process}} in templates.
	Label string
	// MatchEnv lists env var names that must all be set on attn's own
	// process for the marker to match. Presence-only check (the values
	// are not consulted).
	MatchEnv []string
	// CmdlineContains is a substring that must appear in the matched
	// ancestor's /proc/<pid>/cmdline for the marker to match.
	CmdlineContains string
}

// Verdict is the outcome of a walker run.
type Verdict int

const (
	// VerdictFallthrough means no marker matched; the caller should use
	// its existing logic.
	VerdictFallthrough Verdict = iota
	// VerdictSuppress means a delegate marker matched; suppress active
	// channels.
	VerdictSuppress
	// VerdictFocusCheck means a focus_check marker matched; the caller
	// should consult its existing focus comparison.
	VerdictFocusCheck
)

// Result is what Walk returns.
type Result struct {
	Verdict     Verdict
	MatchedPID  int
	MatchedName string
	Label       string
	// Reason is a human-readable summary used for verbose output.
	Reason string
}

// EnvLookup reports whether the named env var is set on attn's own process.
type EnvLookup func(name string) bool

// CmdlineFor returns the joined cmdline for pid, or "" on error.
type CmdlineFor func(pid int) string

// Walk evaluates markers against an ancestor chain and returns the first
// matching marker's verdict. The first chain entry (self) is skipped to
// mirror MatchKnown's semantics. Markers are evaluated in declaration order
// for each ancestor, with the first match winning. The walker invokes
// cmdlineFor only when a marker actually has a CmdlineContains distinguisher.
func Walk(chain []proctree.ProcessInfo, markers []Marker, env EnvLookup, cmdlineFor CmdlineFor) Result {
	if len(markers) == 0 || len(chain) <= 1 {
		return Result{Verdict: VerdictFallthrough}
	}

	for i, p := range chain {
		if i == 0 {
			continue // Skip self.
		}
		for _, m := range markers {
			if m.Name != p.Name {
				continue
			}
			if !envMatches(m.MatchEnv, env) {
				continue
			}
			if m.CmdlineContains != "" {
				if !strings.Contains(cmdlineFor(p.PID), m.CmdlineContains) {
					continue
				}
			}
			return Result{
				Verdict:     verdictFor(m.Type),
				MatchedPID:  p.PID,
				MatchedName: p.Name,
				Label:       m.Label,
				Reason:      reasonFor(m, p),
			}
		}
	}
	return Result{Verdict: VerdictFallthrough}
}

func envMatches(names []string, env EnvLookup) bool {
	if len(names) == 0 {
		return true
	}
	if env == nil {
		return false
	}
	for _, n := range names {
		if !env(n) {
			return false
		}
	}
	return true
}

func verdictFor(t Type) Verdict {
	switch t {
	case TypeDelegate:
		return VerdictSuppress
	case TypeFocusCheck:
		return VerdictFocusCheck
	}
	return VerdictFallthrough
}

func reasonFor(m Marker, p proctree.ProcessInfo) string {
	parts := []string{string(m.Type), fmt.Sprintf("%s(pid=%d)", p.Name, p.PID)}
	if len(m.MatchEnv) > 0 {
		parts = append(parts, "env="+strings.Join(m.MatchEnv, ","))
	}
	if m.CmdlineContains != "" {
		parts = append(parts, "cmdline~"+m.CmdlineContains)
	}
	return strings.Join(parts, " ")
}
