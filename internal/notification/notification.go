package notification

// Urgency represents the notification urgency level.
type Urgency string

const (
	UrgencyLow      Urgency = "low"
	UrgencyNormal   Urgency = "normal"
	UrgencyCritical Urgency = "critical"
)

// Notification is the core message type, also used as the relay wire format.
type Notification struct {
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	Urgency   Urgency `json:"urgency"`
	TimeoutMS int     `json:"timeout_ms"`
	Context   string  `json:"context,omitempty"`
}

// FormatBody returns the body with context prefix if set.
func (n Notification) FormatBody() string {
	if n.Context == "" {
		return n.Body
	}
	return "[" + n.Context + "] " + n.Body
}
