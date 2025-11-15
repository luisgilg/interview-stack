package domain

import (
	"context"
	"time"
)

// WriteEventType enumerates supported write-behind operations.
type WriteEventType string

const (
	// WriteEventCreate indicates a new product should be stored.
	WriteEventCreate WriteEventType = "create"
	// WriteEventUpdate indicates an existing product should be updated.
	WriteEventUpdate WriteEventType = "update"
	// WriteEventDelete indicates an existing product should be removed.
	WriteEventDelete WriteEventType = "delete"
)

// WriteEvent represents the payload stored in the Redis stream.
type WriteEvent struct {
	Type      WriteEventType `json:"type"`
	ID        string         `json:"id"`
	Payload   *Product       `json:"payload,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Source    string         `json:"source"`
}

// WriteQueue abstracts the enqueueing of events to Redis Streams.
type WriteQueue interface {
	Enqueue(ctx context.Context, event WriteEvent) error
}
