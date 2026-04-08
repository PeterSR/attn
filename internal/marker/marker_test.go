package marker

import (
	"testing"

	"github.com/petersr/attn/internal/proctree"
)

// envMap returns a presence-only EnvLookup backed by a static map.
func envMap(set map[string]bool) EnvLookup {
	return func(name string) bool { return set[name] }
}

// cmdlineMap returns a CmdlineFor backed by a static map.
func cmdlineMap(m map[int]string) CmdlineFor {
	return func(pid int) string { return m[pid] }
}

// chain builds a fake ancestor chain. The first entry is treated as self.
func chain(entries ...proctree.ProcessInfo) []proctree.ProcessInfo {
	return entries
}

func TestWalkNoMarkersFallthrough(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "bash"},
		proctree.ProcessInfo{PID: 1, Name: "systemd"},
	)
	got := Walk(c, nil, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough", got.Verdict)
	}
}

func TestWalkNameOnlyMatchFocusCheck(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "bash"},
		proctree.ProcessInfo{PID: 50, Name: "warp-terminal"},
		proctree.ProcessInfo{PID: 1, Name: "systemd"},
	)
	markers := []Marker{
		{Name: "warp-terminal", Type: TypeFocusCheck, Label: "Warp"},
	}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFocusCheck {
		t.Errorf("Verdict = %v, want VerdictFocusCheck", got.Verdict)
	}
	if got.MatchedPID != 50 || got.MatchedName != "warp-terminal" {
		t.Errorf("matched = %d/%s, want 50/warp-terminal", got.MatchedPID, got.MatchedName)
	}
	if got.Label != "Warp" {
		t.Errorf("Label = %q, want Warp", got.Label)
	}
	if got.Reason == "" {
		t.Error("Reason is empty")
	}
}

func TestWalkNameOnlyNoMatchFallthrough(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "bash"},
	)
	markers := []Marker{{Name: "warp-terminal", Type: TypeFocusCheck}}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough", got.Verdict)
	}
}

func TestWalkInnermostAncestorWins(t *testing.T) {
	// chain[1] node should win over chain[3] warp-terminal because the
	// walker iterates ancestors bottom-up (chain[1] is closer to self).
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
		proctree.ProcessInfo{PID: 98, Name: "bash"},
		proctree.ProcessInfo{PID: 50, Name: "warp-terminal"},
	)
	markers := []Marker{
		{Name: "warp-terminal", Type: TypeFocusCheck},
		{Name: "node", Type: TypeDelegate},
	}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictSuppress {
		t.Errorf("Verdict = %v, want VerdictSuppress", got.Verdict)
	}
	if got.MatchedPID != 99 {
		t.Errorf("MatchedPID = %d, want 99", got.MatchedPID)
	}
}

func TestWalkDeclarationOrderWithinSameAncestor(t *testing.T) {
	// Two markers can both match the same ancestor; declaration order wins.
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeFocusCheck, Label: "first"},
		{Name: "node", Type: TypeDelegate, Label: "second"},
	}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFocusCheck {
		t.Errorf("Verdict = %v, want VerdictFocusCheck (declaration order)", got.Verdict)
	}
	if got.Label != "first" {
		t.Errorf("Label = %q, want first", got.Label)
	}
}

func TestWalkMatchEnvPresent(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeDelegate, MatchEnv: []string{"WEBTERM_ID"}},
	}
	env := envMap(map[string]bool{"WEBTERM_ID": true})
	got := Walk(c, markers, env, cmdlineMap(nil))
	if got.Verdict != VerdictSuppress {
		t.Errorf("Verdict = %v, want VerdictSuppress", got.Verdict)
	}
}

func TestWalkMatchEnvMissingNoMatch(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeDelegate, MatchEnv: []string{"WEBTERM_ID"}},
	}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough", got.Verdict)
	}
}

func TestWalkMatchEnvMultiVarAND(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeDelegate, MatchEnv: []string{"A", "B"}},
	}
	// Only A set — should not match.
	env := envMap(map[string]bool{"A": true})
	got := Walk(c, markers, env, cmdlineMap(nil))
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough (only A set)", got.Verdict)
	}
}

func TestWalkMatchEnvMultiVarSatisfied(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeDelegate, MatchEnv: []string{"A", "B"}},
	}
	env := envMap(map[string]bool{"A": true, "B": true})
	got := Walk(c, markers, env, cmdlineMap(nil))
	if got.Verdict != VerdictSuppress {
		t.Errorf("Verdict = %v, want VerdictSuppress", got.Verdict)
	}
}

func TestWalkCmdlineContainsMatch(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeDelegate, CmdlineContains: "webterm"},
	}
	cmd := cmdlineMap(map[int]string{99: "node /opt/webterm/server.js"})
	got := Walk(c, markers, envMap(nil), cmd)
	if got.Verdict != VerdictSuppress {
		t.Errorf("Verdict = %v, want VerdictSuppress", got.Verdict)
	}
}

