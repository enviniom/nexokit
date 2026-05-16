package commands

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/enviniom/nexokit/internal/cli"
)

// SeedCommand discovers and runs seed files.
type SeedCommand struct{}

func (c *SeedCommand) Name() string        { return "seed" }
func (c *SeedCommand) Description() string { return "Run seed files from seeds/" }

func (c *SeedCommand) Run(ctx context.Context, args []string, stdio cli.Stdio) error {
	seedsDir := "seeds"
	funcs, err := discoverSeedFunctions(seedsDir)
	if err != nil {
		return err
	}
	if len(funcs) == 0 {
		stdio.Println("No seed files found in seeds/")
		return nil
	}

	stdio.Printf("Running %d seed(s)...\n", len(funcs))
	return runSeeds(ctx, funcs, stdio)
}

// seedFunc represents a discovered seed function.
type seedFunc struct {
	file string
	name string
}

func discoverSeedFunctions(dir string) ([]seedFunc, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read seeds directory: %w", err)
	}

	var funcs []seedFunc
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", path, err)
		}
		if f.Name.Name != "seeds" {
			continue
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if !strings.HasSuffix(fn.Name.Name, "Seed") {
				continue
			}
			if !isNoArgErrorFunc(fn) {
				continue
			}
			funcs = append(funcs, seedFunc{file: e.Name(), name: fn.Name.Name})
		}
	}
	return funcs, nil
}

func isNoArgErrorFunc(fn *ast.FuncDecl) bool {
	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		return false
	}
	if fn.Type.Results == nil {
		return false
	}
	totalResults := 0
	for _, field := range fn.Type.Results.List {
		count := len(field.Names)
		if count == 0 {
			count = 1
		}
		totalResults += count
	}
	if totalResults != 1 {
		return false
	}
	resultType := fn.Type.Results.List[0].Type
	if resultType == nil {
		return false
	}
	ident, ok := resultType.(*ast.Ident)
	return ok && ident.Name == "error"
}

// runSeeds generates a temporary Go program that imports the seeds package
// and calls each discovered seed function. A temporary main.go is required
// because Go needs an executable main package to run code; seed files live
// in package seeds and export functions, so we synthesise a runner that
// imports them and invokes each one in order.
func runSeeds(ctx context.Context, funcs []seedFunc, stdio cli.Stdio) error {
	runnerDir := filepath.Join(".tmp", fmt.Sprintf("seed-runner-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(runnerDir, 0755); err != nil {
		return fmt.Errorf("failed to create runner directory: %w", err)
	}
	defer os.RemoveAll(runnerDir)

	mainGo := generateRunnerSource(funcs)

	runnerFile := filepath.Join(runnerDir, "main.go")
	if err := os.WriteFile(runnerFile, []byte(mainGo), 0644); err != nil {
		return fmt.Errorf("failed to write runner: %w", err)
	}

	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = runnerDir
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("seed runner failed: %w", err)
	}

	return nil
}

// generateRunnerSource returns the Go source for a main package that calls
// every discovered seed function. It is a pure function and can be unit-tested.
func generateRunnerSource(funcs []seedFunc) string {
	var calls strings.Builder
	for _, fn := range funcs {
		calls.WriteString(fmt.Sprintf("\tif err := seeds.%s(); err != nil {\n", fn.name))
		calls.WriteString(fmt.Sprintf("\t\tfmt.Fprintf(os.Stderr, \"seed %%s failed: %%v\\n\", %q, err)\n", fn.name))
		calls.WriteString("\t\tos.Exit(1)\n")
		calls.WriteString("\t}\n")
		calls.WriteString(fmt.Sprintf("\tfmt.Println(\"seed %s: ok\")\n", fn.name))
	}

	return fmt.Sprintf(`package main

import (
	"fmt"
	"os"

	"github.com/enviniom/nexokit/seeds"
)

func main() {
%s
}
`, calls.String())
}
