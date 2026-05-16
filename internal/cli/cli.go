package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
)

// Command is the interface for all CLI subcommands.
type Command interface {
	Name() string
	Description() string
	Run(ctx context.Context, args []string, stdio Stdio) error
}

// Stdio holds the standard I/O streams for a command.
type Stdio struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Registry holds all available commands.
type Registry struct {
	commands map[string]Command
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry.
func (r *Registry) Register(cmd Command) {
	r.commands[cmd.Name()] = cmd
}

// Get retrieves a command by name.
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// Names returns all registered command names sorted.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.commands))
	for name := range r.commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Execute runs the CLI with the given arguments and commands.
func Execute(args []string, stdio Stdio, cmds []Command) int {
	registry := NewRegistry()
	for _, cmd := range cmds {
		registry.Register(cmd)
	}

	if len(args) == 0 {
		printUsage(registry, stdio.Out)
		return 1
	}

	name := args[0]
	if name == "help" || name == "--help" || name == "-h" {
		printUsage(registry, stdio.Out)
		return 0
	}

	cmd, ok := registry.Get(name)
	if !ok {
		fmt.Fprintf(stdio.Err, "unknown command: %s\n\n", name)
		printUsage(registry, stdio.Err)
		return 1
	}

	ctx := context.Background()
	if err := cmd.Run(ctx, args[1:], stdio); err != nil {
		fmt.Fprintf(stdio.Err, "error: %v\n", err)
		return 1
	}

	return 0
}

func printUsage(registry *Registry, w io.Writer) {
	fmt.Fprintln(w, "nexokit CLI")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage: nexokit <command> [args...]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Commands:")
	for _, name := range registry.Names() {
		cmd, _ := registry.Get(name)
		fmt.Fprintf(w, "  %-12s %s\n", name, cmd.Description())
	}
}
