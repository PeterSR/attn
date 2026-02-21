package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/petersr/attn/internal/notification"
)

// Channel sends a notification through a specific backend.
type Channel interface {
	Name() string
	Send(ctx context.Context, n notification.Notification) error
}

// PerChannelTimeout is the default timeout for each channel's Send.
const PerChannelTimeout = 10 * time.Second

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
