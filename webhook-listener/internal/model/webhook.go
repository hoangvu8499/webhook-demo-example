package model

// WebhookConfig is loaded from the webhooks table and cached in Redis by webhook_id.
type WebhookConfig struct {
	ID        string   `json:"id"`
	AccountID string   `json:"account_id"`
	PostURL   string   `json:"post_url"`
	Events    []string `json:"events"`
	Status    int      `json:"status"`
}

// RetryRecord holds a row from webhook_events that is ready for re-delivery.
type RetryRecord struct {
	ID           string
	WebhookID    string
	AccountID    string
	PostURL      string
	EventName    string
	Payload      string // raw JSON as stored in DB
	AttemptCount int
	MaxAttempts  int
}
