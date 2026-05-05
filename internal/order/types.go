// Package order is the checkout / order-lifecycle bounded context.
package order

import (
	"context"

	"github.com/quangdangfit/goticket/internal/order/dto"
	"github.com/quangdangfit/goticket/internal/order/model"
)

// Repository persists orders, items, and idempotency keys atomically.
type Repository interface {
	CreateWithItems(ctx context.Context, o *model.Order, items []model.OrderItem, idem *model.IdempotencyKey) error
	GetByID(ctx context.Context, id string) (*model.Order, error)
	GetByIdempotency(ctx context.Context, userID, key string) (*model.Order, error)
	UpdateStatus(ctx context.Context, id string, from, to model.OrderStatus) error
}

// PriceLookup returns the unit price for a ticket type.
type PriceLookup interface {
	UnitPrice(ctx context.Context, ticketTypeID string) (priceMinor int64, currency string, err error)
}

// PromoApplier returns the discount minor for a code (or 0).
type PromoApplier interface {
	Apply(ctx context.Context, code, userID string, subtotal int64) (discountMinor int64, err error)
}

// Service is the user-facing port mounted on HTTP.
type Service interface {
	Checkout(ctx context.Context, userID string, in dto.CheckoutInput) (*dto.CheckoutOutput, error)
	Get(ctx context.Context, userID, id string) (*dto.Order, error)
	Cancel(ctx context.Context, userID, id string) error

	// MarkPaid / MarkFailed satisfy payment.OrderUpdater.
	MarkPaid(ctx context.Context, orderID string) error
	MarkFailed(ctx context.Context, orderID string) error
}
