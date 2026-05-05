// Package inventory is the hot-path bounded context that owns the
// "is this still available?" answer during the on-sale window.
package inventory

import (
	"context"

	"github.com/quangdangfit/goticket/internal/inventory/dto"
)

// Inventory is the service-facing port.
type Inventory interface {
	Warm(ctx context.Context, showtimeID string, specs []dto.QuotaSpec) error
	Hold(ctx context.Context, in dto.HoldInput) (*dto.Hold, error)
	Release(ctx context.Context, holdID string) error
	Confirm(ctx context.Context, holdID string) error
	Get(ctx context.Context, holdID string) (*dto.Hold, error)
	Available(ctx context.Context, showtimeID, ticketTypeID string) (int, error)
}
