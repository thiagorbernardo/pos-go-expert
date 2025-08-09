package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"rate-limiter/internal/config"
	"rate-limiter/internal/limiter"
)

type fakeChecker struct {
	allow bool
}

func (f fakeChecker) Check(_ context.Context, _ string, _ int64, _ time.Duration, _ time.Time) (limiter.Result, error) {
	if f.allow {
		return limiter.Result{Allowed: true}, nil
	}
	return limiter.Result{Allowed: false, RetryAfter: 3 * time.Second}, nil
}

func TestMiddleware_DeniesWith429(t *testing.T) {
	cfg := &config.Config{Port: "8080", Mode: config.ModeIP, DefaultLimitPerSec: 1, DefaultBlockSeconds: 5, TokenHeader: "API_KEY"}
	mw := NewRateLimitMiddleware(fakeChecker{allow: false}, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rr.Code)
	}
	if rr.Header().Get("Retry-After") == "" {
		t.Fatalf("expected Retry-After header")
	}
}

func TestMiddleware_AllowsNext(t *testing.T) {
	cfg := &config.Config{Port: "8080", Mode: config.ModeIP, DefaultLimitPerSec: 1, DefaultBlockSeconds: 5, TokenHeader: "API_KEY"}
	mw := NewRateLimitMiddleware(fakeChecker{allow: true}, cfg)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mw.Handler(next).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
