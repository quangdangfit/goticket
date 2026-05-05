// Package mock is a no-op gateway used in dev. CreateIntent returns a
// deterministic stub URL; VerifyWebhook returns the JSON body as-is.
package mock

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	"github.com/quangdangfit/goticket/internal/payment"
	"github.com/quangdangfit/goticket/internal/payment/dto"
)

type mockGateway struct{}

// New returns a mock payment.Gateway.
func New() payment.Gateway { return &mockGateway{} }

func (mockGateway) Name() string { return "mock" }

func (mockGateway) CreateIntent(_ context.Context, in dto.IntentInput) (*dto.Intent, error) {
	id := strings.ReplaceAll(uuid.NewString(), "-", "")
	return &dto.Intent{IntentID: id, URL: "https://example.invalid/pay/" + id}, nil
}

type webhookBody struct {
	EventID  string `json:"event_id"`
	IntentID string `json:"intent_id"`
	Type     string `json:"type"`
}

func (mockGateway) VerifyWebhook(_ map[string]string, body []byte) (*dto.WebhookEvent, error) {
	var wb webhookBody
	if err := json.Unmarshal(body, &wb); err != nil {
		return nil, err
	}
	return &dto.WebhookEvent{
		Provider: "mock", EventID: wb.EventID, IntentID: wb.IntentID,
		Type: wb.Type, Raw: body,
	}, nil
}
