package marker

import (
	"testing"

	"github.com/petersr/attn/internal/proctree"
)

func TestEvaluateSuppressBeatsForce(t *testing.T) {
	in := EvalInputs{
		SuppressEnv: []string{"DND"},
		ForceEnv:    []string{"DND"}, // both reference same var; suppress wins
		Env:         envMap(map[string]bool{"DND": true}),
	}
	got := Evaluate(in)
	if !got.ForceSuppress {
		t.Error("ForceSuppress = false, want true")
	}
	if got.ForceFire {
		t.Error("ForceFire = true, want false (suppress should win)")
	}
	if got.SuppressVar != "DND" {
		t.Errorf("SuppressVar = %q, want DND", got.SuppressVar)
	}
}

func TestEvaluateForceBeatsWalk(t *testing.T) {
	in := EvalInputs{
		Chain: []proctree.ProcessInfo{
			{PID: 100, Name: "attn"},
			{PID: 99, Name: "node"},
		},
		Markers:  []Marker{{Name: "node", Type: TypeDelegate}},
		ForceEnv: []string{"ATTN_FORCE"},
		Env:      envMap(map[string]bool{"ATTN_FORCE": true}),
	}
	got := Evaluate(in)
	if !got.ForceFire {
		t.Error("ForceFire = false, want true")
	}
	if got.Walk.Verdict != VerdictFallthrough {
		t.Errorf("Walk.Verdict = %v, want VerdictFallthrough (force should short-circuit walk)", got.Walk.Verdict)
	}
}

func TestEvaluateWalkWhenNoGlobals(t *testing.T) {
	in := EvalInputs{
		Chain: []proctree.ProcessInfo{
			{PID: 100, Name: "attn"},
			{PID: 99, Name: "node"},
		},
		Markers: []Marker{{Name: "node", Type: TypeDelegate}},
		Env:     envMap(nil),
	}
	got := Evaluate(in)
	if got.ForceSuppress || got.ForceFire {
		t.Error("globals set when none configured")
	}
	if got.Walk.Verdict != VerdictSuppress {
		t.Errorf("Walk.Verdict = %v, want VerdictSuppress", got.Walk.Verdict)
	}
}

func TestEvaluateSkipWalkGlobalsOnly(t *testing.T) {
	in := EvalInputs{
		Chain: []proctree.ProcessInfo{
			{PID: 100, Name: "attn"},
			{PID: 99, Name: "node"},
		},
		Markers:     []Marker{{Name: "node", Type: TypeDelegate}},
		SuppressEnv: []string{"DND"},
		Env:         envMap(map[string]bool{"DND": true}),
		SkipWalk:    true,
	}
	got := Evaluate(in)
	if !got.ForceSuppress {
		t.Error("ForceSuppress = false, want true (globals must apply with SkipWalk)")
	}
}

func TestEvaluateSkipWalkNoGlobals(t *testing.T) {
	in := EvalInputs{
		Chain: []proctree.ProcessInfo{
			{PID: 100, Name: "attn"},
			{PID: 99, Name: "node"},
		},
		Markers:  []Marker{{Name: "node", Type: TypeDelegate}},
		Env:      envMap(nil),
		SkipWalk: true,
	}
	got := Evaluate(in)
	if got.ForceSuppress || got.ForceFire {
		t.Error("globals set when none triggered")
	}
	if got.Walk.Verdict != VerdictFallthrough {
		t.Errorf("Walk.Verdict = %v, want VerdictFallthrough (SkipWalk should not run walker)", got.Walk.Verdict)
	}
}

func TestEvaluateEmptyEverythingFallthrough(t *testing.T) {
	got := Evaluate(EvalInputs{Env: envMap(nil)})
	if got.ForceSuppress || got.ForceFire || got.Walk.Verdict != VerdictFallthrough {
		t.Errorf("expected zero Evaluation, got %+v", got)
	}
}
