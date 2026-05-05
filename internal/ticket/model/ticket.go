// Package model holds GORM entities for the ticket bounded context.
package model

import "time"

// TicketType is a price tier within a showtime (e.g. VIP, Standard).
type TicketType struct {
	ID            string `gorm:"primaryKey;type:char(26)"`
	ShowtimeID    string `gorm:"type:char(26);index"`
	Name          string `gorm:"size:128"`
	Description   string `gorm:"size:512"`
	PriceMinor    int64
	Currency      string `gorm:"size:3"`
	TotalQuota    int
	PerUserLimit  int
	HasSeatMap    bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName overrides GORM's pluralizer.
func (TicketType) TableName() string { return "ticket_types" }

// Seat is an addressable slot within a ticket type when seat-map is enabled.
type Seat struct {
	ID           string `gorm:"primaryKey;type:char(26)"`
	TicketTypeID string `gorm:"type:char(26);index"`
	Section      string `gorm:"size:64"`
	RowLabel     string `gorm:"size:16"`
	SeatNumber   string `gorm:"size:16"`
}

// TableName overrides GORM's pluralizer.
func (Seat) TableName() string { return "seats" }
