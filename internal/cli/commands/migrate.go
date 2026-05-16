package commands

import (
	"context"
	"fmt"
	"unicode"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
)

// MigrateCommand handles database migrations.
type MigrateCommand struct{}

func (c *MigrateCommand) Name() string        { return "migrate" }
func (c *MigrateCommand) Description() string { return "Run database migrations (up, down, status, reset, create)" }

func (c *MigrateCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: migrate <up|down|status|reset|create>")
	}

	sub := args[0]
	migrationsDir := "migrations"

	switch sub {
	case "up", "down", "status", "reset":
		return c.runWithDB(ctx, sub, migrationsDir, stdio)
	case "create":
		if len(args) < 2 {
			return fmt.Errorf("usage: migrate create <name>")
		}
		name := args[1]
		if !isValidMigrationName(name) {
			return fmt.Errorf("migration name must be snake_case (lowercase letters, digits, underscores)")
		}
		return db.CreateMigration(migrationsDir, name)
	default:
		return fmt.Errorf("unknown migrate subcommand: %s", sub)
	}
}

func (c *MigrateCommand) runWithDB(ctx context.Context, sub, migrationsDir string, stdio cli.Stdio) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close(database)

	switch sub {
	case "up":
		return db.RunMigrations(database, migrationsDir)
	case "down":
		return db.RollbackMigrations(database, migrationsDir)
	case "status":
		return db.MigrationStatus(database, migrationsDir)
	case "reset":
		return db.ResetMigrations(ctx, database, migrationsDir)
	default:
		return fmt.Errorf("unknown migrate subcommand: %s", sub)
	}
}

func isValidMigrationName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		if !unicode.IsLower(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}
	return true
}
