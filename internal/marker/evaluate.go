package marker

import "github.com/petersr/attn/internal/proctree"

// EvalInputs bundles everything Evaluate needs. SkipWalk is set to true on
// the relay path (hops > 0) where the local ancestor chain has no bearing
// on the originating notification.
type EvalInputs struct {
	Chain       []proctree.ProcessInfo
	Markers     []Marker
	SuppressEnv []string
	ForceEnv    []string
	Env         EnvLookup
	CmdlineFor  CmdlineFor
	SkipWalk    bool
}

// Evaluation is the result of running the precedence ladder.
type Evaluation struct {
	ForceSuppress bool
	ForceFire     bool
	// SuppressVar / ForceVar name the env var that triggered the global
	// short-circuit (if any), so callers can surface it in verbose output.
	SuppressVar string
	ForceVar    string
	// Walk is the marker walker's result; zero-value if SkipWalk is true
	// or if a global short-circuit fired.
	Walk Result
}

// Evaluate applies the precedence ladder:
//  1. Any name in SuppressEnv set → ForceSuppress.
//  2. Any name in ForceEnv set → ForceFire.
//  3. SkipWalk → return both flags false.
//  4. Otherwise run Walk.
func Evaluate(in EvalInputs) Evaluation {
	if name, ok := firstSet(in.SuppressEnv, in.Env); ok {
		return Evaluation{ForceSuppress: true, SuppressVar: name}
	}
	if name, ok := firstSet(in.ForceEnv, in.Env); ok {
		return Evaluation{ForceFire: true, ForceVar: name}
	}
	if in.SkipWalk {
		return Evaluation{}
	}
	return Evaluation{Walk: Walk(in.Chain, in.Markers, in.Env, in.CmdlineFor)}
}

// firstSet returns the first env var name from names that is set, or "" / false.
func firstSet(names []string, env EnvLookup) (string, bool) {
	if len(names) == 0 || env == nil {
		return "", false
	}
	for _, n := range names {
		if env(n) {
			return n, true
		}
	}
	return "", false
}
