package app

import (
	"context"
	"fmt"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/cache"
	"github.com/enviniom/nexokit/internal/infra/db"
	"github.com/enviniom/nexokit/internal/infra/logger"
	"github.com/enviniom/nexokit/internal/server"
	"log/slog"
)

// Bootstrap initializes the application in the correct order:
// config → logger → db → cache → container → router → server.
func Bootstrap(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	log, err := logger.New(cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	log.Info("configuration loaded", slog.String("env", cfg.App.Env))

	database, err := db.Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Info("database connected")

	c := cache.NewNoop()

	container := NewContainer(cfg, database, log, c)

	sqlDB, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	healthDeps := server.HealthDeps{
		DB:           sqlDB,
		Cache:        c,
		CacheEnabled: cfg.Cache.Driver != "none",
	}

	ginWriter, err := logger.GinWriter(cfg.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to create gin writer: %w", err)
	}

	router := server.NewRouter(cfg, log, ginWriter, healthDeps, container.RegisterModules)

	srv := server.New(cfg, router)

	app := &App{
		Config:    cfg,
		Logger:    log,
		DB:        database,
		Server:    srv,
		Cache:     c,
		Container: container,
	}

	return app, nil
}
