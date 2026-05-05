// Package sender holds Sender implementations.
package sender

import (
	"context"

	"github.com/quangdangfit/goticket/internal/notification"
	"github.com/quangdangfit/goticket/pkg/logger"
)

type logSender struct{}

// NewLog returns a Sender that just logs the outbound message — useful in
// dev and CI before SMTP/SMS providers are wired.
func NewLog() notification.Sender { return logSender{} }

func (logSender) Send(ctx context.Context, to, subject, body string) error {
	logger.FromContext(ctx).Info("notification.send",
		"to", to, "subject", subject, "body_len", len(body))
	return nil
}
