// Package ticket is the ticket-types / seats bounded context.
package ticket

import (
	"context"

	"github.com/quangdangfit/goticket/internal/ticket/dto"
	"github.com/quangdangfit/goticket/internal/ticket/model"
)

// Repository persists ticket types and seats.
type Repository interface {
	CreateType(ctx context.Context, t *model.TicketType) error
	GetType(ctx context.Context, id string) (*model.TicketType, error)
	ListByShowtime(ctx context.Context, showtimeID string) ([]*model.TicketType, error)
}

// AvailabilityReader returns the live remaining count for a ticket type
// (sourced from the inventory bounded context).
type AvailabilityReader interface {
	Available(ctx context.Context, showtimeID, ticketTypeID string) (int, error)
}

// Service is the user-facing port.
type Service interface {
	Create(ctx context.Context, showtimeID string, in dto.CreateTicketTypeInput) (*dto.TicketType, error)
	ListByShowtime(ctx context.Context, showtimeID string) ([]*dto.TicketType, error)
	// UnitPrice satisfies order.PriceLookup.
	UnitPrice(ctx context.Context, ticketTypeID string) (int64, string, error)
}
