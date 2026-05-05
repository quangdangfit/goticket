// Package dto holds request/response shapes for the event HTTP port.
package dto

import "time"

// CreateEventInput is the body of POST /admin/events.
type CreateEventInput struct {
	Title       string `json:"title"        validate:"required,min=1,max=255"`
	Description string `json:"description"  validate:"required"`
	Organizer   string `json:"organizer"    validate:"required,max=255"`
	PosterURL   string `json:"poster_url"   validate:"omitempty,url,max=1024"`
	VenueID     string `json:"venue_id"     validate:"required,len=26"`
}

// UpdateEventInput is the body of PATCH /admin/events/:id.
type UpdateEventInput struct {
	Title       *string `json:"title,omitempty"        validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	PosterURL   *string `json:"poster_url,omitempty"   validate:"omitempty,url,max=1024"`
	Status      *string `json:"status,omitempty"       validate:"omitempty,oneof=draft published cancelled"`
}

// EventQuery filters listing.
type EventQuery struct {
	Status string `form:"status"`
	City   string `form:"city"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

// Event is the public-facing representation.
type Event struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Organizer   string     `json:"organizer"`
	PosterURL   string     `json:"poster_url"`
	Status      string     `json:"status"`
	Venue       *Venue     `json:"venue,omitempty"`
	Showtimes   []Showtime `json:"showtimes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Venue mirrors the model in JSON shape.
type Venue struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Capacity int    `json:"capacity"`
}

// Showtime is a public showtime view.
type Showtime struct {
	ID           string    `json:"id"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	SalesOpenAt  time.Time `json:"sales_open_at"`
	SalesCloseAt time.Time `json:"sales_close_at"`
	Status       string    `json:"status"`
}
