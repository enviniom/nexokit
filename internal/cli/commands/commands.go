// Package commands implements the nexokit CLI subcommands.
package commands

import "github.com/enviniom/nexokit/internal/cli"

// All returns all available CLI commands.
func All() []cli.Command {
	return []cli.Command{
		&ServeCommand{},
		&ConfigCommand{},
		&StatusCommand{},
		&MigrateCommand{},
	}
}
