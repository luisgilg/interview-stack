package cache_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	appcache "github.com/example/interview-stack/go-service/internal/application/cache"
)

type fakeClock struct {
	current time.Time
}

func (c *fakeClock) Now() time.Time {
	return c.current
}

func (c *fakeClock) Set(t time.Time) {
	c.current = t
}

type fakeCacheClient struct {
	mu   sync.Mutex
	data map[string][]byte
}

func newFakeCacheClient() *fakeCacheClient {
	return &fakeCacheClient{data: make(map[string][]byte)}
}

func (f *fakeCacheClient) Get(_ context.Context, key string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	value, ok := f.data[key]
	if !ok {
		return nil, appcache.ErrCacheMiss
	}
	out := make([]byte, len(value))
	copy(out, value)
	return out, nil
}

func (f *fakeCacheClient) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cloned := make([]byte, len(value))
	copy(cloned, value)
	f.data[key] = cloned
	return nil
}

func (f *fakeCacheClient) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	return nil
}

type testPayload struct {
	Items []string `json:"items"`
}

func encodeEnvelope(t *testing.T, payload any, storedAt time.Time) []byte {
	t.Helper()
	entry := struct {
		Payload any       `json:"payload"`
		Stored  time.Time `json:"stored_at"`
	}{
		Payload: payload,
		Stored:  storedAt,
	}
	raw, err := json.Marshal(entry)
	require.NoError(t, err)
	return raw
}

func TestServiceReturnsBypassWhenClientMissing(t *testing.T) {
	clock := &fakeClock{current: time.Now()}
	service := appcache.NewService(nil, appcache.Config{
		Enabled:    true,
		DefaultTTL: time.Minute,
		StaleTTL:   time.Minute,
	}, nil, clock)

	calls := 0
	result, meta, err := appcache.Fetch(service, context.Background(), "key", func(context.Context) (testPayload, error) {
		calls++
		return testPayload{Items: []string{"a"}}, nil
	})
	require.NoError(t, err)
	require.Equal(t, appcache.StatusBypass, meta.Status)
	require.Equal(t, 1, calls)
	require.Equal(t, []string{"a"}, result.Items)
}

func TestServiceStoresValueOnMiss(t *testing.T) {
	client := newFakeCacheClient()
	clock := &fakeClock{current: time.Unix(0, 0)}
	service := appcache.NewService(client, appcache.Config{
		Enabled:    true,
		DefaultTTL: time.Minute,
		StaleTTL:   time.Minute,
	}, nil, clock)

	calls := 0
	result, meta, err := appcache.Fetch(service, context.Background(), "products:list", func(context.Context) (testPayload, error) {
		calls++
		return testPayload{Items: []string{"x", "y"}}, nil
	})
	require.NoError(t, err)
	require.Equal(t, appcache.StatusMiss, meta.Status)
	require.Equal(t, 1, calls)
	require.Equal(t, []string{"x", "y"}, result.Items)
	clock.Set(clock.current.Add(10 * time.Second))
	raw, getErr := client.Get(context.Background(), "products:list")
	require.NoError(t, getErr)
	var stored struct {
		Payload testPayload `json:"payload"`
	}
	require.NoError(t, json.Unmarshal(raw, &stored))
	require.Equal(t, []string{"x", "y"}, stored.Payload.Items)
}

func TestServiceReturnsFreshHit(t *testing.T) {
	client := newFakeCacheClient()
	now := time.Unix(100, 0)
	client.data["products:list"] = encodeEnvelope(t, testPayload{Items: []string{"cached"}}, now)
	clock := &fakeClock{current: now.Add(10 * time.Second)}
	service := appcache.NewService(client, appcache.Config{
		Enabled:    true,
		DefaultTTL: time.Minute,
		StaleTTL:   time.Minute,
	}, nil, clock)

	calls := 0
	_, meta, err := appcache.Fetch(service, context.Background(), "products:list", func(context.Context) (testPayload, error) {
		calls++
		return testPayload{}, nil
	})
	require.NoError(t, err)
	require.Equal(t, appcache.StatusFresh, meta.Status)
	require.Equal(t, 0, calls)
}

func TestServiceReturnsStaleAndRefreshes(t *testing.T) {
	client := newFakeCacheClient()
	now := time.Unix(200, 0)
	client.data["products:list"] = encodeEnvelope(t, testPayload{Items: []string{"stale"}}, now.Add(-70*time.Second))
	clock := &fakeClock{current: now}
	service := appcache.NewService(client, appcache.Config{
		Enabled:    true,
		DefaultTTL: time.Minute,
		StaleTTL:   time.Minute,
	}, nil, clock)

	refreshCh := make(chan struct{}, 1)
	loader := func(context.Context) (testPayload, error) {
		refreshCh <- struct{}{}
		return testPayload{Items: []string{"refreshed"}}, nil
	}

	result, meta, err := appcache.Fetch(service, context.Background(), "products:list", loader)
	require.NoError(t, err)
	require.Equal(t, appcache.StatusStale, meta.Status)
	require.Equal(t, []string{"stale"}, result.Items)
	select {
	case <-refreshCh:
		// refreshed successfully
	case <-time.After(time.Second):
		t.Fatal("expected refresh to be triggered")
	}
}
