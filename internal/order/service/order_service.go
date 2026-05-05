// Package service implements the order.Service port. Checkout orchestrates
// idempotency → inventory.Hold → DB persistence inside a transaction.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/quangdangfit/goticket/internal/inventory"
	invdto "github.com/quangdangfit/goticket/internal/inventory/dto"
	"github.com/quangdangfit/goticket/internal/order"
	"github.com/quangdangfit/goticket/internal/order/dto"
	"github.com/quangdangfit/goticket/internal/order/model"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
	"github.com/quangdangfit/goticket/pkg/idempotency"
)

type orderService struct {
	repo  order.Repository
	inv   inventory.Inventory
	price order.PriceLookup
	promo order.PromoApplier
	idem  idempotency.Guard
}

// New constructs an order.Service.
func New(
	repo order.Repository,
	inv inventory.Inventory,
	price order.PriceLookup,
	promo order.PromoApplier,
	idem idempotency.Guard,
) order.Service {
	return &orderService{repo: repo, inv: inv, price: price, promo: promo, idem: idem}
}

func (s *orderService) Checkout(ctx context.Context, userID string, in dto.CheckoutInput) (*dto.CheckoutOutput, error) {
	if userID == "" {
		return nil, apperr.ErrUnauthorized
	}

	// 1. Idempotency: Redis fast-path replay → DB unique-key fallback at commit.
	if s.idem != nil {
		err := s.idem.Reserve(ctx, userID, in.IdempotencyKey, 15*time.Minute)
		if errors.Is(err, idempotency.ErrReplay) {
			if existing, err2 := s.repo.GetByIdempotency(ctx, userID, in.IdempotencyKey); err2 == nil {
				return &dto.CheckoutOutput{Order: toDTO(existing)}, nil
			}
		} else if err != nil {
			return nil, err
		}
	}

	// 2. Hold inventory atomically (Redis Lua).
	hold, err := s.inv.Hold(ctx, invdto.HoldInput{UserID: userID, Items: in.Items})
	if err != nil {
		return nil, err
	}

	// 3. Price + currency from ticket types (cross-domain port).
	var subtotal int64
	currency := "VND"
	items := make([]model.OrderItem, 0, len(in.Items))
	for _, it := range in.Items {
		price, cur, err := s.price.UnitPrice(ctx, it.TicketTypeID)
		if err != nil {
			_ = s.inv.Release(ctx, hold.ID)
			return nil, err
		}
		currency = cur
		subtotal += price * int64(it.Quantity)
		items = append(items, model.OrderItem{
			ID:             newID(),
			TicketTypeID:   it.TicketTypeID,
			Quantity:       it.Quantity,
			UnitPriceMinor: price,
		})
	}

	// 4. Promo (optional).
	var discount int64
	if in.PromoCode != "" && s.promo != nil {
		d, err := s.promo.Apply(ctx, in.PromoCode, userID, subtotal)
		if err != nil {
			_ = s.inv.Release(ctx, hold.ID)
			return nil, err
		}
		discount = d
	}

	// 5. Persist atomically.
	o := &model.Order{
		ID:            newID(),
		UserID:        userID,
		ShowtimeID:    in.ShowtimeID,
		HoldID:        hold.ID,
		Status:        model.StatusPending,
		SubtotalMinor: subtotal,
		DiscountMinor: discount,
		TotalMinor:    subtotal - discount,
		Currency:      currency,
		PromoCode:     in.PromoCode,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	for i := range items {
		items[i].OrderID = o.ID
	}
	idem := &model.IdempotencyKey{
		UserID: userID, Key: in.IdempotencyKey, OrderID: o.ID,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.CreateWithItems(ctx, o, items, idem); err != nil {
		_ = s.inv.Release(ctx, hold.ID)
		return nil, fmt.Errorf("persist order: %w", err)
	}
	o.Items = items
	return &dto.CheckoutOutput{Order: toDTO(o)}, nil
}

func (s *orderService) Get(ctx context.Context, userID, id string) (*dto.Order, error) {
	o, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if o.UserID != userID {
		return nil, apperr.ErrForbidden
	}
	d := toDTO(o)
	return &d, nil
}

// MarkPaid transitions a pending order to paid and confirms the inventory
// hold (making the deduction permanent).
func (s *orderService) MarkPaid(ctx context.Context, orderID string) error {
	if err := s.repo.UpdateStatus(ctx, orderID, model.StatusPending, model.StatusPaid); err != nil {
		return err
	}
	o, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	return s.inv.Confirm(ctx, o.HoldID)
}

// MarkFailed transitions pending → cancelled and releases the hold.
func (s *orderService) MarkFailed(ctx context.Context, orderID string) error {
	o, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, orderID, model.StatusPending, model.StatusCancelled); err != nil {
		return err
	}
	return s.inv.Release(ctx, o.HoldID)
}

func (s *orderService) Cancel(ctx context.Context, userID, id string) error {
	o, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if o.UserID != userID {
		return apperr.ErrForbidden
	}
	if err := s.repo.UpdateStatus(ctx, id, model.StatusPending, model.StatusCancelled); err != nil {
		return err
	}
	return s.inv.Release(ctx, o.HoldID)
}

func toDTO(o *model.Order) dto.Order {
	items := make([]dto.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, dto.OrderItem{
			TicketTypeID: it.TicketTypeID, Quantity: it.Quantity, UnitPriceMinor: it.UnitPriceMinor,
		})
	}
	return dto.Order{
		ID: o.ID, Status: string(o.Status), SubtotalMinor: o.SubtotalMinor,
		DiscountMinor: o.DiscountMinor, TotalMinor: o.TotalMinor, Currency: o.Currency,
		HoldID: o.HoldID, Items: items, CreatedAt: o.CreatedAt,
	}
}

func newID() string { return strings.ReplaceAll(uuid.NewString(), "-", "")[:26] }
