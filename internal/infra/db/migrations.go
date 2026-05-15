package db

import (
	"context"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// RunMigrations applies all pending migrations from the given directory.
func RunMigrations(db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigrations reverts the last batch of migrations.
func RollbackMigrations(db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Down(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

// MigrationStatus returns the current migration status.
func MigrationStatus(db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	return goose.Status(sqlDB, migrationsDir)
}

// CreateMigration generates a new migration file with the given name.
func CreateMigration(migrationsDir, name string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	return goose.Create(nil, migrationsDir, name, "sql")
}

// ResetMigrations rolls back all migrations.
func ResetMigrations(ctx context.Context, db *gorm.DB, migrationsDir string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Reset(sqlDB, migrationsDir); err != nil {
		return fmt.Errorf("failed to reset migrations: %w", err)
	}

	return nil
}
