package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Mode string

const (
	ModeIP    Mode = "ip"
	ModeToken Mode = "token"
	ModeAuto  Mode = "auto"
)

type TokenOverride struct {
	LimitPerSecond  int64
	BlockForSeconds int64
}

type Config struct {
	Port                string
	Mode                Mode
	DefaultLimitPerSec  int64
	DefaultBlockSeconds int64
	TokenHeader         string

	RedisAddr     string
	RedisDB       int
	RedisPassword string

	TokenOverrides map[string]TokenOverride
}

func Load() (*Config, error) {
	// Load .env if present
	_ = godotenv.Load()

	cfg := &Config{
		Port:                getString("PORT", "8080"),
		Mode:                Mode(getString("RATE_LIMIT_MODE", string(ModeAuto))),
		DefaultLimitPerSec:  getInt64("RATE_LIMIT_RPS", 10),
		DefaultBlockSeconds: getInt64("RATE_LIMIT_BLOCK_SECONDS", 300),
		TokenHeader:         getString("RATE_LIMIT_TOKEN_HEADER", "API_KEY"),

		RedisAddr:     getString("REDIS_ADDR", "localhost:6379"),
		RedisDB:       int(getInt64("REDIS_DB", 0)),
		RedisPassword: getString("REDIS_PASSWORD", ""),

		TokenOverrides: map[string]TokenOverride{},
	}

	if err := parseTokenOverrides(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func parseTokenOverrides(cfg *Config) error {
	raw := strings.TrimSpace(os.Getenv("RATE_LIMIT_TOKEN_OVERRIDES"))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		fields := strings.Split(p, ":")
		if len(fields) != 3 {
			return fmt.Errorf("invalid RATE_LIMIT_TOKEN_OVERRIDES item: %s", p)
		}
		token := strings.TrimSpace(fields[0])
		lim, err := strconv.ParseInt(strings.TrimSpace(fields[1]), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid limit in RATE_LIMIT_TOKEN_OVERRIDES '%s': %w", p, err)
		}
		blk, err := strconv.ParseInt(strings.TrimSpace(fields[2]), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid block seconds in RATE_LIMIT_TOKEN_OVERRIDES '%s': %w", p, err)
		}
		cfg.TokenOverrides[token] = TokenOverride{LimitPerSecond: lim, BlockForSeconds: blk}
	}
	return nil
}

func getString(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func getInt64(key string, def int64) int64 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return n
		}
	}
	return def
}
