// Package service implements the inventory.Inventory port using Redis Lua
// scripts. The service is the sole writer of `inv:*` and `hold:*` keys.
package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/quangdangfit/goticket/internal/inventory"
	"github.com/quangdangfit/goticket/internal/inventory/dto"
	apperr "github.com/quangdangfit/goticket/pkg/errors"
)

//go:embed lua/hold.lua
var holdScript string

//go:embed lua/release.lua
var releaseScript string

//go:embed lua/confirm.lua
var confirmScript string

type redisInventory struct {
	c    *redis.Client
	hold *redis.Script
	rel  *redis.Script
	conf *redis.Script
}

// New constructs an inventory.Inventory backed by Redis.
func New(c *redis.Client) inventory.Inventory {
	return &redisInventory{
		c:    c,
		hold: redis.NewScript(holdScript),
		rel:  redis.NewScript(releaseScript),
		conf: redis.NewScript(confirmScript),
	}
}

func invKey(showtime, ticketType string) string {
	return "inv:" + showtime + ":" + ticketType
}

func holdKey(id string) string { return "hold:" + id }

func (r *redisInventory) Warm(ctx context.Context, showtimeID string, specs []dto.QuotaSpec) error {
	pipe := r.c.Pipeline()
	for _, s := range specs {
		pipe.Set(ctx, invKey(showtimeID, s.TicketTypeID), s.Total, 0)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("warm inventory: %w", err)
	}
	return nil
}

func (r *redisInventory) Hold(ctx context.Context, in dto.HoldInput) (*dto.Hold, error) {
	if in.TTL <= 0 {
		in.TTL = 10 * time.Minute
	}
	id := strings.ReplaceAll(uuid.NewString(), "-", "")[:26]
	payload, err := json.Marshal(in.Items)
	if err != nil {
		return nil, fmt.Errorf("encode items: %w", err)
	}
	args := []any{id, in.UserID, int(in.TTL.Seconds()), string(payload)}
	for _, it := range in.Items {
		args = append(args, invKey(it.ShowtimeID, it.TicketTypeID), it.Quantity, "")
	}
	res, err := r.hold.Run(ctx, r.c, []string{holdKey(id)}, args...).Result()
	if err != nil {
		return nil, fmt.Errorf("hold script: %w", err)
	}
	if s, ok := res.(string); ok && strings.HasPrefix(s, "SOLD_OUT") {
		return nil, apperr.ErrSoldOut
	}
	return &dto.Hold{
		ID:        id,
		UserID:    in.UserID,
		Status:    "held",
		Items:     in.Items,
		ExpiresAt: time.Now().UTC().Add(in.TTL),
	}, nil
}

func (r *redisInventory) Release(ctx context.Context, holdID string) error {
	h, err := r.Get(ctx, holdID)
	if err != nil {
		return err
	}
	if h.Status != "held" {
		return nil // idempotent
	}
	args := make([]any, 0, len(h.Items)*2)
	for _, it := range h.Items {
		args = append(args, invKey(it.ShowtimeID, it.TicketTypeID), it.Quantity)
	}
	if _, err := r.rel.Run(ctx, r.c, []string{holdKey(holdID)}, args...).Result(); err != nil {
		return fmt.Errorf("release script: %w", err)
	}
	return nil
}

func (r *redisInventory) Confirm(ctx context.Context, holdID string) error {
	res, err := r.conf.Run(ctx, r.c, []string{holdKey(holdID)}).Result()
	if err != nil {
		return fmt.Errorf("confirm script: %w", err)
	}
	n, _ := res.(int64)
	switch n {
	case 1, 0: // ok / already confirmed
		return nil
	case -1:
		return apperr.ErrNotFound
	default:
		return fmt.Errorf("confirm: %w", apperr.ErrConflict)
	}
}

func (r *redisInventory) Get(ctx context.Context, holdID string) (*dto.Hold, error) {
	m, err := r.c.HGetAll(ctx, holdKey(holdID)).Result()
	if err != nil || len(m) == 0 {
		if errors.Is(err, redis.Nil) || len(m) == 0 {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("get hold: %w", err)
	}
	h := &dto.Hold{ID: m["id"], UserID: m["user_id"], Status: m["status"]}
	if p := m["payload"]; p != "" {
		_ = json.Unmarshal([]byte(p), &h.Items)
	}
	return h, nil
}

func (r *redisInventory) Available(ctx context.Context, showtimeID, ticketTypeID string) (int, error) {
	v, err := r.c.Get(ctx, invKey(showtimeID, ticketTypeID)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("available: %w", err)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("available parse: %w", err)
	}
	return n, nil
}
