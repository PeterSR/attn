package channel

import (
	"testing"
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
