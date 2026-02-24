//go:build linux

package desktop

import (
	"context"
	"fmt"

	internaldbus "github.com/petersr/attn/internal/dbus"
	"github.com/petersr/attn/internal/notification"

	godbus "github.com/godbus/dbus/v5"
)

// Channel sends desktop notifications via D-Bus org.freedesktop.Notifications.
type Channel struct{}

func New() *Channel {
	return &Channel{}
}

func (c *Channel) Name() string { return "desktop" }

func (c *Channel) Send(_ context.Context, n notification.Notification) error {
	conn, err := internaldbus.SessionBus()
	if err != nil {
		return fmt.Errorf("connect to D-Bus: %w", err)
	}
	defer conn.Close()

	obj := conn.Object(
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
	)

	urgency := urgencyToByte(n.Urgency)

	call := obj.Call(
		"org.freedesktop.Notifications.Notify",
		0,
		"attn",                // app_name
		uint32(0),             // replaces_id
		"",                    // app_icon
		n.Title,               // summary
		n.Body,                // body
		[]string{},            // actions
		map[string]godbus.Variant{ // hints
			"urgency": godbus.MakeVariant(urgency),
		},
		int32(n.TimeoutMS), // expire_timeout
	)

	if call.Err != nil {
		return fmt.Errorf("notify-send D-Bus call: %w", call.Err)
	}
	return nil
}

func urgencyToByte(u notification.Urgency) byte {
	switch u {
	case notification.UrgencyLow:
		return 0
	case notification.UrgencyCritical:
		return 2
	default:
		return 1 // normal
	}
}
