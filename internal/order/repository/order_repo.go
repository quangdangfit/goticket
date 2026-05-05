// Package repository implements the order.Repository port against MySQL.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/order"
	"github.com/quangdangfit/goticket/internal/order/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

type mysqlOrderRepository struct{ db dbs.MySQL }

// New returns an order.Repository backed by MySQL.
func New(db dbs.MySQL) order.Repository { return &mysqlOrderRepository{db: db} }

func (r *mysqlOrderRepository) CreateWithItems(
	ctx context.Context, o *model.Order, items []model.OrderItem, idem *model.IdempotencyKey,
) error {
	return r.db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if idem != nil {
			if err := tx.Create(idem).Error; err != nil {
				return fmt.Errorf("create idempotency: %w", err)
			}
		}
		if err := tx.Create(o).Error; err != nil {
			return fmt.Errorf("create order: %w", err)
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return fmt.Errorf("create items: %w", err)
			}
		}
		return nil
	})
}

func (r *mysqlOrderRepository) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var o model.Order
	err := r.db.DB(ctx).Preload("Items").Where("id = ?", id).First(&o).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return &o, nil
}

func (r *mysqlOrderRepository) GetByIdempotency(ctx context.Context, userID, key string) (*model.Order, error) {
	var k model.IdempotencyKey
	err := r.db.DB(ctx).Where("user_id = ? AND `key` = ?", userID, key).First(&k).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get idempotency: %w", err)
	}
	return r.GetByID(ctx, k.OrderID)
}

func (r *mysqlOrderRepository) UpdateStatus(ctx context.Context, id string, from, to model.OrderStatus) error {
	return r.db.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var o model.Order
		// SELECT … FOR UPDATE to serialize FSM transitions.
		err := tx.Clauses().Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", id).First(&o).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperr.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("lock order: %w", err)
		}
		if o.Status != from {
			return fmt.Errorf("status %s != %s: %w", o.Status, from, apperr.ErrConflict)
		}
		if err := tx.Model(&o).Update("status", to).Error; err != nil {
			return fmt.Errorf("update status: %w", err)
		}
		return nil
	})
}
