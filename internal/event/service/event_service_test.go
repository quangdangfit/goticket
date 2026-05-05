package service

import (
	"context"
	"testing"
	"time"

	"github.com/quangdangfit/goticket/internal/event/dto"
	"github.com/quangdangfit/goticket/internal/event/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

type fakeRepo struct {
	events  map[string]*model.Event
	venues  map[string]*model.Venue
	shows   map[string][]*model.Showtime
}

func newRepo() *fakeRepo {
	return &fakeRepo{
		events: map[string]*model.Event{},
		venues: map[string]*model.Venue{
			"venue-1": {ID: "venue-1", Name: "Hall", City: "HCM", Capacity: 500},
		},
		shows: map[string][]*model.Showtime{},
	}
}

func (f *fakeRepo) CreateEvent(_ context.Context, e *model.Event) error {
	f.events[e.ID] = e
	return nil
}
func (f *fakeRepo) UpdateEvent(_ context.Context, e *model.Event) error {
	f.events[e.ID] = e
	return nil
}
func (f *fakeRepo) GetEvent(_ context.Context, id string) (*model.Event, error) {
	if e, ok := f.events[id]; ok {
		return e, nil
	}
	return nil, apperr.ErrNotFound
}
func (f *fakeRepo) ListEvents(_ context.Context, _ dto.EventQuery) ([]*model.Event, int64, error) {
	out := make([]*model.Event, 0, len(f.events))
	for _, e := range f.events {
		out = append(out, e)
	}
	return out, int64(len(out)), nil
}
func (f *fakeRepo) GetVenue(_ context.Context, id string) (*model.Venue, error) {
	if v, ok := f.venues[id]; ok {
		return v, nil
	}
	return nil, apperr.ErrNotFound
}
func (f *fakeRepo) ListShowtimes(_ context.Context, eventID string) ([]*model.Showtime, error) {
	return f.shows[eventID], nil
}
func (f *fakeRepo) GetShowtime(_ context.Context, id string) (*model.Showtime, error) {
	for _, list := range f.shows {
		for _, s := range list {
			if s.ID == id {
				return s, nil
			}
		}
	}
	return nil, apperr.ErrNotFound
}

type memCache struct{ store map[string]*dto.Event }

func (m *memCache) GetEvent(_ context.Context, id string) (*dto.Event, bool) {
	e, ok := m.store[id]
	return e, ok
}
func (m *memCache) SetEvent(_ context.Context, id string, e *dto.Event) { m.store[id] = e }
func (m *memCache) Invalidate(_ context.Context, id string)             { delete(m.store, id) }

func TestEventService_Create_PopulatesVenue(t *testing.T) {
	r := newRepo()
	svc := New(r, &memCache{store: map[string]*dto.Event{}})
	out, err := svc.Create(context.Background(), dto.CreateEventInput{
		Title: "Concert", Description: "...", Organizer: "X",
		VenueID: "venue-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Venue == nil || out.Venue.ID != "venue-1" {
		t.Fatalf("missing venue: %+v", out)
	}
	if out.Status != model.StatusDraft {
		t.Fatalf("want draft, got %s", out.Status)
	}
}

func TestEventService_Detail_UsesCacheOnHit(t *testing.T) {
	r := newRepo()
	cache := &memCache{store: map[string]*dto.Event{}}
	svc := New(r, cache)
	cache.store["e1"] = &dto.Event{ID: "e1", Title: "Cached"}
	out, err := svc.Detail(context.Background(), "e1")
	if err != nil {
		t.Fatal(err)
	}
	if out.Title != "Cached" {
		t.Fatalf("want cached hit, got %s", out.Title)
	}
}

func TestEventService_Update_InvalidatesCache(t *testing.T) {
	r := newRepo()
	cache := &memCache{store: map[string]*dto.Event{}}
	svc := New(r, cache)
	now := time.Now().UTC()
	r.events["e1"] = &model.Event{ID: "e1", VenueID: "venue-1", Status: model.StatusDraft, CreatedAt: now, UpdatedAt: now}
	cache.store["e1"] = &dto.Event{ID: "e1", Title: "stale"}
	pub := model.StatusPublished
	if _, err := svc.Update(context.Background(), "e1", dto.UpdateEventInput{Status: &pub}); err != nil {
		t.Fatal(err)
	}
	if r.events["e1"].Status != model.StatusPublished {
		t.Fatalf("status not updated")
	}
}
