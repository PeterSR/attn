package relay

import "github.com/petersr/attn/internal/notification"

// MaxHops is the maximum number of relay hops allowed to prevent infinite loops.
const MaxHops = 10

// Message is the wire format for the relay protocol (JSON-lines).
type Message struct {
	Version int                        `json:"v"`
	Type    string                     `json:"type"` // "notify", "ping"
	Notify  *notification.Notification `json:"notify,omitempty"`
	Hops    int                        `json:"hops,omitempty"`
}

// Response is sent back from the server to the client.
type Response struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
