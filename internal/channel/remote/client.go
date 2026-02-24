package remote

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/petersr/attn/internal/notification"
	"github.com/petersr/attn/internal/relay"
)

// Channel sends notifications through a Unix socket relay.
type Channel struct {
	socketPath string
	hops       int
}

func New(socketPath string, hops int) *Channel {
	return &Channel{socketPath: socketPath, hops: hops}
}

func (c *Channel) Name() string { return "remote" }

func (c *Channel) Send(ctx context.Context, n notification.Notification) error {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("connect to relay socket %s: %w", c.socketPath, err)
	}
	defer conn.Close()

	msg := relay.Message{
		Version: 1,
		Type:    "notify",
		Notify:  &n,
		Hops:    c.hops + 1,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	data = append(data, '\n')

	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("write to relay: %w", err)
	}

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return fmt.Errorf("no response from relay")
	}

	var resp relay.Response
	if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
		return fmt.Errorf("invalid relay response: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("relay error: %s", resp.Error)
	}

	return nil
}
