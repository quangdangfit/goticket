// Package promo is the discount-codes bounded context.
package promo

import (
	"context"

	"github.com/quangdangfit/goticket/internal/promo/dto"
	"github.com/quangdangfit/goticket/internal/promo/model"
)

// Repository persists codes and redemptions.
type Repository interface {
	Create(ctx context.Context, c *model.Code) error
	Get(ctx context.Context, code string) (*model.Code, error)
	UserUseCount(ctx context.Context, code, userID string) (int, error)
	IncrementUsed(ctx context.Context, code string) error
	RecordRedemption(ctx context.Context, r *model.Redemption) error
}

// Service is the user-facing port.
type Service interface {
	Create(ctx context.Context, in dto.CreatePromoInput) error
	Apply(ctx context.Context, code, userID string, subtotal int64) (int64, error)
	Redeem(ctx context.Context, code, userID, orderID string, discount int64) error
}
