package relay

import "github.com/petersr/attn/internal/notification"

// Message is the wire format for the relay protocol (JSON-lines).
type Message struct {
	Version int                        `json:"v"`
	Type    string                     `json:"type"` // "notify", "ping"
	Notify  *notification.Notification `json:"notify,omitempty"`
}

// Response is sent back from the server to the client.
type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
