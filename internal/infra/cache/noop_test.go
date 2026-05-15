package cache

import (
	"context"
	"testing"
	"time"
)

func TestNoopCache(t *testing.T) {
	c := NewNoop()
	ctx := context.Background()

	val, err := c.Get(ctx, "key")
	if err != nil {
		t.Errorf("unexpected error from Get: %v", err)
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

	if err := c.Close(); err != nil {
		t.Errorf("unexpected error from Close: %v", err)
	}
}
