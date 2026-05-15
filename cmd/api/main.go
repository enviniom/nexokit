package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/enviniom/nexokit/internal/app"
	"log/slog"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.Bootstrap(ctx)
	if err != nil {
		slog.Error("failed to bootstrap application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() {
		if err := application.Start(); err != nil {
			application.Logger.Error("server error", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), application.Config.Shutdown.Timeout)
	defer cancel()

	if err := application.Stop(shutdownCtx); err != nil {
		application.Logger.Error("shutdown error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
