package model

import "time"

// KafkaEvent is the message payload published to the webhook-events topic.
// Format produced by webhook-event-generator's PayloadGeneratorService.
type KafkaEvent struct {
	EventName  string      `json:"event_name"`
	EventTime  time.Time   `json:"event_time"`
	WebhookID  string      `json:"webhook_id"`
	Subscriber *Subscriber `json:"subscriber"`
	Segment    *Segment    `json:"segment,omitempty"`
}

type Subscriber struct {
	ID             string            `json:"id"`
	Email          string            `json:"email"`
	Status         string            `json:"status,omitempty"`
	Source         string            `json:"source,omitempty"`
	FirstName      string            `json:"first_name,omitempty"`
	LastName       string            `json:"last_name,omitempty"`
	Segments       []interface{}     `json:"segments,omitempty"`
	CustomFields   map[string]string `json:"custom_fields,omitempty"`
	OptinIP        string            `json:"optin_ip,omitempty"`
	OptinTimestamp string            `json:"optin_timestamp,omitempty"`
	CreatedAt      string            `json:"created_at,omitempty"`
}

type Segment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Delivery status codes for webhook_events.status
const (
	StatusPending           = 0
	StatusSuccess           = 1
	StatusPendingRetry      = 2
	StatusFailedPermanently = 3
)
