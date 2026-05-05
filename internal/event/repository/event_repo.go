// Package repository implements the event.Repository port against MySQL/GORM.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/event"
	"github.com/quangdangfit/goticket/internal/event/dto"
	"github.com/quangdangfit/goticket/internal/event/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

type mysqlEventRepository struct{ db dbs.MySQL }

// New returns an event.Repository backed by MySQL.
func New(db dbs.MySQL) event.Repository { return &mysqlEventRepository{db: db} }

func (r *mysqlEventRepository) CreateEvent(ctx context.Context, e *model.Event) error {
	if err := r.db.DB(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

func (r *mysqlEventRepository) UpdateEvent(ctx context.Context, e *model.Event) error {
	if err := r.db.DB(ctx).Save(e).Error; err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func (r *mysqlEventRepository) GetEvent(ctx context.Context, id string) (*model.Event, error) {
	var e model.Event
	err := r.db.DB(ctx).Where("id = ?", id).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}
	return &e, nil
}

func (r *mysqlEventRepository) ListEvents(ctx context.Context, q dto.EventQuery) ([]*model.Event, int64, error) {
	tx := r.db.DB(ctx).Model(&model.Event{})
	if q.Status != "" {
		tx = tx.Where("status = ?", q.Status)
	}
	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}
	page, limit := q.Page, q.Limit
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	var out []*model.Event
	if err := tx.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	return out, total, nil
}

func (r *mysqlEventRepository) GetVenue(ctx context.Context, id string) (*model.Venue, error) {
	var v model.Venue
	err := r.db.DB(ctx).Where("id = ?", id).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get venue: %w", err)
	}
	return &v, nil
}

func (r *mysqlEventRepository) ListShowtimes(ctx context.Context, eventID string) ([]*model.Showtime, error) {
	var out []*model.Showtime
	if err := r.db.DB(ctx).Where("event_id = ?", eventID).Order("starts_at ASC").Find(&out).Error; err != nil {
		return nil, fmt.Errorf("list showtimes: %w", err)
	}
	return out, nil
}

func (r *mysqlEventRepository) GetShowtime(ctx context.Context, id string) (*model.Showtime, error) {
	var s model.Showtime
	err := r.db.DB(ctx).Where("id = ?", id).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get showtime: %w", err)
	}
	return &s, nil
}
