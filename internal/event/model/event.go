// Package model holds GORM entities for the event bounded context.
package model

import "time"

// Event lifecycle states.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusCancelled = "cancelled"
)

// Showtime lifecycle states.
const (
	ShowtimeScheduled = "scheduled"
	ShowtimeOnSale    = "on_sale"
	ShowtimeClosed    = "closed"
	ShowtimeCancelled = "cancelled"
)

// Venue is the physical location where an event is held.
type Venue struct {
	ID        string `gorm:"primaryKey;type:char(26)"`
	Name      string `gorm:"size:255"`
	Address   string `gorm:"size:512"`
	City      string `gorm:"size:128"`
	Capacity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName overrides GORM's pluralizer.
func (Venue) TableName() string { return "venues" }

// Event is a sellable concert/show with one or more showtimes.
type Event struct {
	ID          string `gorm:"primaryKey;type:char(26)"`
	Title       string `gorm:"size:255"`
	Description string `gorm:"type:text"`
	Organizer   string `gorm:"size:255"`
	PosterURL   string `gorm:"size:1024"`
	Status      string `gorm:"size:16"`
	VenueID     string `gorm:"type:char(26);index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// TableName overrides GORM's pluralizer.
func (Event) TableName() string { return "events" }

// Showtime is a single performance slot of an event with its own sales window.
type Showtime struct {
	ID            string `gorm:"primaryKey;type:char(26)"`
	EventID       string `gorm:"type:char(26);index"`
	StartsAt      time.Time
	EndsAt        time.Time
	SalesOpenAt   time.Time
	SalesCloseAt  time.Time
	Status        string `gorm:"size:16"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName overrides GORM's pluralizer.
func (Showtime) TableName() string { return "showtimes" }
