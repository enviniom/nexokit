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
	"strconv"
	"strings"

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
	return runSeedsWithEnvironment(ctx, funcs, stdio, nil)
}

func runSeedsWithEnvironment(ctx context.Context, funcs []seedFunc, stdio cli.Stdio, values map[string]string) error {
	module, err := currentSeedModule()
	if err != nil {
		return fmt.Errorf("seed runner failed: module validation failed")
	}
	if err := os.MkdirAll(".tmp", 0700); err != nil {
		return fmt.Errorf("failed to create runner directory")
	}
	runnerDir, err := os.MkdirTemp(".tmp", "seed-runner-")
	if err != nil {
		return fmt.Errorf("failed to create runner directory: %w", err)
	}
	defer os.RemoveAll(runnerDir)

	mainGo := generateRunnerSourceForModule(funcs, module)

	runnerFile := filepath.Join(runnerDir, "main.go")
	if err := os.WriteFile(runnerFile, []byte(mainGo), 0644); err != nil {
		return fmt.Errorf("failed to write runner: %w", err)
	}

	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Dir = runnerDir
	cmd.Stdout = stdio.Out
	cmd.Stderr = stdio.Err
	if values != nil {
		cmd.Env = seedEnvironment(os.Environ(), values)
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("seed runner failed: %w", err)
	}

	return nil
}

// currentSeedModule reads the adopted module at execution, not compilation.
func currentSeedModule() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "module" {
			continue
		}
		name := fields[1]
		if strings.HasPrefix(name, "\"") {
			name, err = strconv.Unquote(name)
			if err != nil {
				return "", err
			}
		}
		if name == "" || strings.ContainsAny(name, " \t\r\n\"\\") || strings.HasPrefix(name, "/") || strings.Contains(name, "..") {
			return "", fmt.Errorf("invalid module")
		}
		for _, r := range name {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune(".-_~/", r)) {
				return "", fmt.Errorf("invalid module")
			}
		}
		return name, nil
	}
	return "", fmt.Errorf("missing module")
}

// seedEnvironment removes inherited DB routing and root credentials before
// appending the complete effective configuration. Standalone seed keeps its
// existing environment because it calls the runner without this override.
func seedEnvironment(inherited []string, values map[string]string) []string {
	result := []string{}
	for _, entry := range inherited {
		key, _, _ := strings.Cut(entry, "=")
		_, overridden := values[key]
		if overridden || key == "DATABASE_URL" || strings.HasPrefix(key, "DB_") || strings.HasPrefix(key, "PG") || strings.HasPrefix(key, "ROOT_") {
			continue
		}
		result = append(result, entry)
	}
	for key, value := range values {
		if strings.HasPrefix(key, "ROOT_") || (strings.HasPrefix(key, "PG") && key != "PGPASSWORD" && key != "PGPASSFILE" && key != "PGSSLMODE" && key != "PGPORT") {
			continue
		}
		result = append(result, key+"="+value)
	}
	return result
}

// generateRunnerSource returns the Go source for a main package that calls
// every discovered seed function. It is a pure function and can be unit-tested.
func generateRunnerSource(funcs []seedFunc) string {
	return generateRunnerSourceForModule(funcs, "github.com/enviniom/nexokit")
}

func generateRunnerSourceForModule(funcs []seedFunc, module string) string {
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

	%s
)

func main() {
%s
}
`, strconv.Quote(module+"/seeds"), calls.String())
}
