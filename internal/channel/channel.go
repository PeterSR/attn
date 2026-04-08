package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/petersr/attn/internal/focus"
	"github.com/petersr/attn/internal/marker"
	"github.com/petersr/attn/internal/notification"
	"github.com/petersr/attn/internal/screen"
)

// Channel sends a notification through a specific backend.
type Channel interface {
	Name() string
	Send(ctx context.Context, n notification.Notification) error
}

// PerChannelTimeout is the default timeout for each channel's Send.
const PerChannelTimeout = 10 * time.Second

// When controls when a channel fires relative to screen state.
type When string

const (
	WhenNever  When = "never"
	WhenActive When = "active"
	WhenIdle   When = "idle"
	WhenAlways When = "always"
)

// Entry pairs a channel with its firing condition.
type Entry struct {
	Channel Channel
	When    When
}

// ScreenState captures the current screen and focus state,
// evaluated once per dispatch to avoid redundant detection.
//
// The marker-related fields (ForceSuppress, ForceFire, MarkerVerdict, ...)
// are populated by the cmd/markers.go overlay after DetectScreenState runs;
// the channel package itself has no config knowledge.
type ScreenState struct {
	Idle          bool // Screen is off or locked.
	InProcessTree bool // Focused window shares process tree with attn.
	DetectionOK   bool // Screen state was successfully detected.

	// ForceSuppress is set when a global [suppress] env var is present.
	// Suppresses every channel regardless of When.
	ForceSuppress bool
	// ForceFire is set when a global [force] env var is present. Fires
	// every channel regardless of When (except WhenNever).
	ForceFire bool
	// MarkerVerdict is the result of walking the proctree marker rules.
	MarkerVerdict marker.Verdict
	// MarkerLabel is the matched marker's label, used for {{.Process}}.
	MarkerLabel string
	// MarkerReason is a human-readable summary for verbose output.
	MarkerReason string
}

// ShouldFire returns true if the given When condition is met by the state.
//
// Precedence (applied at the top of every When except Never):
//  1. ForceSuppress short-circuits to false.
//  2. ForceFire short-circuits to true.
//  3. For WhenActive only, marker verdicts apply: VerdictSuppress → false,
//     VerdictFocusCheck → fall through to the existing in-process-tree check,
//     VerdictFallthrough → existing logic. Markers do not affect WhenIdle
//     (idle channels are the AFK fallback and should still fire).
func ShouldFire(when When, state ScreenState) bool {
	switch when {
	case WhenNever:
		return false
	case WhenAlways:
		if state.ForceSuppress {
			return false
		}
		return true
	case WhenActive:
		if state.ForceSuppress {
			return false
		}
		if state.ForceFire {
			return true
		}
		switch state.MarkerVerdict {
		case marker.VerdictSuppress:
			return false
		case marker.VerdictFocusCheck:
			// Fall through to the existing in-process-tree check below.
		}
		if !state.DetectionOK {
			return true // Fail-open: fire if we can't detect.
		}
		if state.Idle {
			return false
		}
		if state.InProcessTree {
			return false
		}
		return true
	case WhenIdle:
		if state.ForceSuppress {
			return false
		}
		if state.ForceFire {
			return true
		}
		if !state.DetectionOK {
			return false // Fail-closed: don't fire if we can't detect.
		}
		return state.Idle
	default:
		return false
	}
}

// DetectScreenState evaluates screen and focus state once. Only performs
// detection if at least one channel entry needs it. When hops > 0, the
// notification arrived via relay and process-tree focus detection is
// skipped — the local process tree is irrelevant for remote notifications.
func DetectScreenState(entries []Entry, hops int) ScreenState {
	needsDetection := false
	for _, e := range entries {
		if e.When == WhenActive || e.When == WhenIdle {
			needsDetection = true
			break
		}
	}
	if !needsDetection {
		return ScreenState{}
	}

	screenState := screen.Get()
	state := ScreenState{
		DetectionOK: screenState != screen.StateUnknown,
		Idle:        screenState == screen.StateIdle,
	}

	// Only check process tree for locally-originated notifications (hops == 0).
	// For relayed notifications, the local process tree is irrelevant.
	if hops == 0 && !state.Idle && state.DetectionOK {
		needsProcessTree := false
		for _, e := range entries {
			if e.When == WhenActive {
				needsProcessTree = true
				break
			}
		}
		if needsProcessTree {
			state.InProcessTree = focus.IsInProcessTree()
		}
	}

	return state
}

