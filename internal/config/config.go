package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	App      AppConfig
	DB       DBConfig
	CORS     CORSConfig
	Log      LogConfig
	Shutdown ShutdownConfig
	Cache    CacheConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Name string
	Env  string
	URL  string
	Port int
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	SSLMode         string
	DatabaseURL     string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// CORSConfig holds CORS-related settings.
type CORSConfig struct {
	AllowedOrigins string
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level      string
	Format     string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
	GinFile    string
	ErrorFile  string
}

// ShutdownConfig holds graceful shutdown settings.
type ShutdownConfig struct {
	Timeout time.Duration
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	Driver string
}

// Load reads environment variables and returns a typed Config.
// It fails fast on missing required values.
func Load() (*Config, error) {
	_ = godotenv.Load() // optional .env file

	port, err := getInt("APP_PORT", 8080)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_PORT: %w", err)
	}

	dbPort, err := getInt("DB_PORT", 5432)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	shutdownSeconds, err := getInt("SHUTDOWN_TIMEOUT_SECONDS", 30)
	if err != nil {
		return nil, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_SECONDS: %w", err)
	}

	maxOpenConns, err := getInt("DB_MAX_OPEN_CONNS", 25)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	maxIdleConns, err := getInt("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	connMaxLifetimeSeconds, err := getInt("DB_CONN_MAX_LIFETIME_SECONDS", 3600)
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME_SECONDS: %w", err)
	}

	logMaxSize, err := getInt("LOG_MAX_SIZE", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid LOG_MAX_SIZE: %w", err)
	}

	logMaxBackups, err := getInt("LOG_MAX_BACKUPS", 3)
	if err != nil {
		return nil, fmt.Errorf("invalid LOG_MAX_BACKUPS: %w", err)
	}

	logMaxAge, err := getInt("LOG_MAX_AGE", 28)
	if err != nil {
		return nil, fmt.Errorf("invalid LOG_MAX_AGE: %w", err)
	}

	cfg := &Config{
		App: AppConfig{
			Name: getString("APP_NAME", "nexokit"),
			Env:  getString("APP_ENV", "development"),
			URL:  getString("APP_URL", "http://localhost:8080"),
			Port: port,
		},
		DB: DBConfig{
			Host:            getString("DB_HOST", "localhost"),
			Port:            dbPort,
			Name:            getString("DB_NAME", "nexokit"),
			User:            getString("DB_USER", "nexokit"),
			Password:        getString("DB_PASSWORD", "nexokit"),
			SSLMode:         getString("DB_SSL_MODE", "disable"),
			DatabaseURL:     getString("DATABASE_URL", ""),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: time.Duration(connMaxLifetimeSeconds) * time.Second,
		},
		CORS: CORSConfig{
			AllowedOrigins: getString("CORS_ALLOWED_ORIGINS", "*"),
		},
		Log: LogConfig{
			Level:      getString("LOG_LEVEL", "info"),
			Format:     getString("LOG_FORMAT", "json"),
			File:       getString("LOG_FILE", "logs/app.log"),
			MaxSize:    logMaxSize,
			MaxBackups: logMaxBackups,
			MaxAge:     logMaxAge,
			Compress:   getBool("LOG_COMPRESS", true),
			GinFile:    getString("LOG_GIN_FILE", "logs/gin.log"),
			ErrorFile:  getString("LOG_ERROR_FILE", "logs/error.log"),
		},
		Shutdown: ShutdownConfig{
			Timeout: time.Duration(shutdownSeconds) * time.Second,
		},
		Cache: CacheConfig{
			Driver: getString("CACHE_DRIVER", "none"),
		},
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback, err
	}
	return n, nil
}

func getBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
