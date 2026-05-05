// Package repository implements the ticket.Repository port.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/ticket"
	"github.com/quangdangfit/goticket/internal/ticket/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

type mysqlTicketRepository struct{ db dbs.MySQL }

// New returns a ticket.Repository backed by MySQL.
func New(db dbs.MySQL) ticket.Repository { return &mysqlTicketRepository{db: db} }

func (r *mysqlTicketRepository) CreateType(ctx context.Context, t *model.TicketType) error {
	if err := r.db.DB(ctx).Create(t).Error; err != nil {
		return fmt.Errorf("create ticket type: %w", err)
	}
	return nil
}

func (r *mysqlTicketRepository) GetType(ctx context.Context, id string) (*model.TicketType, error) {
	var t model.TicketType
	err := r.db.DB(ctx).Where("id = ?", id).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get ticket type: %w", err)
	}
	return &t, nil
}

func (r *mysqlTicketRepository) ListByShowtime(ctx context.Context, showtimeID string) ([]*model.TicketType, error) {
	var out []*model.TicketType
	if err := r.db.DB(ctx).Where("showtime_id = ?", showtimeID).Find(&out).Error; err != nil {
		return nil, fmt.Errorf("list ticket types: %w", err)
	}
	return out, nil
}
