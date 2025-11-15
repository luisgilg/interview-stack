package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	obsmetrics "github.com/example/interview-stack/go-service/internal/observability/metrics"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// ErrCacheMiss indicates that the key was not present in the cache.
var ErrCacheMiss = errors.New("cache miss")

// Client abstracts the cache backend implementation.
type Client interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Config controls how the cache service behaves.
type Config struct {
	Enabled    bool
	DefaultTTL time.Duration
	StaleTTL   time.Duration
}

// Status captures the freshness of the returned value.
type Status string

const (
	StatusBypass Status = "bypass"
	StatusMiss   Status = "miss"
	StatusFresh  Status = "fresh"
	StatusStale  Status = "stale"
)

// Metadata describes cache behaviour for an operation.
type Metadata struct {
	Status Status
}

type cacheEnvelope struct {
	Payload json.RawMessage `json:"payload"`
	Stored  time.Time       `json:"stored_at"`
}

// Service offers cache-aside helpers with SWR semantics.
type Service struct {
	client  Client
	enabled bool
	ttl     time.Duration
	stale   time.Duration
	logger  domain.Logger
	clock   domain.Clock
}

// NewService builds a cache service around the provided client.
func NewService(client Client, cfg Config, logger domain.Logger, clock domain.Clock) *Service {
	service := &Service{
		client:  client,
		enabled: cfg.Enabled && client != nil && cfg.DefaultTTL > 0,
		ttl:     cfg.DefaultTTL,
		stale:   cfg.StaleTTL,
		logger:  logger,
		clock:   clock,
	}
	if service.stale < 0 {
		service.stale = 0
	}
	return service
}

// Enabled reports whether the cache is active.
func (s *Service) Enabled() bool {
	return s != nil && s.enabled
}

// Store writes a value to the cache immediately.
func (s *Service) Store(ctx context.Context, key string, value any) error {
	return persist(s, ctx, key, value)
}

// Delete removes a key from the cache.
func (s *Service) Delete(ctx context.Context, key string) error {
	if s == nil || s.client == nil {
		return nil
	}
	return s.client.Delete(ctx, key)
}

// Fetch executes the loader and stores/retrieves values while respecting SWR semantics.
func Fetch[T any](svc *Service, ctx context.Context, key string, loader func(context.Context) (T, error)) (result T, meta Metadata, err error) {
	meta = Metadata{Status: StatusBypass}
	defer func() {
		obsmetrics.RecordCacheStatus(string(meta.Status))
	}()

	if loader == nil {
		err = errors.New("loader is required")
		return
	}
	if svc == nil || !svc.enabled || svc.client == nil || key == "" {
		result, err = loader(ctx)
		return
	}

	var raw []byte
	raw, err = svc.client.Get(ctx, key)
	if err == nil {
		var storedAt time.Time
		result, storedAt, err = decodeValue[T](raw)
		if err != nil {
			svc.logger.Warn("cache decode failed", domain.KV("key", key), domain.KV("error", err.Error()))
			result, meta, err = loadAndStore(svc, ctx, key, loader, StatusMiss)
			return
		}
		age := svc.clock.Now().Sub(storedAt)
		if age <= svc.ttl {
			meta.Status = StatusFresh
			err = nil
			return
		}
		if age <= svc.ttl+svc.stale {
			triggerRefresh(svc, key, loader)
			meta.Status = StatusStale
			err = nil
			return
		}
		result, meta, err = loadAndStore(svc, ctx, key, loader, StatusMiss)
		return
	}
	if errors.Is(err, ErrCacheMiss) {
		result, meta, err = loadAndStore(svc, ctx, key, loader, StatusMiss)
		return
	}
	svc.logger.Warn("cache get failed", domain.KV("key", key), domain.KV("error", err.Error()))
	result, meta, err = loadWithStatus(svc, ctx, loader, StatusBypass)
	return
}

func loadAndStore[T any](s *Service, ctx context.Context, key string, loader func(context.Context) (T, error), status Status) (T, Metadata, error) {
	val, err := loader(ctx)
	if err != nil {
		return val, Metadata{Status: status}, err
	}
	if err := persist(s, ctx, key, val); err != nil {
		s.logger.Warn("cache store failed", domain.KV("key", key), domain.KV("error", err.Error()))
	}
	return val, Metadata{Status: status}, nil
}

func loadWithStatus[T any](_ *Service, ctx context.Context, loader func(context.Context) (T, error), status Status) (T, Metadata, error) {
	val, err := loader(ctx)
	return val, Metadata{Status: status}, err
}

func persist[T any](s *Service, ctx context.Context, key string, value T) error {
	if s == nil || s.client == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	entry := cacheEnvelope{
		Payload: payload,
		Stored:  s.clock.Now(),
	}
	raw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	ttl := s.ttl + s.stale
	if ttl <= 0 {
		ttl = s.ttl
	}
	if ttl <= 0 {
		return nil
	}
	return s.client.Set(ctx, key, raw, ttl)
}

func decodeValue[T any](raw []byte) (T, time.Time, error) {
	var zero T
	var entry cacheEnvelope
	if err := json.Unmarshal(raw, &entry); err != nil {
		return zero, time.Time{}, err
	}
	if len(entry.Payload) == 0 {
		return zero, entry.Stored, errors.New("empty payload")
	}
	var out T
	if err := json.Unmarshal(entry.Payload, &out); err != nil {
		return zero, entry.Stored, err
	}
	return out, entry.Stored, nil
}

func triggerRefresh[T any](s *Service, key string, loader func(context.Context) (T, error)) {
	if s == nil || s.client == nil {
		return
	}
	go func() {
		ctx := context.Background()
		val, err := loader(ctx)
		if err != nil {
			s.logger.Warn("cache refresh failed", domain.KV("key", key), domain.KV("error", err.Error()))
			return
		}
		if err := persist(s, ctx, key, val); err != nil {
			s.logger.Warn("cache refresh store failed", domain.KV("key", key), domain.KV("error", err.Error()))
		}
	}()
}