func TestWalkCmdlineContainsMiss(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeDelegate, CmdlineContains: "webterm"},
	}
	cmd := cmdlineMap(map[int]string{99: "node server.js"})
	got := Walk(c, markers, envMap(nil), cmd)
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough", got.Verdict)
	}
}

func TestWalkAllThreeDistinguishers(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{
			Name:            "node",
			Type:            TypeDelegate,
			MatchEnv:        []string{"WEBTERM_ID"},
			CmdlineContains: "webterm",
		},
	}
	env := envMap(map[string]bool{"WEBTERM_ID": true})
	cmd := cmdlineMap(map[int]string{99: "node /opt/webterm/server.js"})
	got := Walk(c, markers, env, cmd)
	if got.Verdict != VerdictSuppress {
		t.Errorf("Verdict = %v, want VerdictSuppress", got.Verdict)
	}
}

func TestWalkPartialFail(t *testing.T) {
	// Env set but cmdline doesn't match → no match.
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{
		{
			Name:            "node",
			Type:            TypeDelegate,
			MatchEnv:        []string{"WEBTERM_ID"},
			CmdlineContains: "webterm",
		},
	}
	env := envMap(map[string]bool{"WEBTERM_ID": true})
	cmd := cmdlineMap(map[int]string{99: "node server.js"})
	got := Walk(c, markers, env, cmd)
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough", got.Verdict)
	}
}

func TestWalkDelegateVerdictPropagates(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{{Name: "node", Type: TypeDelegate}}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictSuppress {
		t.Errorf("Verdict = %v, want VerdictSuppress", got.Verdict)
	}
}

func TestWalkFocusCheckVerdictPropagates(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "warp-terminal"},
	)
	markers := []Marker{{Name: "warp-terminal", Type: TypeFocusCheck}}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFocusCheck {
		t.Errorf("Verdict = %v, want VerdictFocusCheck", got.Verdict)
	}
}

func TestWalkSelfSkipped(t *testing.T) {
	// self (chain[0]) is named identically to a marker — must NOT match.
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "node"},
		proctree.ProcessInfo{PID: 99, Name: "bash"},
	)
	markers := []Marker{{Name: "node", Type: TypeDelegate}}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough (self should be skipped)", got.Verdict)
	}
}

func TestWalkPid1WalksToEnd(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "bash"},
		proctree.ProcessInfo{PID: 1, Name: "systemd"},
	)
	markers := []Marker{{Name: "systemd", Type: TypeDelegate}}
	got := Walk(c, markers, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictSuppress {
		t.Errorf("Verdict = %v, want VerdictSuppress", got.Verdict)
	}
	if got.MatchedPID != 1 {
		t.Errorf("MatchedPID = %d, want 1", got.MatchedPID)
	}
}

func TestWalkResultMetadata(t *testing.T) {
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 42, Name: "node"},
	)
	markers := []Marker{
		{Name: "node", Type: TypeDelegate, Label: "WebTerm", MatchEnv: []string{"X"}},
	}
	env := envMap(map[string]bool{"X": true})
	got := Walk(c, markers, env, cmdlineMap(nil))
	if got.MatchedPID != 42 {
		t.Errorf("MatchedPID = %d, want 42", got.MatchedPID)
	}
	if got.MatchedName != "node" {
		t.Errorf("MatchedName = %q, want node", got.MatchedName)
	}
	if got.Label != "WebTerm" {
		t.Errorf("Label = %q, want WebTerm", got.Label)
	}
	if got.Reason == "" {
		t.Error("Reason is empty")
	}
}

func TestWalkCmdlineLookupLazy(t *testing.T) {
	// A marker without CmdlineContains must NOT trigger cmdlineFor.
	calls := 0
	c := chain(
		proctree.ProcessInfo{PID: 100, Name: "attn"},
		proctree.ProcessInfo{PID: 99, Name: "node"},
	)
	markers := []Marker{{Name: "node", Type: TypeDelegate}}
	cmd := func(pid int) string {
		calls++
		return ""
	}
	Walk(c, markers, envMap(nil), cmd)
	if calls != 0 {
		t.Errorf("cmdlineFor called %d times, want 0 (no marker uses CmdlineContains)", calls)
	}
}

func TestWalkEmptyChain(t *testing.T) {
	got := Walk(nil, []Marker{{Name: "x", Type: TypeDelegate}}, envMap(nil), cmdlineMap(nil))
	if got.Verdict != VerdictFallthrough {
		t.Errorf("Verdict = %v, want VerdictFallthrough", got.Verdict)
	}
}
