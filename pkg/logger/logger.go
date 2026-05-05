// Package logger provides a thin wrapper around log/slog producing JSON logs
// and helpers for carrying request-scoped fields (request_id, user_id) through
// context.Context.
package logger

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey int

const (
	keyRequestID ctxKey = iota
	keyUserID
	keyOrderID
)

// Init configures the default slog logger to emit JSON at the given level.
func Init(level string) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	slog.SetDefault(slog.New(h))
}

// WithRequestID stores a request id on ctx.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

// WithUserID stores a user id on ctx.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyUserID, id)
}

// WithOrderID stores an order id on ctx.
func WithOrderID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyOrderID, id)
}

// RequestID returns the request id stored on ctx, or "" if none.
func RequestID(ctx context.Context) string { s, _ := ctx.Value(keyRequestID).(string); return s }

// FromContext returns a slog.Logger pre-populated with any request-scoped
// fields present on ctx.
func FromContext(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if v, ok := ctx.Value(keyRequestID).(string); ok && v != "" {
		l = l.With("request_id", v)
	}
	if v, ok := ctx.Value(keyUserID).(string); ok && v != "" {
		l = l.With("user_id", v)
	}
	if v, ok := ctx.Value(keyOrderID).(string); ok && v != "" {
		l = l.With("order_id", v)
	}
	return l
}
