package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/petersr/attn/internal/notification"
)

// Channel sends notifications via a generic HTTP webhook.
type Channel struct {
	url     string
	method  string
	headers map[string]string
}

func New(url, method string, headers map[string]string) *Channel {
	if method == "" {
		method = "POST"
	}
	return &Channel{url: url, method: method, headers: headers}
}

func (c *Channel) Name() string { return "webhook" }

func (c *Channel) Send(ctx context.Context, n notification.Notification) error {
	payload := map[string]string{
		"title":   n.Title,
		"body":    n.Body,
		"urgency": string(n.Urgency),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, c.method, c.url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook: HTTP %d", resp.StatusCode)
	}
	return nil
}
