// Package dto holds request/response shapes for the payment HTTP port.
package dto

// IntentInput is consumed by Gateway.CreateIntent.
type IntentInput struct {
	OrderID     string `json:"order_id"`
	AmountMinor int64  `json:"amount_minor"`
	Currency    string `json:"currency"`
}

// Intent is the payment-provider's pending charge handle.
type Intent struct {
	IntentID string `json:"intent_id"`
	URL      string `json:"url"`
}

// WebhookEvent is the provider-agnostic shape produced by Gateway.VerifyWebhook.
type WebhookEvent struct {
	Provider string
	EventID  string
	IntentID string
	Type     string // e.g. "payment.settled", "payment.failed"
	Raw      []byte
}
