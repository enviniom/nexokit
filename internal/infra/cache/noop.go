package cache

import (
	"context"
	"time"
)

// NoopCache is a no-op implementation of Cache.
// It satisfies the Cache interface without performing any operations.
type NoopCache struct{}

// NewNoop creates a new NoopCache instance.
func NewNoop() *NoopCache {
	return &NoopCache{}
}

// Get always returns a cache miss.
func (n *NoopCache) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrCacheMiss
}

// Set is a no-op.
func (n *NoopCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return nil
}

// Delete is a no-op.
func (n *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Exists always returns false with no error.
func (n *NoopCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// Close is a no-op.
func (n *NoopCache) Close() error {
	return nil
}
