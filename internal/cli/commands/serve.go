package commands

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/enviniom/nexokit/internal/app"
	"github.com/enviniom/nexokit/internal/cli"
)

// ServeCommand starts the HTTP server.
type ServeCommand struct{}

func (c *ServeCommand) Name() string        { return "serve" }
func (c *ServeCommand) Description() string { return "Start the HTTP server" }

func (c *ServeCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.Bootstrap(ctx)
	if err != nil {
		return fmt.Errorf("failed to bootstrap application: %w", err)
	}

	go func() {
		if err := application.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(stdio.Err, "server error: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), application.Config.Shutdown.Timeout)
	defer cancel()

	if err := application.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	return nil
}
