// Package stripe is a minimal Stripe-shaped gateway adapter. The HMAC
// verification mirrors Stripe's `Stripe-Signature` v1 scheme (sha256 over
// "<timestamp>.<body>"). The real Stripe SDK is intentionally not pulled
// in — replace this with stripe-go when integrating for real.
package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/quangdangfit/goticket/internal/payment"
	"github.com/quangdangfit/goticket/internal/payment/dto"
)

type stripeGateway struct {
	secret string
	apiKey string
}

// New returns a Stripe-shaped payment.Gateway.
func New(secret, apiKey string) payment.Gateway {
	return &stripeGateway{secret: secret, apiKey: apiKey}
}

func (stripeGateway) Name() string { return "stripe" }

func (stripeGateway) CreateIntent(_ context.Context, _ dto.IntentInput) (*dto.Intent, error) {
	return nil, errors.New("stripe.CreateIntent: not wired in this build")
}

// VerifyWebhook checks the Stripe-Signature header against HMAC-SHA256 of
// "<timestamp>.<body>" using the configured secret. Tolerance: 5 minutes.
func (g stripeGateway) VerifyWebhook(headers map[string]string, body []byte) (*dto.WebhookEvent, error) {
	sig := headers["Stripe-Signature"]
	if sig == "" {
		return nil, errors.New("missing stripe-signature header")
	}
	var ts, v1 string
	for _, part := range strings.Split(sig, ",") {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			ts = kv[1]
		case "v1":
			v1 = kv[1]
		}
	}
	if ts == "" || v1 == "" {
		return nil, errors.New("malformed stripe-signature header")
	}
	mac := hmac.New(sha256.New, []byte(g.secret))
	_, _ = fmt.Fprintf(mac, "%s.%s", ts, body)
	want := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(want), []byte(v1)) {
		return nil, errors.New("bad stripe signature")
	}
	// Timestamp skew: ts is unix seconds.
	var tsec int64
	_, err := fmt.Sscanf(ts, "%d", &tsec)
	if err != nil {
		return nil, fmt.Errorf("bad timestamp: %w", err)
	}
	if time.Since(time.Unix(tsec, 0)) > 5*time.Minute {
		return nil, errors.New("stripe signature too old")
	}
	return &dto.WebhookEvent{Provider: "stripe", Raw: body}, nil
}