// DispatchFiltered fires a notification to channels whose When condition
// matches the current screen state. Returns a combined error of any failures.
func DispatchFiltered(ctx context.Context, entries []Entry, state ScreenState, n notification.Notification) error {
	var channels []Channel
	for _, e := range entries {
		if ShouldFire(e.When, state) {
			channels = append(channels, e.Channel)
		}
	}
	return Dispatch(ctx, channels, n)
}

// DispatchResult records the outcome for a single channel in a verbose dispatch.
type DispatchResult struct {
	Name   string
	When   When
	Fired  bool
	Err    error
	Reason string // Non-empty when Fired is false (e.g. "when=never").
}

// DispatchFilteredVerbose is like DispatchFiltered but returns per-channel results
// for verbose output. The returned error is the same combined error as Dispatch.
func DispatchFilteredVerbose(ctx context.Context, entries []Entry, state ScreenState, n notification.Notification) ([]DispatchResult, error) {
	results := make([]DispatchResult, len(entries))

	var toFire []int // indices into entries/results
	for i, e := range entries {
		results[i].Name = e.Channel.Name()
		results[i].When = e.When
		if ShouldFire(e.When, state) {
			results[i].Fired = true
			toFire = append(toFire, i)
		} else {
			results[i].Reason = skipReason(e.When, state)
		}
	}

	if len(toFire) == 0 {
		return results, nil
	}

	type sendResult struct {
		idx int
		err error
	}
	ch := make(chan sendResult, len(toFire))
	for _, idx := range toFire {
		go func(idx int) {
			cctx, cancel := context.WithTimeout(ctx, PerChannelTimeout)
			defer cancel()
			ch <- sendResult{idx: idx, err: entries[idx].Channel.Send(cctx, n)}
		}(idx)
	}

	var errs []error
	for range toFire {
		r := <-ch
		if r.err != nil {
			results[r.idx].Err = r.err
			errs = append(errs, fmt.Errorf("%s: %w", results[r.idx].Name, r.err))
		}
	}
	return results, errors.Join(errs...)
}

// skipReason explains why ShouldFire returned false for the given condition and state.
func skipReason(when When, state ScreenState) string {
	if when == WhenNever {
		return "when=never"
	}
	if state.ForceSuppress {
		return "suppressed by env"
	}
	switch when {
	case WhenActive:
		if state.MarkerVerdict == marker.VerdictSuppress {
			if state.MarkerReason != "" {
				return "marker: " + state.MarkerReason
			}
			return "marker: suppressed"
		}
		if state.Idle {
			return "screen idle"
		}
		if state.InProcessTree {
			return "in process tree"
		}
		return "skipped"
	case WhenIdle:
		if !state.DetectionOK {
			return "detection failed (fail-closed)"
		}
		return "screen active"
	default:
		return "skipped"
	}
}

// Dispatch fires a notification to all channels concurrently.
// Returns a combined error of any failures; one channel failing
// does not prevent others from firing.
func Dispatch(ctx context.Context, channels []Channel, n notification.Notification) error {
	if len(channels) == 0 {
		return nil
	}

	type result struct {
		name string
		err  error
	}

	ch := make(chan result, len(channels))
	for _, c := range channels {
		go func(c Channel) {
			cctx, cancel := context.WithTimeout(ctx, PerChannelTimeout)
			defer cancel()
			ch <- result{name: c.Name(), err: c.Send(cctx, n)}
		}(c)
	}

	var errs []error
	for range channels {
		r := <-ch
		if r.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.name, r.err))
		}
	}
	return errors.Join(errs...)
}
