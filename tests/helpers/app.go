package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/app"
)

// NewTestApp bootstraps a full application instance suitable for integration tests.
// It uses in-memory or test-specific configuration and returns the App for testing.
// Cleanup is registered with t.Cleanup; the caller does not need to call Stop().
func NewTestApp(t *testing.T) *app.App {
	t.Helper()

	ctx := context.Background()
	application, err := app.Bootstrap(ctx)
	if err != nil {
		t.Fatalf("failed to bootstrap test app: %v", err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = application.Stop(ctx)
	})

	return application
}
