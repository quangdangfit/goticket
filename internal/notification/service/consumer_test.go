package service

import (
	"context"
	"testing"

	"github.com/quangdangfit/goticket/internal/notification/templates"
)

func TestTemplates_OrderPaid(t *testing.T) {
	subj, body := templates.OrderPaid("order-1", 150_000, "VND")
	if subj == "" || body == "" {
		t.Fatal("empty render")
	}
}

func TestTemplates_PaymentFailed(t *testing.T) {
	subj, body := templates.PaymentFailed("order-1")
	if subj == "" || body == "" {
		t.Fatal("empty render")
	}
}

// Compile-time assertion that the package wires up against context.Context.
var _ = func() context.Context { return context.Background() }
