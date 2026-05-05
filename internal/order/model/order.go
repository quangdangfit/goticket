// Package model holds GORM entities for the order bounded context.
package model

import "time"

// OrderStatus values.
type OrderStatus string

// Lifecycle states.
const (
	StatusPending   OrderStatus = "pending"
	StatusPaid      OrderStatus = "paid"
	StatusFulfilled OrderStatus = "fulfilled"
	StatusCancelled OrderStatus = "cancelled"
	StatusExpired   OrderStatus = "expired"
)

// Order is a checkout aggregate root.
type Order struct {
	ID             string `gorm:"primaryKey;type:char(26)"`
	UserID         string `gorm:"type:char(26);index"`
	ShowtimeID     string `gorm:"type:char(26)"`
	HoldID         string `gorm:"type:char(26);index"`
	Status         OrderStatus `gorm:"size:16"`
	SubtotalMinor  int64
	DiscountMinor  int64
	TotalMinor     int64
	Currency       string `gorm:"size:3"`
	PromoCode      string `gorm:"size:64"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Items          []OrderItem `gorm:"foreignKey:OrderID"`
}

// TableName overrides GORM's pluralizer.
func (Order) TableName() string { return "orders" }

// OrderItem is a line item.
type OrderItem struct {
	ID             string `gorm:"primaryKey;type:char(26)"`
	OrderID        string `gorm:"type:char(26);index"`
	TicketTypeID   string `gorm:"type:char(26)"`
	Quantity       int
	UnitPriceMinor int64
}

// TableName overrides GORM's pluralizer.
func (OrderItem) TableName() string { return "order_items" }

// IdempotencyKey is the durable second-line guard against duplicate checkouts.
type IdempotencyKey struct {
	UserID    string    `gorm:"primaryKey;type:char(26)"`
	Key       string    `gorm:"primaryKey;column:key;size:128"`
	OrderID   string    `gorm:"type:char(26)"`
	CreatedAt time.Time
}

// TableName overrides GORM's pluralizer.
func (IdempotencyKey) TableName() string { return "idempotency_keys" }
