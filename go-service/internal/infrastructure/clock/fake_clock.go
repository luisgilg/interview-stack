package clock

import (
	"sync"
	"time"
)

// FakeClock returns a deterministic time for tests.
type FakeClock struct {
	mu  sync.RWMutex
	now time.Time
}

func NewFakeClock(initial time.Time) *FakeClock {
	return &FakeClock{now: initial}
}

func (f *FakeClock) Now() time.Time {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.now
}

// Set advances the fake clock to the provided instant.
func (f *FakeClock) Set(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = t
}
