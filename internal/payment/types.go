// Package payment is the gateway / webhook bounded context.
package payment

import (
	"context"

	"github.com/quangdangfit/goticket/internal/payment/dto"
	"github.com/quangdangfit/goticket/internal/payment/model"
)

// Gateway is the port to an external payment provider.
type Gateway interface {
	Name() string
	CreateIntent(ctx context.Context, in dto.IntentInput) (*dto.Intent, error)
	VerifyWebhook(headers map[string]string, body []byte) (*dto.WebhookEvent, error)
}

// Repository persists payment + raw webhook event records.
type Repository interface {
	CreatePayment(ctx context.Context, p *model.Payment) error
	GetByIntent(ctx context.Context, provider, intentID string) (*model.Payment, error)
	UpdateStatus(ctx context.Context, id, status string) error
	RecordEvent(ctx context.Context, e *model.Event) error
}

// OrderUpdater is the cross-domain port called when a webhook settles.
type OrderUpdater interface {
	MarkPaid(ctx context.Context, orderID string) error
	MarkFailed(ctx context.Context, orderID string) error
}

// Service is the user-facing port.
type Service interface {
	StartIntent(ctx context.Context, orderID string, amountMinor int64, currency string) (*dto.Intent, error)
	HandleWebhook(ctx context.Context, headers map[string]string, body []byte) error
}
