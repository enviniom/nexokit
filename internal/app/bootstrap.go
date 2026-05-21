package app

import (
	"context"
	"fmt"
	"net"

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

	var c cache.Cache = cache.NewNoop()

	switch cfg.Cache.Driver {
	case "redis":
		redisAddr := net.JoinHostPort(cfg.Redis.Host, fmt.Sprintf("%d", cfg.Redis.Port))
		redisCache, redisErr := cache.NewRedis(cache.RedisConfig{
			Addr:        redisAddr,
			Password:    cfg.Redis.Password,
			DB:          cfg.Redis.DB,
			DialTimeout: cfg.Redis.DialTimeout,
		})
		if redisErr != nil {
			log.Warn("failed to connect to redis cache, falling back to noop cache", slog.String("error", redisErr.Error()))
		} else {
			c = redisCache
		}
	case "none":
		// no-op cache by design
	default:
		log.Warn("unsupported cache driver, falling back to noop cache", slog.String("driver", cfg.Cache.Driver))
	}

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
