package storage

import (
	"context"
	"time"
)

// CounterStore abstracts persistence for rate limiting state.
type CounterStore interface {
	// Incr increases counter in the current window; when first increment, it sets window TTL.
	Incr(ctx context.Context, key string, window time.Duration) (count int64, err error)

	// SetBlock marks an identifier as blocked for a certain duration.
	SetBlock(ctx context.Context, id string, blockFor time.Duration) error

	// IsBlocked returns whether id is currently blocked and the remaining TTL.
	IsBlocked(ctx context.Context, id string) (blocked bool, ttl time.Duration, err error)
}
