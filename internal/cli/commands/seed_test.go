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

func TestDiscoverSeedFunctions_MissingDir(t *testing.T) {
	funcs, err := discoverSeedFunctions("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(funcs) != 0 {
		t.Errorf("expected 0 functions for missing dir, got %d", len(funcs))
	}
}
