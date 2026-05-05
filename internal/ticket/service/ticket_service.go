// Package service implements the ticket.Service port.
package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quangdangfit/goticket/internal/ticket"
	"github.com/quangdangfit/goticket/internal/ticket/dto"
	"github.com/quangdangfit/goticket/internal/ticket/model"
)

type ticketService struct {
	repo ticket.Repository
	avail ticket.AvailabilityReader
}

// New constructs a ticket.Service. avail may be nil during phase 3
// (before the inventory domain ships); it is wired in later.
func New(repo ticket.Repository, avail ticket.AvailabilityReader) ticket.Service {
	return &ticketService{repo: repo, avail: avail}
}

func (s *ticketService) Create(ctx context.Context, showtimeID string, in dto.CreateTicketTypeInput) (*dto.TicketType, error) {
	now := time.Now().UTC()
	t := &model.TicketType{
		ID:           strings.ReplaceAll(uuid.NewString(), "-", "")[:26],
		ShowtimeID:   showtimeID,
		Name:         in.Name,
		Description:  in.Description,
		PriceMinor:   in.PriceMinor,
		Currency:     in.Currency,
		TotalQuota:   in.TotalQuota,
		PerUserLimit: in.PerUserLimit,
		HasSeatMap:   in.HasSeatMap,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateType(ctx, t); err != nil {
		return nil, err
	}
	return s.toDTO(ctx, t), nil
}

func (s *ticketService) ListByShowtime(ctx context.Context, showtimeID string) ([]*dto.TicketType, error) {
	rows, err := s.repo.ListByShowtime(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	out := make([]*dto.TicketType, 0, len(rows))
	for _, t := range rows {
		out = append(out, s.toDTO(ctx, t))
	}
	return out, nil
}

// UnitPrice satisfies order.PriceLookup so the order service can resolve
// price/currency without depending on the ticket repository directly.
func (s *ticketService) UnitPrice(ctx context.Context, ticketTypeID string) (int64, string, error) {
	t, err := s.repo.GetType(ctx, ticketTypeID)
	if err != nil {
		return 0, "", err
	}
	return t.PriceMinor, t.Currency, nil
}

func (s *ticketService) toDTO(ctx context.Context, t *model.TicketType) *dto.TicketType {
	avail := t.TotalQuota
	if s.avail != nil {
		if v, err := s.avail.Available(ctx, t.ShowtimeID, t.ID); err == nil {
			avail = v
		}
	}
	return &dto.TicketType{
		ID: t.ID, ShowtimeID: t.ShowtimeID, Name: t.Name, Description: t.Description,
		PriceMinor: t.PriceMinor, Currency: t.Currency,
		TotalQuota: t.TotalQuota, Available: avail,
		PerUserLimit: t.PerUserLimit, HasSeatMap: t.HasSeatMap,
	}
}
