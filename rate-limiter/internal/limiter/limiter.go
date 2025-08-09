package limiter

import (
	"context"
	"fmt"
	"time"

	"rate-limiter/internal/storage"
)

type Result struct {
	Allowed    bool
	RetryAfter time.Duration
}

// Checker is the contract used by the middleware to evaluate rate limits.
type Checker interface {
	Check(ctx context.Context, identifier string, limitPerSecond int64, blockFor time.Duration, now time.Time) (Result, error)
}

type Limiter struct {
	store storage.CounterStore
}

func New(store storage.CounterStore) *Limiter {
	return &Limiter{store: store}
}

// Check increases the counter for the identifier within a 1s window and decides allow/deny.
func (l *Limiter) Check(ctx context.Context, identifier string, limitPerSecond int64, blockFor time.Duration, now time.Time) (Result, error) {
	if limitPerSecond <= 0 {
		return Result{Allowed: true}, nil
	}

	// First, check blocked state
	blocked, ttl, err := l.store.IsBlocked(ctx, identifier)
	if err != nil {
		return Result{}, err
	}
	if blocked {
		return Result{Allowed: false, RetryAfter: ttl}, nil
	}

	// Window key by epoch second
	sec := now.Unix()
	key := windowKey(identifier, sec)
	count, err := l.store.Incr(ctx, key, time.Second)
	if err != nil {
		return Result{}, err
	}
	if count > limitPerSecond {
		// Exceeded. Block further requests for blockFor duration.
		if blockFor > 0 {
			if err := l.store.SetBlock(ctx, identifier, blockFor); err != nil {
				return Result{}, err
			}
		}
		return Result{Allowed: false, RetryAfter: blockFor}, nil
	}
	return Result{Allowed: true}, nil
}

func windowKey(identifier string, epochSec int64) string {
	return fmt.Sprintf("rl:cnt:%s:%d", identifier, epochSec)
}
