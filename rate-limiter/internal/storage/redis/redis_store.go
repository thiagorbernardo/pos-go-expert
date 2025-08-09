package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Store struct {
	client *goredis.Client
}

func New(client *goredis.Client) *Store {
	return &Store{client: client}
}

func (s *Store) Incr(ctx context.Context, key string, window time.Duration) (int64, error) {
	// INCR and set TTL when new (value == 1)
	val, err := s.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if val == 1 {
		_ = s.client.Expire(ctx, key, window).Err()
	}
	return val, nil
}

func (s *Store) SetBlock(ctx context.Context, id string, blockFor time.Duration) error {
	key := blockKey(id)
	return s.client.Set(ctx, key, "1", blockFor).Err()
}

func (s *Store) IsBlocked(ctx context.Context, id string) (bool, time.Duration, error) {
	key := blockKey(id)
	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}
	if ttl > 0 {
		return true, ttl, nil
	}
	// When key missing, TTL returns -2; when no expire, -1
	return false, 0, nil
}

func blockKey(id string) string {
	return "rl:block:" + id
}
