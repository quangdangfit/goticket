// Package service implements notification.Consumer using kafka-go.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/quangdangfit/goticket/internal/notification"
	"github.com/quangdangfit/goticket/internal/notification/templates"
	"github.com/quangdangfit/goticket/pkg/logger"
)

// Profile lookup is intentionally minimal — full DTOs are not loaded; the
// notification consumer reads only what it needs to render.
type Profile interface {
	Email(ctx context.Context, userID string) (string, error)
}

type kafkaConsumer struct {
	brokers []string
	groupID string
	send    notification.Sender
	users   Profile
}

// New builds a Consumer subscribed to the given Kafka brokers.
func New(brokers []string, groupID string, send notification.Sender, users Profile) notification.Consumer {
	return &kafkaConsumer{brokers: brokers, groupID: groupID, send: send, users: users}
}

type orderPaidPayload struct {
	OrderID    string `json:"order_id"`
	UserID     string `json:"user_id"`
	TotalMinor int64  `json:"total_minor"`
	Currency   string `json:"currency"`
}

func (c *kafkaConsumer) Run(ctx context.Context) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.brokers, GroupID: c.groupID,
		GroupTopics: []string{"order.paid", "payment.failed"},
	})
	defer func() { _ = r.Close() }()
	for {
		m, err := r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			logger.FromContext(ctx).Error("kafka fetch", "err", err)
			continue
		}
		if err := c.handle(ctx, m); err != nil {
			logger.FromContext(ctx).Error("notification handle", "topic", m.Topic, "err", err)
		}
		if err := r.CommitMessages(ctx, m); err != nil {
			logger.FromContext(ctx).Error("kafka commit", "err", err)
		}
	}
}

func (c *kafkaConsumer) handle(ctx context.Context, m kafka.Message) error {
	switch m.Topic {
	case "order.paid":
		var p orderPaidPayload
		if err := json.Unmarshal(m.Value, &p); err != nil {
			return fmt.Errorf("decode order.paid: %w", err)
		}
		email, err := c.users.Email(ctx, p.UserID)
		if err != nil {
			return err
		}
		subject, body := templates.OrderPaid(p.OrderID, p.TotalMinor, p.Currency)
		return c.send.Send(ctx, email, subject, body)
	case "payment.failed":
		var p orderPaidPayload
		_ = json.Unmarshal(m.Value, &p)
		email, err := c.users.Email(ctx, p.UserID)
		if err != nil {
			return err
		}
		subject, body := templates.PaymentFailed(p.OrderID)
		return c.send.Send(ctx, email, subject, body)
	default:
		return nil
	}
}
