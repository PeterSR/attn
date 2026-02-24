package ntfy

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/petersr/attn/internal/notification"
)

// Channel sends notifications via ntfy.sh (or a self-hosted instance).
type Channel struct {
	server string
	topic  string
	token  string
}

func New(server, topic, token string) *Channel {
	return &Channel{server: server, topic: topic, token: token}
}

func (c *Channel) Name() string { return "ntfy" }

func (c *Channel) Send(ctx context.Context, n notification.Notification) error {
	url := strings.TrimRight(c.server, "/") + "/" + c.topic

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(n.Body))
	if err != nil {
		return err
	}

	req.Header.Set("Title", n.Title)

	switch n.Urgency {
	case notification.UrgencyLow:
		req.Header.Set("Priority", "low")
	case notification.UrgencyCritical:
		req.Header.Set("Priority", "urgent")
	default:
		req.Header.Set("Priority", "default")
	}

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("ntfy: HTTP %d", resp.StatusCode)
	}
	return nil
}
