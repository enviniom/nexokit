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
	return "Initialize local application identity (preview by default)"
}
func (*InitCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	if len(args) == 0 {
		return runInteractiveInit(ctx, stdio)
	}
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(stdio.Err)
	name := flags.String("name", "", "Application display name")
	module := flags.String("module", "", "Independent Go module path")
	apply := flags.Bool("apply", false, "Write the validated plan")
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
		return errors.New("usage: init --name <application> --module <module> [--dry-run | --apply]")
	}
	if seen["apply"] && seen["dry-run"] {
		return errors.New("--apply and --dry-run are mutually exclusive")
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	root, e := os.Getwd()
	if e != nil {
		return e
	}
	plan, e := projectinit.New(root, *name, *module)
	if e != nil {
		return e
	}
	printInitPreview(stdio, plan)
	if !*apply {
		stdio.Println("No files written. Review this plan, then repeat with --apply.")
		return nil
	}
	return applyInit(ctx, stdio, plan)
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

func applyInit(ctx context.Context, stdio cli.Stdio, plan *projectinit.Plan) error {
	if e := ctx.Err(); e != nil {
		return e
	}
	if e := plan.Apply(); e != nil {
		return e
	}
	stdio.Println("Applied. Application identity changed; cmd/nexokit remains the developer tool.")
	stdio.Println("Existing .env is unchanged. Configure PostgreSQL, PASETO_KEY, CORS and rate limits deliberately.")
	stdio.Println("Next: go build ./...; go test ./...; then run migrate up, seed and create-root yourself after configuring the database.")
	return nil
}
