package app

import (
	"context"
	"fmt"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/infra/db"
	"github.com/enviniom/nexokit/internal/server"
	"gorm.io/gorm"
	"log/slog"
)

// App aggregates the core application dependencies.
type App struct {
	Config    *config.Config
	Logger    *slog.Logger
	DB        *gorm.DB
	Server    *server.Server
	Cache     cache.Cache
	Container *Container
}

// Start begins serving HTTP requests.
func (a *App) Start() error {
	a.Logger.Info("starting server", slog.String("address", a.Config.App.URL))
	return a.Server.Start()
}

// Stop gracefully shuts down all application components.
func (a *App) Stop(ctx context.Context) error {
	a.Logger.Info("shutting down server")
	if err := a.Server.Stop(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	a.Logger.Info("closing database connection")
	if err := db.Close(a.DB); err != nil {
		return fmt.Errorf("database close failed: %w", err)
	}

	a.Logger.Info("closing cache connection")
	if err := a.Cache.Close(); err != nil {
		return fmt.Errorf("cache close failed: %w", err)
	}

	a.Logger.Info("shutdown complete")
	return nil
}
