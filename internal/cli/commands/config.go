package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/config"
)

// ConfigCommand prints the resolved configuration with secrets masked.
type ConfigCommand struct{}

func (c *ConfigCommand) Name() string        { return "config" }
func (c *ConfigCommand) Description() string { return "Print resolved configuration (secrets masked)" }

func (c *ConfigCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	safe := toDisplayConfig(cfg)
	out, err := json.MarshalIndent(safe, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Fprintln(stdio.Out, string(out))
	return nil
}

type displayConfig struct {
	App      displayAppConfig      `json:"app"`
	DB       displayDBConfig       `json:"db"`
	CORS     displayCORSConfig     `json:"cors"`
	Log      displayLogConfig      `json:"log"`
	Shutdown displayShutdownConfig `json:"shutdown"`
	Cache    displayCacheConfig    `json:"cache"`
}

type displayAppConfig struct {
	Name string `json:"name"`
	Env  string `json:"env"`
	URL  string `json:"url"`
	Port int    `json:"port"`
}

type displayDBConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Name            string        `json:"name"`
	User            string        `json:"user"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	Password        string        `json:"password"`
	DatabaseURL     string        `json:"database_url"`
}

type displayCORSConfig struct {
	AllowedOrigins string `json:"allowed_origins"`
}

type displayLogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	File       string `json:"file"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
	GinFile    string `json:"gin_file"`
	ErrorFile  string `json:"error_file"`
}

type displayShutdownConfig struct {
	Timeout time.Duration `json:"timeout"`
}

type displayCacheConfig struct {
	Driver string `json:"driver"`
}

func toDisplayConfig(cfg *config.Config) displayConfig {
	return displayConfig{
		App: displayAppConfig{
			Name: cfg.App.Name,
			Env:  cfg.App.Env,
			URL:  cfg.App.URL,
			Port: cfg.App.Port,
		},
		DB: displayDBConfig{
			Host:            cfg.DB.Host,
			Port:            cfg.DB.Port,
			Name:            cfg.DB.Name,
			User:            cfg.DB.User,
			SSLMode:         cfg.DB.SSLMode,
			MaxOpenConns:    cfg.DB.MaxOpenConns,
			MaxIdleConns:    cfg.DB.MaxIdleConns,
			ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
			Password:        "***masked***",
			DatabaseURL:     "***masked***",
		},
		CORS: displayCORSConfig{
			AllowedOrigins: cfg.CORS.AllowedOrigins,
		},
		Log: displayLogConfig{
			Level:      cfg.Log.Level,
			Format:     cfg.Log.Format,
			File:       cfg.Log.File,
			MaxSize:    cfg.Log.MaxSize,
			MaxBackups: cfg.Log.MaxBackups,
			MaxAge:     cfg.Log.MaxAge,
			Compress:   cfg.Log.Compress,
			GinFile:    cfg.Log.GinFile,
			ErrorFile:  cfg.Log.ErrorFile,
		},
		Shutdown: displayShutdownConfig{
			Timeout: cfg.Shutdown.Timeout,
		},
		Cache: displayCacheConfig{
			Driver: cfg.Cache.Driver,
		},
	}
}
