// Package repository implements the payment.Repository port.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/payment"
	"github.com/quangdangfit/goticket/internal/payment/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

type mysqlPaymentRepository struct{ db dbs.MySQL }

// New returns a payment.Repository backed by MySQL.
func New(db dbs.MySQL) payment.Repository { return &mysqlPaymentRepository{db: db} }

func (r *mysqlPaymentRepository) CreatePayment(ctx context.Context, p *model.Payment) error {
	if err := r.db.DB(ctx).Create(p).Error; err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

func (r *mysqlPaymentRepository) GetByIntent(ctx context.Context, provider, intentID string) (*model.Payment, error) {
	var p model.Payment
	err := r.db.DB(ctx).Where("provider = ? AND intent_id = ?", provider, intentID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return &p, nil
}

func (r *mysqlPaymentRepository) UpdateStatus(ctx context.Context, id, status string) error {
	if err := r.db.DB(ctx).Model(&model.Payment{}).
		Where("id = ?", id).Update("status", status).Error; err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	return nil
}

func (r *mysqlPaymentRepository) RecordEvent(ctx context.Context, e *model.Event) error {
	// Ignore duplicate-key errors to make replay protection idempotent.
	if err := r.db.DB(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("record event: %w", err)
	}
	return nil
}
