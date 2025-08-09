## Rate Limiter (Go)

A configurable rate limiter middleware for Go HTTP servers with Redis persistence and strategy to switch storage backends.

### Features
- Limit by IP or Access Token; token overrides IP when present
- Config via env or `.env` (PORT, limits, Redis)
- Block window after exceed (HTTP 429, Retry-After)
- Redis-backed with pluggable storage interface
- Docker Compose with Redis

### Run locally
```bash
cd rate-limiter
go run ./cmd/server
```

Environment variables (defaults):
- `PORT` (8080)
- `RATE_LIMIT_MODE`=auto|ip|token (auto)
- `RATE_LIMIT_RPS` (10)
- `RATE_LIMIT_BLOCK_SECONDS` (300)
- `RATE_LIMIT_TOKEN_HEADER` (API_KEY)
- `RATE_LIMIT_TOKEN_OVERRIDES` example: `abc123:100:60,free:20:15`
- `REDIS_ADDR` (localhost:6379) in Docker use `redis:6379`
- `REDIS_DB` (0)
- `REDIS_PASSWORD` (empty)

### Docker
```bash
docker compose up --build
```

### Usage
- IP limit: send requests normally
- Token limit: include header `API_KEY: <TOKEN>`

When exceeded: 429 with body message and `Retry-After` seconds.

### Replace storage strategy
Implement `internal/storage.CounterStore` and wire it in `cmd/server/main.go`.

### Tests
```bash
go test ./...
```


