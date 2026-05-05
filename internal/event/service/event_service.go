// Package service implements the event.Service port.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quangdangfit/goticket/internal/event"
	"github.com/quangdangfit/goticket/internal/event/dto"
	"github.com/quangdangfit/goticket/internal/event/model"
)

type eventService struct {
	repo  event.Repository
	cache event.Cache
}

// New constructs an event.Service.
func New(repo event.Repository, cache event.Cache) event.Service {
	return &eventService{repo: repo, cache: cache}
}

func (s *eventService) Create(ctx context.Context, in dto.CreateEventInput) (*dto.Event, error) {
	if _, err := s.repo.GetVenue(ctx, in.VenueID); err != nil {
		return nil, fmt.Errorf("create event: %w", err)
	}
	now := time.Now().UTC()
	e := &model.Event{
		ID:          newID(),
		Title:       in.Title,
		Description: in.Description,
		Organizer:   in.Organizer,
		PosterURL:   in.PosterURL,
		Status:      model.StatusDraft,
		VenueID:     in.VenueID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateEvent(ctx, e); err != nil {
		return nil, err
	}
	return s.Detail(ctx, e.ID)
}

func (s *eventService) Update(ctx context.Context, id string, in dto.UpdateEventInput) (*dto.Event, error) {
	e, err := s.repo.GetEvent(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Title != nil {
		e.Title = *in.Title
	}
	if in.Description != nil {
		e.Description = *in.Description
	}
	if in.PosterURL != nil {
		e.PosterURL = *in.PosterURL
	}
	if in.Status != nil {
		e.Status = *in.Status
	}
	e.UpdatedAt = time.Now().UTC()
	if err := s.repo.UpdateEvent(ctx, e); err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Invalidate(ctx, id)
	}
	return s.Detail(ctx, id)
}

func (s *eventService) Detail(ctx context.Context, id string) (*dto.Event, error) {
	if s.cache != nil {
		if hit, ok := s.cache.GetEvent(ctx, id); ok {
			return hit, nil
		}
	}
	e, err := s.repo.GetEvent(ctx, id)
	if err != nil {
		return nil, err
	}
	v, err := s.repo.GetVenue(ctx, e.VenueID)
	if err != nil {
		return nil, err
	}
	shows, err := s.repo.ListShowtimes(ctx, e.ID)
	if err != nil {
		return nil, err
	}
	out := toDTO(e, v, shows)
	if s.cache != nil {
		s.cache.SetEvent(ctx, id, out)
	}
	return out, nil
}

func (s *eventService) List(ctx context.Context, q dto.EventQuery) ([]*dto.Event, int64, error) {
	rows, total, err := s.repo.ListEvents(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*dto.Event, 0, len(rows))
	for _, e := range rows {
		out = append(out, toDTO(e, nil, nil))
	}
	return out, total, nil
}

func toDTO(e *model.Event, v *model.Venue, shows []*model.Showtime) *dto.Event {
	out := &dto.Event{
		ID: e.ID, Title: e.Title, Description: e.Description,
		Organizer: e.Organizer, PosterURL: e.PosterURL, Status: e.Status,
		CreatedAt: e.CreatedAt,
	}
	if v != nil {
		out.Venue = &dto.Venue{ID: v.ID, Name: v.Name, Address: v.Address, City: v.City, Capacity: v.Capacity}
	}
	for _, sh := range shows {
		out.Showtimes = append(out.Showtimes, dto.Showtime{
			ID: sh.ID, StartsAt: sh.StartsAt, EndsAt: sh.EndsAt,
			SalesOpenAt: sh.SalesOpenAt, SalesCloseAt: sh.SalesCloseAt, Status: sh.Status,
		})
	}
	return out
}

func newID() string { return strings.ReplaceAll(uuid.NewString(), "-", "")[:26] }
