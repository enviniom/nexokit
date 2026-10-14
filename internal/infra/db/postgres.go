package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"net/url"
	"os"
	"strings"

	"github.com/enviniom/nexokit/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to PostgreSQL using GORM.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := buildDSN(cfg)

	logLevel := logger.Silent
	if cfg.IsLocal() || cfg.IsTest() {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// Close closes the underlying database connection pool.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// ConnectQuiet preconnects with context and never logs SQL, hashes or server
// errors. Failed connections are closed; successful pools belong to the caller.
func ConnectQuiet(ctx context.Context, cfg *config.Config) (*gorm.DB, error) {
	if os.Getenv("PGSERVICE") != "" {
		return nil, fmt.Errorf("database connection failed")
	}
	dsn := buildDSN(cfg) + " passfile=" + quoteDSN(os.DevNull)
	password := cfg.DB.Password
	if cfg.DB.DatabaseURL != "" {
		uri, err := url.Parse(cfg.DB.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("database connection failed")
		}
		password = ""
		if uri.User != nil {
			password, _ = uri.User.Password()
		}
		query := uri.Query()
		query.Set("passfile", os.DevNull)
		if _, ok := query["service"]; ok {
			return nil, fmt.Errorf("database connection failed")
		}
		if _, ok := query["servicefile"]; ok {
			return nil, fmt.Errorf("database connection failed")
		}
		if values, ok := query["password"]; ok && len(values) > 0 {
			password = values[0]
		}
		if query.Get("sslmode") == "" {
			query.Set("sslmode", "prefer")
		}
		if uri.Port() == "" && query.Get("port") == "" {
			query.Set("port", "5432")
		}
		uri.RawQuery = query.Encode()
		dsn = uri.String()
	}
	parsed, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection failed")
	}
	// ParseConfig may consult pgpass/PGPASSWORD for an empty password. The
	// initializer uses the selected credential literally, including empty.
	parsed.Password = password
	pool := stdlib.OpenDB(*parsed)
	database, err := gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), DisableAutomaticPing: true,
	})
	if err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("database connection failed")
	}
	pool.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	pool.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	pool.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)
	if err := pool.PingContext(ctx); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("database connection failed")
	}
	return database, nil
}

func quoteDSN(value string) string {
	if value != "" && !strings.ContainsAny(value, " \t\n\r'\\") {
		return value
	}
	return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'").Replace(value) + "'"
}

func buildDSN(cfg *config.Config) string {
	if cfg.DB.DatabaseURL != "" {
		return cfg.DB.DatabaseURL
	}
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		quoteDSN(cfg.DB.Host), quoteDSN(cfg.DB.User), quoteDSN(cfg.DB.Password), quoteDSN(cfg.DB.Name), cfg.DB.Port, quoteDSN(cfg.DB.SSLMode),
	)
}
