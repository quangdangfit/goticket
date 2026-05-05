// Package model holds GORM entities for the payment bounded context.
package model

import "time"

// Payment lifecycle states.
const (
	StatusPending  = "pending"
	StatusSettled  = "settled"
	StatusFailed   = "failed"
	StatusRefunded = "refunded"
)

// Payment is one attempt at collecting funds for an order.
type Payment struct {
	ID          string `gorm:"primaryKey;type:char(26)"`
	OrderID     string `gorm:"type:char(26);index"`
	Provider    string `gorm:"size:32"`
	IntentID    string `gorm:"size:128;uniqueIndex:uk_payments_intent,priority:2"`
	AmountMinor int64
	Currency    string `gorm:"size:3"`
	Status      string `gorm:"size:16"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName overrides GORM's pluralizer.
func (Payment) TableName() string { return "payments" }

// Event is a raw webhook event log used for replay protection (UNIQUE on
// (provider, event_id)) and debugging.
type Event struct {
	ID         string `gorm:"primaryKey;type:char(26)"`
	Provider   string `gorm:"size:32"`
	EventID    string `gorm:"size:128"`
	IntentID   string `gorm:"size:128"`
	Type       string `gorm:"size:64"`
	Raw        string `gorm:"type:mediumtext"`
	ReceivedAt time.Time
}

// TableName overrides GORM's pluralizer.
func (Event) TableName() string { return "payment_events" }
