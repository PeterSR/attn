package cmd

import (
	"os"

	"github.com/petersr/attn/internal/channel"
	"github.com/petersr/attn/internal/config"
	"github.com/petersr/attn/internal/marker"
	"github.com/petersr/attn/internal/proctree"
)

// buildMarkers translates the on-disk marker config into runtime markers.
// Invalid types are dropped silently — config.Load() validates them, so this
// only runs against already-validated input.
func buildMarkers(cfg config.Config) []marker.Marker {
	if len(cfg.Proctree.Markers) == 0 {
		return nil
	}
	out := make([]marker.Marker, 0, len(cfg.Proctree.Markers))
	for _, m := range cfg.Proctree.Markers {
		t := marker.Type(m.Type)
		if !t.Valid() {
			continue
		}
		out = append(out, marker.Marker{
			Name:            m.Name,
			Type:            t,
			Label:           m.Label,
			MatchEnv:        m.MatchEnv,
			CmdlineContains: m.CmdlineContains,
		})
	}
	return out
}

// applyMarkerOverlay populates the marker-related fields on state from cfg
// and the current process's ancestor chain.
//
// Globals (suppress / force) apply at any hops because they describe the
// host's mood (DND, meeting). The marker walk only runs when hops == 0 —
// for relayed notifications the local chain belongs to the wrong machine.
func applyMarkerOverlay(state *channel.ScreenState, cfg config.Config, hops int) {
	env := func(name string) bool {
		_, ok := os.LookupEnv(name)
		return ok
	}

	in := marker.EvalInputs{
		Markers:     buildMarkers(cfg),
		SuppressEnv: cfg.Suppress.IfEnv,
		ForceEnv:    cfg.Force.IfEnv,
		Env:         env,
		SkipWalk:    hops != 0,
	}

	if hops == 0 && len(in.Markers) > 0 {
		in.Chain = proctree.AncestorsNamed(os.Getpid())
		in.CmdlineFor = proctree.Cmdline
	}

	ev := marker.Evaluate(in)
	state.ForceSuppress = ev.ForceSuppress
	state.ForceFire = ev.ForceFire
	state.MarkerVerdict = ev.Walk.Verdict
	state.MarkerLabel = ev.Walk.Label
	state.MarkerReason = markerOverlayReason(ev)
}

// markerOverlayReason builds a verbose-output reason string covering both
// global env shortcuts and walker matches, so cmd/send.go has a single
// field to print.
func markerOverlayReason(ev marker.Evaluation) string {
	switch {
	case ev.ForceSuppress:
		if ev.SuppressVar != "" {
			return "force_suppress (env=" + ev.SuppressVar + ")"
		}
		return "force_suppress"
	case ev.ForceFire:
		if ev.ForceVar != "" {
			return "force_fire (env=" + ev.ForceVar + ")"
		}
		return "force_fire"
	default:
		return ev.Walk.Reason
	}
}
