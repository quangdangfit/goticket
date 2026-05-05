package repository

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/quangdangfit/goticket/internal/event"
	"github.com/quangdangfit/goticket/internal/event/dto"
)

type redisCache struct {
	c   *redis.Client
	ttl time.Duration
}

// NewCache returns a Redis-backed event.Cache with the given base TTL plus
// per-key jitter (±10%) to avoid synchronized expiries.
func NewCache(c *redis.Client, ttl time.Duration) event.Cache {
	return &redisCache{c: c, ttl: ttl}
}

func (r *redisCache) key(id string) string { return "event:" + id }

func (r *redisCache) GetEvent(ctx context.Context, id string) (*dto.Event, bool) {
	b, err := r.c.Get(ctx, r.key(id)).Bytes()
	if err != nil {
		return nil, false
	}
	var e dto.Event
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, false
	}
	return &e, true
}

func (r *redisCache) SetEvent(ctx context.Context, id string, e *dto.Event) {
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	jitter := time.Duration(rand.Int63n(int64(r.ttl) / 5)) //nolint:gosec // jitter only
	_ = r.c.Set(ctx, r.key(id), b, r.ttl-r.ttl/10+jitter).Err()
}

func (r *redisCache) Invalidate(ctx context.Context, id string) {
	_ = r.c.Del(ctx, r.key(id)).Err()
}
