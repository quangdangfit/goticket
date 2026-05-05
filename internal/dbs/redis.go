package dbs

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/quangdangfit/goticket/config"
)

// Redis is the narrow port that domains depend on.
type Redis interface {
	Client() *redis.Client
	Close() error
}

type redisClient struct{ c *redis.Client }

// NewRedis dials a Redis server and returns the wrapped client.
func NewRedis(ctx context.Context, cfg config.RedisConfig) (Redis, error) {
	c := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	if err := c.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &redisClient{c: c}, nil
}

func (r *redisClient) Client() *redis.Client { return r.c }
func (r *redisClient) Close() error          { return r.c.Close() }
