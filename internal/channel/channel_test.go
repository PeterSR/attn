package channel

import (
	"testing"

	"github.com/petersr/attn/internal/marker"
)

func TestShouldFire(t *testing.T) {
	tests := []struct {
		name  string
		when  When
		state ScreenState
		want  bool
	}{
		// Never.
		{"never/active screen", WhenNever, ScreenState{DetectionOK: true}, false},
		{"never/idle screen", WhenNever, ScreenState{DetectionOK: true, Idle: true}, false},
		{"never/unknown", WhenNever, ScreenState{}, false},

		// Always.
		{"always/active screen", WhenAlways, ScreenState{DetectionOK: true}, true},
		{"always/idle screen", WhenAlways, ScreenState{DetectionOK: true, Idle: true}, true},
		{"always/unknown", WhenAlways, ScreenState{}, true},

		// Active.
		{"active/screen active, not focused", WhenActive, ScreenState{DetectionOK: true}, true},
		{"active/screen active, focused", WhenActive, ScreenState{DetectionOK: true, InProcessTree: true}, false},
		{"active/screen idle", WhenActive, ScreenState{DetectionOK: true, Idle: true}, false},
		{"active/unknown (fail-open)", WhenActive, ScreenState{}, true},

		// Idle.
		{"idle/screen idle", WhenIdle, ScreenState{DetectionOK: true, Idle: true}, true},
		{"idle/screen active", WhenIdle, ScreenState{DetectionOK: true}, false},
		{"idle/unknown (fail-closed)", WhenIdle, ScreenState{}, false},

		// Invalid.
		{"invalid when value", When("bogus"), ScreenState{DetectionOK: true}, false},

		// ForceSuppress short-circuits.
		{"force_suppress/active_skips", WhenActive, ScreenState{DetectionOK: true, ForceSuppress: true}, false},
		{"force_suppress/idle_skips", WhenIdle, ScreenState{DetectionOK: true, Idle: true, ForceSuppress: true}, false},
		{"force_suppress/always_skips", WhenAlways, ScreenState{ForceSuppress: true}, false},
		{"force_suppress/never_still_false", WhenNever, ScreenState{ForceSuppress: true}, false},

		// ForceFire short-circuits.
		{"force_fire/active_fires_when_focused", WhenActive, ScreenState{DetectionOK: true, InProcessTree: true, ForceFire: true}, true},
		{"force_fire/idle_fires_when_active", WhenIdle, ScreenState{DetectionOK: true, ForceFire: true}, true},
		{"force_fire/never_still_false", WhenNever, ScreenState{ForceFire: true}, false},

		// Marker delegate suppresses active.
		{"marker_suppress/delegate_skips_active", WhenActive, ScreenState{DetectionOK: true, MarkerVerdict: marker.VerdictSuppress}, false},

		// Marker focus_check defers to InProcessTree.
		{"marker_focus_check/inprocesstree_true_skips", WhenActive, ScreenState{DetectionOK: true, MarkerVerdict: marker.VerdictFocusCheck, InProcessTree: true}, false},
		{"marker_focus_check/inprocesstree_false_fires", WhenActive, ScreenState{DetectionOK: true, MarkerVerdict: marker.VerdictFocusCheck, InProcessTree: false}, true},

		// Marker fallthrough preserves existing logic.
		{"marker_fallthrough/preserves_active_inprocesstree_skip", WhenActive, ScreenState{DetectionOK: true, InProcessTree: true, MarkerVerdict: marker.VerdictFallthrough}, false},
		{"marker_fallthrough/preserves_active_fire", WhenActive, ScreenState{DetectionOK: true, MarkerVerdict: marker.VerdictFallthrough}, true},

		// Precedence checks.
		{"precedence/suppress_beats_marker_focus_check", WhenActive, ScreenState{DetectionOK: true, ForceSuppress: true, MarkerVerdict: marker.VerdictFocusCheck, InProcessTree: false}, false},
		{"precedence/suppress_beats_force", WhenActive, ScreenState{DetectionOK: true, ForceSuppress: true, ForceFire: true}, false},
		{"precedence/force_beats_marker_suppress", WhenActive, ScreenState{DetectionOK: true, ForceFire: true, MarkerVerdict: marker.VerdictSuppress}, true},

		// Markers do not affect Idle.
		{"marker_does_not_affect_idle", WhenIdle, ScreenState{DetectionOK: true, Idle: true, MarkerVerdict: marker.VerdictSuppress}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldFire(tt.when, tt.state)
			if got != tt.want {
				t.Errorf("ShouldFire(%q, %+v) = %v, want %v", tt.when, tt.state, got, tt.want)
			}
		})
	}
}
