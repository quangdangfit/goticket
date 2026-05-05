// Package service implements the payment.Service port. Outbound provider
// calls are wrapped in a circuit breaker so a failing gateway can't drag
// down checkout latency.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sony/gobreaker"

	"github.com/quangdangfit/goticket/internal/dbs"
	"github.com/quangdangfit/goticket/internal/payment"
	"github.com/quangdangfit/goticket/internal/payment/dto"
	"github.com/quangdangfit/goticket/internal/payment/model"
	"github.com/quangdangfit/goticket/pkg/logger"
)

type paymentService struct {
	repo    payment.Repository
	gw      payment.Gateway
	orders  payment.OrderUpdater
	breaker *gobreaker.CircuitBreaker
	pub     dbs.KafkaPublisher
}

// New constructs a payment.Service.
func New(repo payment.Repository, gw payment.Gateway, orders payment.OrderUpdater, pub dbs.KafkaPublisher) payment.Service {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "payment-gateway-" + gw.Name(),
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     20 * time.Second,
		ReadyToTrip: func(c gobreaker.Counts) bool { return c.ConsecutiveFailures > 3 },
	})
	return &paymentService{repo: repo, gw: gw, orders: orders, breaker: cb, pub: pub}
}

func (s *paymentService) StartIntent(ctx context.Context, orderID string, amount int64, currency string) (*dto.Intent, error) {
	v, err := s.breaker.Execute(func() (any, error) {
		return s.gw.CreateIntent(ctx, dto.IntentInput{OrderID: orderID, AmountMinor: amount, Currency: currency})
	})
	if err != nil {
		return nil, fmt.Errorf("create intent: %w", err)
	}
	intent := v.(*dto.Intent)
	now := time.Now().UTC()
	if err := s.repo.CreatePayment(ctx, &model.Payment{
		ID:          newID(),
		OrderID:     orderID,
		Provider:    s.gw.Name(),
		IntentID:    intent.IntentID,
		AmountMinor: amount,
		Currency:    currency,
		Status:      model.StatusPending,
		CreatedAt:   now, UpdatedAt: now,
	}); err != nil {
		return nil, err
	}
	return intent, nil
}

func (s *paymentService) HandleWebhook(ctx context.Context, headers map[string]string, body []byte) error {
	ev, err := s.gw.VerifyWebhook(headers, body)
	if err != nil {
		return err
	}
	// Replay protection via UNIQUE (provider, event_id) on payment_events.
	if err := s.repo.RecordEvent(ctx, &model.Event{
		ID:         newID(),
		Provider:   ev.Provider,
		EventID:    ev.EventID,
		IntentID:   ev.IntentID,
		Type:       ev.Type,
		Raw:        string(ev.Raw),
		ReceivedAt: time.Now().UTC(),
	}); err != nil {
		// Best-effort: if it's a duplicate, nothing to do.
		logger.FromContext(ctx).Warn("record event", "err", err)
	}
	pay, err := s.repo.GetByIntent(ctx, ev.Provider, ev.IntentID)
	if err != nil {
		return err
	}
	switch ev.Type {
	case "payment.settled":
		if err := s.repo.UpdateStatus(ctx, pay.ID, model.StatusSettled); err != nil {
			return err
		}
		if s.orders != nil {
			if err := s.orders.MarkPaid(ctx, pay.OrderID); err != nil {
				return err
			}
		}
		if s.pub != nil {
			_ = s.pub.Publish(ctx, "order.paid", []byte(pay.OrderID), body)
		}
	case "payment.failed":
		if err := s.repo.UpdateStatus(ctx, pay.ID, model.StatusFailed); err != nil {
			return err
		}
		if s.orders != nil {
			if err := s.orders.MarkFailed(ctx, pay.OrderID); err != nil {
				return err
			}
		}
		if s.pub != nil {
			_ = s.pub.Publish(ctx, "payment.failed", []byte(pay.OrderID), body)
		}
	}
	return nil
}

func newID() string { return strings.ReplaceAll(uuid.NewString(), "-", "")[:26] }
