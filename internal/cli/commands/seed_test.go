package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
)

func selectedBootstrapValues() map[string]string {
	return map[string]string{"DATABASE_URL": "", "DB_HOST": "selected.invalid", "DB_PORT": "5439", "DB_NAME": "owned", "DB_USER": "fake", "DB_PASSWORD": "invented #$'\\\\ password", "DB_SSL_MODE": "disable"}
}

func TestSeedCommand_NoSeedsDir(t *testing.T) {
	var out bytes.Buffer
	stdio := cli.Stdio{Out: &out}
	cmd := &SeedCommand{}

	err := cmd.Run(context.Background(), []string{}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "No seed files") {
		t.Errorf("expected 'No seed files', got %q", out.String())
	}
}

func TestGenerateRunnerSource(t *testing.T) {
	funcs := []seedFunc{
		{name: "DemoDataSeed"},
		{name: "RolesSeed"},
	}

	src := generateRunnerSource(funcs)

	// Verify package declaration and imports
	if !strings.Contains(src, "package main") {
		t.Error("expected package main")
	}
	if !strings.Contains(src, `"github.com/enviniom/nexokit/seeds"`) {
		t.Error("expected seeds import")
	}

	// Verify each seed function is called
	for _, fn := range funcs {
		if !strings.Contains(src, "seeds."+fn.name+"()") {
			t.Errorf("expected call to seeds.%s(), not found", fn.name)
		}
		if !strings.Contains(src, `"seed %s failed: %v\n"`) {
			t.Errorf("expected error message for %s, not found", fn.name)
		}
		if !strings.Contains(src, fmt.Sprintf(`"seed %s: ok"`, fn.name)) {
			t.Errorf("expected success message for %s, not found", fn.name)
		}
	}

	// Verify error handling structure
	if !strings.Contains(src, "os.Exit(1)") {
		t.Error("expected os.Exit(1) on failure")
	}
}

func TestSeedCommand_RunWithSeedsDir(t *testing.T) {
	tmpDir := t.TempDir()
	seedsDir := filepath.Join(tmpDir, "seeds")
	if err := os.MkdirAll(seedsDir, 0755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(seedsDir, "demo_data.go"), []byte(`package seeds

func DemoDataSeed() error {
	return nil
}
`), 0644)

	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	var out bytes.Buffer
	stdio := cli.Stdio{Out: &out}
	cmd := &SeedCommand{}

	// Runner will fail because there is no real seeds package in the temp module,
	// but discovery should succeed and the error should mention the runner.
	err := cmd.Run(context.Background(), []string{}, stdio)
	if err == nil {
		t.Fatal("expected runner failure in isolated temp dir")
	}
	if !strings.Contains(err.Error(), "seed runner failed") {
		t.Fatalf("expected 'seed runner failed' error, got: %v", err)
	}
}

func TestDiscoverSeedFunctions(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "valid.go"), []byte(`package seeds

func ValidSeed() error { return nil }
func helper() error { return nil }
func NoReturnSeed() {}
func WithArgsSeed(a int) error { return nil }
`), 0644)
	_ = os.WriteFile(filepath.Join(dir, "wrong_package.go"), []byte(`package main

func MainSeed() error { return nil }
`), 0644)

	funcs, err := discoverSeedFunctions(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(funcs) != 1 {
		t.Fatalf("expected 1 seed function, got %d", len(funcs))
	}
	if funcs[0].name != "ValidSeed" {
		t.Errorf("expected ValidSeed, got %s", funcs[0].name)
	}
}

func TestSeedRunnerCurrentModuleAndPinnedEnvironment(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	bin := filepath.Join(dir, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
[ "$#" = 2 ] && [ "$1" = run ] && [ "$2" = . ] || exit 21
[ "${ROOT_USER_PASSWORD+x}" != x ] || exit 22
[ "$DATABASE_URL" = "" ] && [ "$DB_NAME" = selected ] || exit 23
[ "${PGHOST+x}" != x ] || exit 24
[ "$DB_HOST" = selected.invalid ] && [ "$DB_PORT" = 5439 ] && [ "$DB_USER" = fake ] && [ "$DB_SSL_MODE" = disable ] || exit 25
expected=$(printf '%s' 'invented #$ '"'"'quotes'"'"' \ end')
[ "$DB_PASSWORD" = "$expected" ] || exit 26
cp main.go "` + filepath.Join(dir, "observed.go") + `"
printf 'seed output\n'
`
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DATABASE_URL", "postgres://wrong.invalid/wrong")
	t.Setenv("ROOT_USER_PASSWORD", "invented-root-secret")
	t.Setenv("PGHOST", "wrong.invalid")
	// Simulate adoption after compilation; the legacy generator still has its
	// original import but the runtime runner must consult this current module.
	if err := os.WriteFile("go.mod", []byte("module \"example.test/adopted\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	values := selectedBootstrapValues()
	values["DB_NAME"] = "selected"
	values["DB_PASSWORD"] = "invented #$ 'quotes' \\ end"
	values["ROOT_USER_PASSWORD"] = "invented-root-secret"
	err := runSeedsWithEnvironment(context.Background(), []seedFunc{{name: "RolesSeed"}}, cli.Stdio{Out: &out, Err: &stderr}, values)
	if err != nil || out.String() != "seed output\n" || stderr.Len() != 0 {
		t.Fatalf("runner: %v stdout %q stderr %q", err, out.String(), stderr.String())
	}
	data, err := os.ReadFile("observed.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"example.test/adopted/seeds"`) || strings.Contains(string(data), "invented-root-secret") || strings.Contains(string(data), "github.com/enviniom/nexokit/seeds") {
		t.Fatal("runner module or credentials")
	}
	entries, err := os.ReadDir(".tmp")
	if err != nil || len(entries) != 0 {
		t.Fatal("runner not cleaned")
	}
	if err := os.Mkdir("seeds", 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("seeds/roles.go", []byte("package seeds\nfunc RolesSeed() error {return nil}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out.Reset()
	stderr.Reset()
	err = (&SeedCommand{}).Run(ctx, nil, cli.Stdio{Out: &out, Err: &stderr})
	if err == nil || !strings.Contains(err.Error(), "seed runner failed") || out.String() != "Running 1 seed(s)...\n" || stderr.Len() != 0 {
		t.Fatal("standalone cancellation or streaming changed")
	}
	entries, err = os.ReadDir(".tmp")
	if err != nil || len(entries) != 0 {
		t.Fatal("canceled runner not cleaned")
	}
}

func TestDiscoverSeedFunctions_MissingDir(t *testing.T) {
	funcs, err := discoverSeedFunctions("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(funcs) != 0 {
		t.Errorf("expected 0 functions for missing dir, got %d", len(funcs))
	}
}
