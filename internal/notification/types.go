// Package notification is a pure consumer of order/payment Kafka topics.
// It owns no HTTP surface and writes no DB rows.
package notification

import "context"

// Sender ships rendered messages out (email, SMS, push).
type Sender interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Consumer is started in main.go via Run(ctx). It exits when ctx is done.
type Consumer interface {
	Run(ctx context.Context) error
}
