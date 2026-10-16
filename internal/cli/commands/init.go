package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
)

// InitCommand adopts the current local starter once; it never provisions services.
type InitCommand struct{}

func (*InitCommand) Name() string { return "init" }
func (*InitCommand) Description() string {
	return "Initialize local identity and environment once (flags preview by default)"
}
func (*InitCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	if len(args) == 0 {
		return runInteractiveInit(ctx, stdio)
	}
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stdio.Err)
	name := flags.String("name", "", "Application display name")
	module := flags.String("module", "", "Independent Go module path")
	apply := flags.Bool("apply", false, "Apply one-time local identity and environment setup; requires explicit local settings")
	envAction := flags.String("env-action", "", "Existing environment choice: keep or rebuild (required with --apply when present)")
	flags.Bool("dry-run", false, "Explicit preview without writes")
	seen := map[string]bool{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			key := strings.SplitN(strings.TrimLeft(arg, "-"), "=", 2)[0]
			if seen[key] {
				return fmt.Errorf("duplicate flag: %s", key)
			}
			seen[key] = true
		}
	}
	if e := flags.Parse(args); e != nil {
		return e
	}
	if flags.NArg() != 0 || *name == "" || *module == "" {
		return errors.New("usage: init --name <application> --module <module> [--env-action keep|rebuild] [--dry-run | --apply]")
	}
	if seen["apply"] && seen["dry-run"] {
		return errors.New("--apply and --dry-run are mutually exclusive")
	}
	if *envAction != "" && *envAction != "keep" && *envAction != "rebuild" {
		return errors.New("--env-action must be keep or rebuild")
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	root, e := os.Getwd()
	if e != nil {
		return e
	}
	if e := projectinit.CheckUninitialized(root); e != nil {
		return e
	}
	plan, e := projectinit.New(root, *name, *module)
	if e != nil {
		return e
	}
	if !*apply {
		printInitPreview(stdio, plan)
		stdio.Println("No files written. Review this plan, then repeat with --apply and complete setup inputs.")
		return nil
	}
	exists, e := plan.EnvironmentExists()
	if e != nil {
		return e
	}
	if exists && *envAction == "" {
		return errors.New("existing environment requires --env-action keep|rebuild; no files written")
	}
	if !exists && *envAction == "keep" {
		return errors.New("cannot keep an absent environment; no files written")
	}
	if !exists || *envAction == "rebuild" {
		values, e := initFlagEnvironment()
		if e != nil {
			return e
		}
		if e := plan.ConfigureEnvironment(values); e != nil {
			return e
		}
	}
	selected, e := plan.SelectEnvironment()
	if e != nil {
		return e
	}
	target, e := validateInitSetup(selected.Values)
	if e != nil {
		return e
	}
	printInitPreview(stdio, plan)
	printInitSetupPreview(stdio, target)
	return executeInitSetup(ctx, stdio, plan, selected)
}

func printInitPreview(stdio cli.Stdio, plan *projectinit.Plan) {
	stdio.Println("Preview: one-time local application initialization")
	for _, change := range plan.Changes {
		stdio.Printf("  %s\n", change.Path)
	}
	stdio.Println("Remaining tool/provenance references (not renamed):")
	for _, path := range plan.Remaining {
		stdio.Printf("  %s\n", path)
	}
}
