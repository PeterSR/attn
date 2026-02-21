package bell

import (
	"context"
	"fmt"
	"os"

	"github.com/petersr/attn/internal/notification"
)

// Channel sends a terminal bell character to stderr.
type Channel struct{}

func New() *Channel {
	return &Channel{}
}

func (c *Channel) Name() string { return "bell" }

func (c *Channel) Send(_ context.Context, n notification.Notification) error {
	if _, err := fmt.Fprint(os.Stderr, "\a"); err != nil {
		return err
	}
	return nil
}
