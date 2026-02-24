// EXPERIMENTAL: macOS support is untested. Contributions welcome.

//go:build darwin

package desktop

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/petersr/attn/internal/notification"
)

// Channel sends desktop notifications via osascript on macOS.
type Channel struct{}

func New() *Channel {
	return &Channel{}
}

func (c *Channel) Name() string { return "desktop" }

func (c *Channel) Send(ctx context.Context, n notification.Notification) error {
	// Escape double quotes for AppleScript strings.
	escBody := strings.ReplaceAll(n.Body, `"`, `\"`)
	escTitle := strings.ReplaceAll(n.Title, `"`, `\"`)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, escBody, escTitle)

	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("osascript: %w: %s", err, string(out))
	}
	return nil
}
