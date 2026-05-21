package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNoopCache(t *testing.T) {
	c := NewNoop()
	ctx := context.Background()

	val, err := c.Get(ctx, "key")
	if !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss from Get, got: %v", err)
	}
	if val != nil {
		t.Error("expected nil value from NoopCache.Get")
	}

	if err := c.Set(ctx, "key", []byte("value"), time.Minute); err != nil {
		t.Errorf("unexpected error from Set: %v", err)
	}

	if err := c.Delete(ctx, "key"); err != nil {
		t.Errorf("unexpected error from Delete: %v", err)
	}

	exists, err := c.Exists(ctx, "key")
	if err != nil {
		t.Fatalf("unexpected error from Exists: %v", err)
	}
	if exists {
		t.Error("expected false from Exists")
	}

	if err := c.Close(); err != nil {
		t.Errorf("unexpected error from Close: %v", err)
	}
}
