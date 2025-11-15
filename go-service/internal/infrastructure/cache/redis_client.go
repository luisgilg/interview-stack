package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	appcache "github.com/example/interview-stack/go-service/internal/application/cache"
	"github.com/example/interview-stack/go-service/internal/config"
)

// RedisClient implements the cache.Client interface using go-redis.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient builds a client against the provided configuration.
func NewRedisClient(cfg config.RedisConfig) *RedisClient {
	if cfg.Host == "" || cfg.Port <= 0 {
		return nil
	}
	options := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	}
	return &RedisClient{client: redis.NewClient(options)}
}

// Get retrieves the raw cache payload.
func (c *RedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	if c == nil || c.client == nil {
		return nil, appcache.ErrCacheMiss
	}
	cmd := c.client.Get(ctx, key)
	bytes, err := cmd.Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, appcache.ErrCacheMiss
		}
		return nil, err
	}
	return bytes, nil
}

// Set writes a payload with the provided TTL.
func (c *RedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return errors.New("redis client not configured")
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a cached entry.
func (c *RedisClient) Delete(ctx context.Context, key string) error {
	if c == nil || c.client == nil {
		return errors.New("redis client not configured")
	}
	return c.client.Del(ctx, key).Err()
}

// Ping validates connectivity with Redis.
func (c *RedisClient) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return errors.New("redis client not configured")
	}
	return c.client.Ping(ctx).Err()
}

// Close releases the underlying Redis connection.
func (c *RedisClient) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}
