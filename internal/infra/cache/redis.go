package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const minTTL = time.Second

// RedisConfig contains Redis client options used by cache.
type RedisConfig struct {
	Addr        string
	Password    string
	DB          int
	DialTimeout time.Duration
}

// RedisCache implements Cache using go-redis.
type RedisCache struct {
	client *redis.Client

	mu     sync.RWMutex
	closed bool
}

// NewRedis creates a Redis-backed cache and validates the connection.
func NewRedis(cfg RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:        cfg.Addr,
		Password:    cfg.Password,
		DB:          cfg.DB,
		DialTimeout: cfg.DialTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := r.ensureOpen(); err != nil {
		return nil, err
	}

	raw, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}

	var value []byte
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := r.ensureOpen(); err != nil {
		return err
	}

	if ttl < minTTL {
		ttl = minTTL
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, raw, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if err := r.ensureOpen(); err != nil {
		return err
	}

	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := r.ensureOpen(); err != nil {
		return false, err
	}

	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

func (r *RedisCache) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	return r.client.Close()
}

func (r *RedisCache) ensureOpen() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return errors.New("redis cache is closed")
	}

	return nil
}
