package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRedisCache(t *testing.T) {
	addr := requireRedisAddr(t)
	ctx := context.Background()

	tests := []struct {
		name string
		run  func(t *testing.T, rc *RedisCache)
	}{
		{
			name: "get roundtrip",
			run: func(t *testing.T, rc *RedisCache) {
				if err := rc.Set(ctx, "k:roundtrip", []byte("value"), time.Minute); err != nil {
					t.Fatalf("set failed: %v", err)
				}

				got, err := rc.Get(ctx, "k:roundtrip")
				if err != nil {
					t.Fatalf("get failed: %v", err)
				}
				if string(got) != "value" {
					t.Fatalf("unexpected value: %q", string(got))
				}
			},
		},
		{
			name: "get miss returns ErrCacheMiss",
			run: func(t *testing.T, rc *RedisCache) {
				_, err := rc.Get(ctx, "k:missing")
				if !errors.Is(err, ErrCacheMiss) {
					t.Fatalf("expected ErrCacheMiss, got %v", err)
				}
			},
		},
		{
			name: "ttl expiry",
			run: func(t *testing.T, rc *RedisCache) {
				if err := rc.Set(ctx, "k:ttl", []byte("value"), time.Second); err != nil {
					t.Fatalf("set failed: %v", err)
				}

				time.Sleep(2 * time.Second)

				_, err := rc.Get(ctx, "k:ttl")
				if !errors.Is(err, ErrCacheMiss) {
					t.Fatalf("expected ErrCacheMiss after expiry, got %v", err)
				}
			},
		},
		{
			name: "delete",
			run: func(t *testing.T, rc *RedisCache) {
				if err := rc.Set(ctx, "k:delete", []byte("value"), time.Minute); err != nil {
					t.Fatalf("set failed: %v", err)
				}
				if err := rc.Delete(ctx, "k:delete"); err != nil {
					t.Fatalf("delete failed: %v", err)
				}

				_, err := rc.Get(ctx, "k:delete")
				if !errors.Is(err, ErrCacheMiss) {
					t.Fatalf("expected ErrCacheMiss after delete, got %v", err)
				}
			},
		},
		{
			name: "exists true false",
			run: func(t *testing.T, rc *RedisCache) {
				if err := rc.Set(ctx, "k:exists", []byte("value"), time.Minute); err != nil {
					t.Fatalf("set failed: %v", err)
				}

				exists, err := rc.Exists(ctx, "k:exists")
				if err != nil {
					t.Fatalf("exists failed: %v", err)
				}
				if !exists {
					t.Fatal("expected exists=true")
				}

				exists, err = rc.Exists(ctx, "k:not-found")
				if err != nil {
					t.Fatalf("exists failed: %v", err)
				}
				if exists {
					t.Fatal("expected exists=false")
				}
			},
		},
		{
			name: "close idempotency and post close error",
			run: func(t *testing.T, rc *RedisCache) {
				if err := rc.Close(); err != nil {
					t.Fatalf("close failed: %v", err)
				}
				if err := rc.Close(); err != nil {
					t.Fatalf("second close should be idempotent: %v", err)
				}

				if err := rc.Set(ctx, "k:after-close", []byte("value"), time.Minute); err == nil {
					t.Fatal("expected error after close")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := newTestRedisCache(t, addr)
			defer func() { _ = rc.Close() }()
			tt.run(t, rc)
		})
	}
}

func requireRedisAddr(t *testing.T) string {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping redis integration tests in short mode")
	}
	return "127.0.0.1:6379"
}

func newTestRedisCache(t *testing.T, addr string) *RedisCache {
	t.Helper()
	rc, err := NewRedis(RedisConfig{Addr: addr, DialTimeout: 2 * time.Second})
	if err != nil {
		t.Skipf("redis not available at %s: %v", addr, err)
	}

	key := fmt.Sprintf("test:flush:%d", time.Now().UnixNano())
	if err := rc.Delete(context.Background(), key); err != nil {
		t.Fatalf("pre-flight delete failed: %v", err)
	}

	return rc
}
