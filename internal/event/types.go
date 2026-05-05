// Package event is the events / venues / showtimes bounded context.
package event

import (
	"context"

	"github.com/quangdangfit/goticket/internal/event/dto"
	"github.com/quangdangfit/goticket/internal/event/model"
)

// Repository persists events, venues, and showtimes.
type Repository interface {
	CreateEvent(ctx context.Context, e *model.Event) error
	UpdateEvent(ctx context.Context, e *model.Event) error
	GetEvent(ctx context.Context, id string) (*model.Event, error)
	ListEvents(ctx context.Context, q dto.EventQuery) ([]*model.Event, int64, error)
	GetVenue(ctx context.Context, id string) (*model.Venue, error)
	ListShowtimes(ctx context.Context, eventID string) ([]*model.Showtime, error)
	GetShowtime(ctx context.Context, id string) (*model.Showtime, error)
}

// Cache is a read-through cache for event detail.
type Cache interface {
	GetEvent(ctx context.Context, id string) (*dto.Event, bool)
	SetEvent(ctx context.Context, id string, e *dto.Event)
	Invalidate(ctx context.Context, id string)
}

// Service is the user-facing port.
type Service interface {
	Create(ctx context.Context, in dto.CreateEventInput) (*dto.Event, error)
	Update(ctx context.Context, id string, in dto.UpdateEventInput) (*dto.Event, error)
	Detail(ctx context.Context, id string) (*dto.Event, error)
	List(ctx context.Context, q dto.EventQuery) ([]*dto.Event, int64, error)
}
