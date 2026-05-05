// Package idempotency provides a fast pre-check (Redis SETNX) layered above
// a durable DB UNIQUE-key fallback. Bloom filter is omitted in this phase
// (Redis NX is sub-ms already); add it if Redis becomes a bottleneck.
package idempotency

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrReplay signals a duplicate key whose first run already returned
// successfully. Caller should look up the prior result.
var ErrReplay = errors.New("idempotency: replay")

// Guard is the redis-backed first-line check. The DB UNIQUE index on
// (user_id, key) is the durable second line of defence.
type Guard interface {
	Reserve(ctx context.Context, userID, key string, ttl time.Duration) error
	Release(ctx context.Context, userID, key string) error
}

type redisGuard struct{ c *redis.Client }

// New returns a Redis-backed Guard.
func New(c *redis.Client) Guard { return &redisGuard{c: c} }

func (r *redisGuard) keyOf(uid, k string) string { return "idem:" + uid + ":" + k }

func (r *redisGuard) Reserve(ctx context.Context, uid, k string, ttl time.Duration) error {
	if k == "" {
		return nil
	}
	ok, err := r.c.SetNX(ctx, r.keyOf(uid, k), "1", ttl).Result()
	if err != nil {
		return fmt.Errorf("idempotency setnx: %w", err)
	}
	if !ok {
		return ErrReplay
	}
	return nil
}

func (r *redisGuard) Release(ctx context.Context, uid, k string) error {
	if k == "" {
		return nil
	}
	return r.c.Del(ctx, r.keyOf(uid, k)).Err()
}
