// Package dto holds request/response shapes for the order HTTP port.
package dto

import (
	"time"

	invdto "github.com/quangdangfit/goticket/internal/inventory/dto"
)

// CheckoutInput is the body of POST /orders.
type CheckoutInput struct {
	IdempotencyKey string        `json:"idempotency_key" validate:"required,min=8,max=128"`
	ShowtimeID     string        `json:"showtime_id"     validate:"required,len=26"`
	Items          []invdto.Item `json:"items"           validate:"required,min=1,max=20,dive"`
	PromoCode      string        `json:"promo_code"      validate:"omitempty,max=64"`
}

// CheckoutOutput is returned from POST /orders.
type CheckoutOutput struct {
	Order Order  `json:"order"`
	IntentURL string `json:"intent_url,omitempty"`
}

// Order is the public-facing order representation.
type Order struct {
	ID            string      `json:"id"`
	Status        string      `json:"status"`
	SubtotalMinor int64       `json:"subtotal_minor"`
	DiscountMinor int64       `json:"discount_minor"`
	TotalMinor    int64       `json:"total_minor"`
	Currency      string      `json:"currency"`
	HoldID        string      `json:"hold_id"`
	Items         []OrderItem `json:"items"`
	CreatedAt     time.Time   `json:"created_at"`
}

// OrderItem is a public line item.
type OrderItem struct {
	TicketTypeID   string `json:"ticket_type_id"`
	Quantity       int    `json:"quantity"`
	UnitPriceMinor int64  `json:"unit_price_minor"`
}
