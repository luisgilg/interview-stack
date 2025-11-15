package clock

import (
	"time"

	"github.com/example/interview-stack/go-service/internal/domain"
)

// SystemClock returns the real system time and is used in production.
type SystemClock struct{}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

var _ domain.Clock = (*SystemClock)(nil)
