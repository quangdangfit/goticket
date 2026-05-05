// Package dto holds request/response shapes for the ticket HTTP port.
package dto

// CreateTicketTypeInput is the body of POST /admin/showtimes/:id/tickets.
type CreateTicketTypeInput struct {
	Name         string `json:"name"           validate:"required,min=1,max=128"`
	Description  string `json:"description"    validate:"omitempty,max=512"`
	PriceMinor   int64  `json:"price_minor"    validate:"required,min=1,max=999999999999"`
	Currency     string `json:"currency"       validate:"required,len=3"`
	TotalQuota   int    `json:"total_quota"    validate:"required,min=1,max=1000000"`
	PerUserLimit int    `json:"per_user_limit" validate:"required,min=1,max=100"`
	HasSeatMap   bool   `json:"has_seat_map"`
}

// TicketType is the public-facing tier view.
type TicketType struct {
	ID           string `json:"id"`
	ShowtimeID   string `json:"showtime_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	PriceMinor   int64  `json:"price_minor"`
	Currency     string `json:"currency"`
	TotalQuota   int    `json:"total_quota"`
	Available    int    `json:"available"`
	PerUserLimit int    `json:"per_user_limit"`
	HasSeatMap   bool   `json:"has_seat_map"`
}
