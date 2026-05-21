package commands

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	"github.com/pressly/goose/v3"
)

// StatusCommand prints application status.
type StatusCommand struct{}

func (c *StatusCommand) Name() string { return "status" }
func (c *StatusCommand) Description() string {
	return "Print app version, DB status, and migration count"
}

func (c *StatusCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
	fmt.Fprintf(stdio.Out, "Version: %s\n", version)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	migrationsDir := "migrations"
	totalFiles := countMigrationFiles(migrationsDir)

	database, connErr := db.Connect(cfg)
	if connErr != nil {
		fmt.Fprintf(stdio.Out, "Database: unreachable (%v)\n", connErr)
		fmt.Fprintf(stdio.Out, "Migrations: %d files in %s/\n", totalFiles, migrationsDir)
		return nil
	}
	defer db.Close(database)

	sqlDB, err := database.DB()
	if err != nil {
		fmt.Fprintf(stdio.Out, "Database: unreachable (%v)\n", err)
		fmt.Fprintf(stdio.Out, "Migrations: %d files in %s/\n", totalFiles, migrationsDir)
		return nil
	}

	if err := sqlDB.Ping(); err != nil {
		fmt.Fprintf(stdio.Out, "Database: unreachable (%v)\n", err)
		fmt.Fprintf(stdio.Out, "Migrations: %d files in %s/\n", totalFiles, migrationsDir)
		return nil
	}

	fmt.Fprintln(stdio.Out, "Database: connected")

	if err := goose.SetDialect("postgres"); err != nil {
		fmt.Fprintf(stdio.Out, "Migration version: unknown (failed to set dialect: %v)\n", err)
	} else {
		currentVersion, err := goose.GetDBVersion(sqlDB)
		if err != nil {
			fmt.Fprintf(stdio.Out, "Migration version: unknown (%v)\n", err)
		} else {
			fmt.Fprintf(stdio.Out, "Migration version: %d\n", currentVersion)
		}
	}

	fmt.Fprintf(stdio.Out, "Migrations: %d files in %s/\n", totalFiles, migrationsDir)

	return nil
}

func countMigrationFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			count++
		}
	}
	return count
}
