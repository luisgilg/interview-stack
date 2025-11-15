package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	obsmetrics "github.com/example/interview-stack/go-service/internal/observability/metrics"

	"github.com/example/interview-stack/go-service/internal/domain"
	"github.com/example/interview-stack/go-service/internal/infrastructure/repository"
)

var errSkipEvent = errors.New("skip event")

// Worker consumes Redis Stream messages and applies them to the configured store.
type Worker struct {
	client      *redis.Client
	stream      string
	group       string
	consumer    string
	source      string
	batchSize   int64
	block       time.Duration
	store       repository.ProductStore
	logger      domain.Logger
	stopBackoff time.Duration
}

// NewWorker builds a worker for the provided consumer group.
func NewWorker(
	client *redis.Client,
	stream string,
	group string,
	consumer string,
	source string,
	batchSize int,
	block time.Duration,
	store repository.ProductStore,
	logger domain.Logger,
) *Worker {
	return &Worker{
		client:      client,
		stream:      stream,
		group:       group,
		consumer:    consumer,
		source:      source,
		batchSize:   int64(batchSize),
		block:       block,
		store:       store,
		logger:      logger,
		stopBackoff: time.Second,
	}
}

// Run starts the consume loop until the context is cancelled.
func (w *Worker) Run(ctx context.Context) {
	if w == nil || w.client == nil {
		return
	}
	w.ensureGroup(ctx)
	for {
		if ctx.Err() != nil {
			return
		}
		streams, err := w.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    w.group,
			Consumer: w.consumer,
			Streams:  []string{w.stream, ">"},
			Count:    w.batchSize,
			Block:    w.block,
		}).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			if ctx.Err() != nil {
				return
			}
			w.logger.Warn("write-behind read failed", domain.KV("error", err.Error()))
			obsmetrics.RecordWriteBehindError()
			time.Sleep(w.stopBackoff)
			continue
		}
		batchStart := time.Now()
		batchSize := 0
		maxLag := float64(-1)
		now := time.Now()
		for _, stream := range streams {
			for _, msg := range stream.Messages {
				batchSize++
				if lag := calculateLagSeconds(now, msg.ID); lag > maxLag {
					maxLag = lag
				}
				if err := w.processMessage(ctx, msg); err != nil {
					if errors.Is(err, errSkipEvent) {
						w.ack(ctx, msg.ID)
						continue
					}
					w.logger.Warn("write-behind apply failed", domain.KV("error", err.Error()))
					obsmetrics.RecordWriteBehindError()
					continue
				}
				w.ack(ctx, msg.ID)
			}
		}
		w.recordBatchMetrics(ctx, batchSize, maxLag, time.Since(batchStart))
	}
}

func (w *Worker) ensureGroup(ctx context.Context) {
	if w == nil || w.client == nil {
		return
	}
	err := w.client.XGroupCreateMkStream(ctx, w.stream, w.group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		w.logger.Warn("failed to create redis stream group", domain.KV("error", err.Error()))
	}
}

func (w *Worker) processMessage(ctx context.Context, msg redis.XMessage) error {
	rawValue, ok := msg.Values["event"]
	if !ok {
		return fmt.Errorf("event field missing")
	}
	var payload []byte
	switch v := rawValue.(type) {
	case string:
		payload = []byte(v)
	case []byte:
		payload = v
	default:
		return fmt.Errorf("unexpected event type %T", rawValue)
	}
	var event domain.WriteEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode event: %w", err)
	}
	if event.Source != "" && event.Source != w.source {
		return errSkipEvent
	}
	switch event.Type {
	case domain.WriteEventCreate:
		return w.applyCreate(ctx, event)
	case domain.WriteEventUpdate:
		return w.applyUpdate(ctx, event)
	case domain.WriteEventDelete:
		return w.applyDelete(ctx, event)
	default:
		return fmt.Errorf("unknown event type %s", event.Type)
	}
}

func (w *Worker) ack(ctx context.Context, id string) {
	if err := w.client.XAck(ctx, w.stream, w.group, id).Err(); err != nil {
		w.logger.Warn("failed to ack message", domain.KV("id", id), domain.KV("error", err.Error()))
		obsmetrics.RecordWriteBehindError()
	}
}

func (w *Worker) applyCreate(ctx context.Context, event domain.WriteEvent) error {
	if event.Payload == nil {
		return errors.New("missing payload for create")
	}
	_, err := w.store.CreateProduct(ctx, *event.Payload)
	return err
}

func (w *Worker) applyUpdate(ctx context.Context, event domain.WriteEvent) error {
	if event.Payload == nil {
		return errors.New("missing payload for update")
	}
	_, err := w.store.UpdateProduct(ctx, event.ID, *event.Payload)
	return err
}

func (w *Worker) applyDelete(ctx context.Context, event domain.WriteEvent) error {
	_, err := w.store.DeleteProduct(ctx, event.ID)
	return err
}

func (w *Worker) recordBatchMetrics(ctx context.Context, size int, lagSeconds float64, duration time.Duration) {
	if size == 0 {
		return
	}
	queueLength, err := w.client.XLen(ctx, w.stream).Result()
	if err != nil {
		w.logger.Warn("failed to read queue length", domain.KV("error", err.Error()))
		obsmetrics.RecordWriteBehindError()
		queueLength = -1
	}
	obsmetrics.RecordWriteBehindBatch(size, duration, lagSeconds, queueLength)
}

func calculateLagSeconds(now time.Time, id string) float64 {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) == 0 {
		return -1
	}
	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return -1
	}
	lag := now.Sub(time.UnixMilli(ms))
	if lag < 0 {
		return 0
	}
	return lag.Seconds()
}
