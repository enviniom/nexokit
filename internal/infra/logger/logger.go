package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/enviniom/nexokit/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates the app logger (structured JSON/text, all levels) → logs/app.log.
func New(cfg config.LogConfig) (*slog.Logger, error) {
	output, err := openLogOutput(cfg.File, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open app log: %w", err)
	}

	handler := newHandler(output, cfg)
	return slog.New(handler), nil
}

// NewErrorLogger creates the error-only logger (ERROR level and above) → logs/error.log.
func NewErrorLogger(cfg config.LogConfig) (*slog.Logger, error) {
	output, err := openLogOutput(cfg.ErrorFile, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to open error log: %w", err)
	}

	opts := &slog.HandlerOptions{Level: slog.LevelError}
	handler := newHandlerWithOptions(output, cfg, opts)
	return slog.New(handler), nil
}

// GinWriter returns an io.Writer for Gin's access logs → logs/gin.log.
func GinWriter(cfg config.LogConfig) (io.Writer, error) {
	return openLogOutput(cfg.GinFile, cfg)
}

func openLogOutput(path string, cfg config.LogConfig) (io.Writer, error) {
	if path == "" {
		return os.Stdout, nil
	}
	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}, nil
}

func newHandler(output io.Writer, cfg config.LogConfig) slog.Handler {
	opts := &slog.HandlerOptions{Level: parseLevel(cfg.Level)}
	return newHandlerWithOptions(output, cfg, opts)
}

func newHandlerWithOptions(output io.Writer, cfg config.LogConfig, opts *slog.HandlerOptions) slog.Handler {
	if cfg.Format == "text" {
		return slog.NewTextHandler(output, opts)
	}
	return slog.NewJSONHandler(output, opts)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
