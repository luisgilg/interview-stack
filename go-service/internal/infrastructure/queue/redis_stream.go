package queue

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// RedisStreamQueue implements the WriteQueue interface using Redis Streams.
type RedisStreamQueue struct {
	client *redis.Client
	stream string
}

// NewRedisStreamQueue builds a new producer using the provided redis client and stream name.
func NewRedisStreamQueue(client *redis.Client, stream string) *RedisStreamQueue {
	return &RedisStreamQueue{
		client: client,
		stream: stream,
	}
}

// Enqueue serializes the event as JSON and appends it to the stream.
func (q *RedisStreamQueue) Enqueue(ctx context.Context, event domain.WriteEvent) error {
	if q == nil || q.client == nil {
		return errors.New("write queue not configured")
	}
	raw, err := json.Marshal(event)
	if err != nil {
		return err
	}
	args := &redis.XAddArgs{
		Stream: q.stream,
		Values: map[string]any{
			"event": raw,
		},
	}
	return q.client.XAdd(ctx, args).Err()
}
