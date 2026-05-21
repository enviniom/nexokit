package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/generator"
	"github.com/enviniom/nexokit/internal/cli/templates"
	"github.com/enviniom/nexokit/internal/infra/db"
)

// MakeCommand handles code generation subcommands.
type MakeCommand struct{}

func (c *MakeCommand) Name() string { return "make" }
func (c *MakeCommand) Description() string {
	return "Generate modules, migrations, and seeds (module, migration, seed)"
}

func (c *MakeCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: make <module|migration|seed> [args...]")
	}

	sub := args[0]
	switch sub {
	case "module":
		return c.runModule(args[1:], stdio)
	case "migration":
		return c.runMigration(args[1:])
	case "seed":
		return c.runSeed(args[1:])
	default:
		return fmt.Errorf("unknown make subcommand: %s", sub)
	}
}

func (c *MakeCommand) runModule(args []string, stdio cli.Stdio) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: make module <name> [--crud] [--migration] [--tenant]")
	}

	name := args[0]
	opts := generator.ModuleOptions{Name: name}
	for _, a := range args[1:] {
		switch a {
		case "--crud":
			opts.CRUD = true
		case "--migration":
			opts.Migration = true
		case "--tenant":
			opts.Tenant = true
		default:
			return fmt.Errorf("unknown flag: %s", a)
		}
	}

	if err := generator.GenerateModule(opts); err != nil {
		return err
	}

	stdio.Printf("Module '%s' generated successfully.\n", name)
	if opts.Migration {
		stdio.Println("Migration file created in migrations/")
	}
	return nil
}

func (c *MakeCommand) runMigration(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: make migration <name>")
	}
	name := args[0]
	if !isValidMigrationName(name) {
		return fmt.Errorf("migration name must be snake_case (lowercase letters, digits, underscores)")
	}
	return db.CreateMigration("migrations", name)
}

func (c *MakeCommand) runSeed(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: make seed <name>")
	}
	name := args[0]
	if !isValidMigrationName(name) {
		return fmt.Errorf("seed name must be snake_case (lowercase letters, digits, underscores)")
	}
	return createSeedFile(name)
}

func createSeedFile(name string) error {
	seedsDir := "seeds"
	if err := os.MkdirAll(seedsDir, 0755); err != nil {
		return fmt.Errorf("failed to create seeds directory: %w", err)
	}

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.go", timestamp, name)
	path := filepath.Join(seedsDir, filename)

	content := fmt.Sprintf(`package seeds

// %sSeed runs the %s seed.
// TODO: implement seed logic. This is executed by the seed runner in WU3.
func %sSeed() error {
	return nil
}
`, name, name, templates.StructName(name))

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write seed file: %w", err)
	}
	return nil
}
