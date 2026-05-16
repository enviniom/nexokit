package commands

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/root"
	"github.com/enviniom/nexokit/internal/config"
)

// CreateRootCommand creates the root user.
type CreateRootCommand struct{}

func (c *CreateRootCommand) Name() string        { return "create-root" }
func (c *CreateRootCommand) Description() string { return "Create the root user (idempotent)" }

func (c *CreateRootCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	fs := flag.NewFlagSet("create-root", flag.ContinueOnError)
	fs.SetOutput(stdio.Err)

	email := fs.String("email", "", "Root user email")
	password := fs.String("password", "", "Root user password")
	force := fs.Bool("force", false, "Skip confirmation prompt in non-local environments")

	if err := fs.Parse(args); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdio.In)

	var err error
	if *email == "" {
		*email, err = prompt(stdio, scanner, "Email: ")
		if err != nil {
			return fmt.Errorf("failed to read email: %w", err)
		}
	}
	if *password == "" {
		*password, err = prompt(stdio, scanner, "Password: ")
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
	}

	input := root.CreateRootInput{
		Email:    strings.TrimSpace(*email),
		Password: *password,
	}

	requireConfirm := true
	if cfg, cfgErr := config.Load(); cfgErr == nil {
		if cfg.IsLocal() || cfg.IsTest() {
			requireConfirm = false
		}
	}

	if requireConfirm && !*force {
		stdio.Println("You are about to create a root user in a non-local environment.")
		stdio.Printf("Email: %s\n", input.Email)
		confirm, err := prompt(stdio, scanner, "Type 'yes' to confirm: ")
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		if strings.TrimSpace(strings.ToLower(confirm)) != "yes" {
			return fmt.Errorf("aborted")
		}
	}

	creator := root.NewCreator(nil, nil)
	if err := creator.Create(input); err != nil {
		if errors.Is(err, root.ErrRootAlreadyExists) {
			stdio.Println("Root user already exists. Skipping.")
			return nil
		}
		return err
	}

	stdio.Println("Root user created successfully.")
	return nil
}

func prompt(stdio cli.Stdio, scanner *bufio.Scanner, label string) (string, error) {
	stdio.Printf("%s", label)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("no input")
	}
	return scanner.Text(), nil
}
