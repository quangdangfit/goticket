// Package service implements promo.Service.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quangdangfit/goticket/internal/promo"
	"github.com/quangdangfit/goticket/internal/promo/dto"
	"github.com/quangdangfit/goticket/internal/promo/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

type promoService struct{ repo promo.Repository }

// New constructs a promo.Service.
func New(repo promo.Repository) promo.Service { return &promoService{repo: repo} }

func (s *promoService) Create(ctx context.Context, in dto.CreatePromoInput) error {
	now := time.Now().UTC()
	return s.repo.Create(ctx, &model.Code{
		Code: in.Code, Type: in.Type, ValueMinor: in.ValueMinor, Percent: in.Percent,
		MaxUses: in.MaxUses, PerUserLimit: in.PerUserLimit,
		StartsAt: in.StartsAt, ExpiresAt: in.ExpiresAt, CreatedAt: now,
	})
}

func (s *promoService) Apply(ctx context.Context, code, userID string, subtotal int64) (int64, error) {
	c, err := s.repo.Get(ctx, code)
	if err != nil {
		return 0, err
	}
	now := time.Now().UTC()
	if now.Before(c.StartsAt) || now.After(c.ExpiresAt) {
		return 0, fmt.Errorf("promo window: %w", apperr.ErrConflict)
	}
	if c.Used >= c.MaxUses {
		return 0, apperr.ErrSoldOut
	}
	if used, err := s.repo.UserUseCount(ctx, code, userID); err != nil {
		return 0, err
	} else if used >= c.PerUserLimit {
		return 0, fmt.Errorf("per-user limit: %w", apperr.ErrConflict)
	}
	switch c.Type {
	case model.TypeFixed:
		if c.ValueMinor > subtotal {
			return subtotal, nil
		}
		return c.ValueMinor, nil
	case model.TypePercent:
		return subtotal * int64(c.Percent) / 100, nil
	default:
		return 0, apperr.ErrValidation
	}
}

func (s *promoService) Redeem(ctx context.Context, code, userID, orderID string, discount int64) error {
	if err := s.repo.IncrementUsed(ctx, code); err != nil {
		return err
	}
	return s.repo.RecordRedemption(ctx, &model.Redemption{
		ID: strings.ReplaceAll(uuid.NewString(), "-", "")[:26],
		Code: code, UserID: userID, OrderID: orderID,
		DiscountMinor: discount, CreatedAt: time.Now().UTC(),
	})
}
