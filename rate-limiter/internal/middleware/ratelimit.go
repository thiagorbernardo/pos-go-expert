package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"rate-limiter/internal/config"
	"rate-limiter/internal/limiter"
)

type RateLimitMiddleware struct {
	limiter limiter.Checker
	cfg     *config.Config
}

func NewRateLimitMiddleware(l limiter.Checker, cfg *config.Config) *RateLimitMiddleware {
	return &RateLimitMiddleware{limiter: l, cfg: cfg}
}

func (m *RateLimitMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier, limit, blockFor := m.resolveRule(r)

		res, err := m.limiter.Check(r.Context(), identifier, limit, time.Duration(blockFor)*time.Second, time.Now())
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !res.Allowed {
			w.Header().Set("Content-Type", "application/json")
			if res.RetryAfter > 0 {
				w.Header().Set("Retry-After", formatRetryAfter(res.RetryAfter))
			}
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"you have reached the maximum number of requests or actions allowed within a certain time frame"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *RateLimitMiddleware) resolveRule(r *http.Request) (identifier string, limit int64, blockSeconds int64) {
	headerToken := strings.TrimSpace(r.Header.Get(m.cfg.TokenHeader))
	ip := clientIP(r)

	// Auto mode: token overrides IP if present and configured; token mode: require token; ip mode: ignore token
	switch m.cfg.Mode {
	case config.ModeToken:
		identifier = safeIdentifier("token:" + headerToken)
		if ov, ok := m.cfg.TokenOverrides[headerToken]; ok {
			return identifier, ov.LimitPerSecond, ov.BlockForSeconds
		}
		return identifier, m.cfg.DefaultLimitPerSec, m.cfg.DefaultBlockSeconds
	case config.ModeIP:
		identifier = safeIdentifier("ip:" + ip)
		return identifier, m.cfg.DefaultLimitPerSec, m.cfg.DefaultBlockSeconds
	default: // auto
		if headerToken != "" {
			identifier = safeIdentifier("token:" + headerToken)
			if ov, ok := m.cfg.TokenOverrides[headerToken]; ok {
				return identifier, ov.LimitPerSecond, ov.BlockForSeconds
			}
			// Token present but no override: use defaults, token takes precedence over IP
			return identifier, m.cfg.DefaultLimitPerSec, m.cfg.DefaultBlockSeconds
		}
		identifier = safeIdentifier("ip:" + ip)
		return identifier, m.cfg.DefaultLimitPerSec, m.cfg.DefaultBlockSeconds
	}
}

func clientIP(r *http.Request) string {
	// Try X-Forwarded-For first, then RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// first value is client
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func safeIdentifier(s string) string {
	// prevent spaces and illegal characters in redis keys
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "\n", "_")
	s = strings.ReplaceAll(s, "\r", "_")
	return s
}

func formatRetryAfter(d time.Duration) string {
	// Round up to seconds
	secs := int(d.Round(time.Second) / time.Second)
	if secs < 1 {
		secs = 1
	}
	return strconvItoa(secs)
}

func strconvItoa(i int) string {
	// local small helper to avoid importing strconv in this file
	const digits = "0123456789"
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	buf := make([]byte, 0, 12)
	for i > 0 {
		buf = append([]byte{digits[i%10]}, buf...)
		i /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
