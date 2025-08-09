package limiter

import (
	"context"
	"testing"
	"time"

	redispkg "rate-limiter/internal/storage/redis"

	miniredis "github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
)

func newTestLimiter(t *testing.T) (*Limiter, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	rdb := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := redispkg.New(rdb)
	lim := New(store)
	return lim, func() { mr.Close() }
}

func TestLimiter_AllowsUnderLimit(t *testing.T) {
	lim, cleanup := newTestLimiter(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0)
	for i := 0; i < 5; i++ {
		res, err := lim.Check(ctx, "ip:1.2.3.4", 5, 10*time.Second, now)
		if err != nil {
			t.Fatalf("check err: %v", err)
		}
		if !res.Allowed {
			t.Fatalf("unexpected deny at i=%d", i)
		}
	}
}

func TestLimiter_BlocksOnExceedAndSetsBlock(t *testing.T) {
	lim, cleanup := newTestLimiter(t)
	defer cleanup()
	ctx := context.Background()
	now := time.Unix(1_700_000_000, 0)

	// limit 2 rps, block 5s
	for i := 0; i < 2; i++ {
		res, err := lim.Check(ctx, "token:abc", 2, 5*time.Second, now)
		if err != nil || !res.Allowed {
			t.Fatalf("warmup err: %v allowed=%v", err, res.Allowed)
		}
	}
	res, err := lim.Check(ctx, "token:abc", 2, 5*time.Second, now)
	if err != nil {
		t.Fatalf("check err: %v", err)
	}
	if res.Allowed {
		t.Fatalf("expected deny on exceed")
	}

	// still blocked next second during block window
	res, err = lim.Check(ctx, "token:abc", 2, 5*time.Second, now.Add(1*time.Second))
	if err != nil {
		t.Fatalf("check err: %v", err)
	}
	if res.Allowed {
		t.Fatalf("expected blocked due to SetBlock")
	}
}
