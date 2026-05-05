// Package dto holds request/response shapes for the inventory port.
package dto

import "time"

// Item is a single (showtime, ticket-type, qty) line in a hold request.
type Item struct {
	ShowtimeID   string `json:"showtime_id"    validate:"required,len=26"`
	TicketTypeID string `json:"ticket_type_id" validate:"required,len=26"`
	Quantity     int    `json:"quantity"       validate:"required,min=1,max=20"`
}

// HoldInput is consumed by Inventory.Hold.
type HoldInput struct {
	UserID string        `json:"user_id"`
	Items  []Item        `json:"items"   validate:"required,min=1,max=20,dive"`
	TTL    time.Duration `json:"-"`
}

// Hold is the persisted hold record returned to callers.
type Hold struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Status    string    `json:"status"`
	Items     []Item    `json:"items"`
	ExpiresAt time.Time `json:"expires_at"`
}

// QuotaSpec is consumed by Inventory.Warm to seed Redis at showtime go-live.
type QuotaSpec struct {
	TicketTypeID string `json:"ticket_type_id"`
	Total        int    `json:"total"`
}
