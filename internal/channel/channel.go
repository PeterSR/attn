package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/petersr/attn/internal/focus"
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
type ScreenState struct {
	Idle          bool // Screen is off or locked.
	InProcessTree bool // Focused window shares process tree with attn.
	DetectionOK   bool // Screen state was successfully detected.
}

// ShouldFire returns true if the given When condition is met by the state.
func ShouldFire(when When, state ScreenState) bool {
	switch when {
	case WhenNever:
		return false
	case WhenAlways:
		return true
	case WhenActive:
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
		if !state.DetectionOK {
			return false // Fail-closed: don't fire if we can't detect.
		}
		return state.Idle
	default:
		return false
	}
}

// DetectScreenState evaluates screen and focus state once. Only performs
// detection if at least one channel entry needs it.
func DetectScreenState(entries []Entry) ScreenState {
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

	// Only check process tree if screen is active and an "active" channel exists.
	if !state.Idle && state.DetectionOK {
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
