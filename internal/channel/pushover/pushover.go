package pushover

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/petersr/attn/internal/notification"
)

const apiURL = "https://api.pushover.net/1/messages.json"

// Channel sends notifications via the Pushover API.
type Channel struct {
	token   string
	userKey string
}

func New(token, userKey string) *Channel {
	return &Channel{token: token, userKey: userKey}
}

func (c *Channel) Name() string { return "pushover" }

func (c *Channel) Send(ctx context.Context, n notification.Notification) error {
	priority := "0"
	switch n.Urgency {
	case notification.UrgencyLow:
		priority = "-1"
	case notification.UrgencyCritical:
		priority = "1"
	}

	form := url.Values{
		"token":    {c.token},
		"user":     {c.userKey},
		"title":    {n.Title},
		"message":  {n.Body},
		"priority": {priority},
	}

	resp, err := http.PostForm(apiURL, form)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("pushover: HTTP %d", resp.StatusCode)
	}
	return nil
}
