package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
)

func TestMakeCommand_RunModule(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "internal", "modules"), 0755); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	stdio := cli.Stdio{Out: &out}
	cmd := &MakeCommand{}

	err = cmd.Run(context.Background(), []string{"module", "testmake"}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("generated successfully")) {
		t.Errorf("expected success message, got %q", out.String())
	}

	moduleDir := filepath.Join(tmpDir, "internal", "modules", "testmake")
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		t.Fatalf("module dir not created")
	}
}

func TestMakeCommand_RunModule_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "internal", "modules", "dupmake"), 0755); err != nil {
		t.Fatal(err)
	}

	stdio := cli.Stdio{}
	cmd := &MakeCommand{}

	err = cmd.Run(context.Background(), []string{"module", "dupmake"}, stdio)
	if err == nil {
		t.Fatal("expected error for existing module")
	}
}

func TestMakeCommand_RunMigration(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmpDir, "migrations"), 0755); err != nil {
		t.Fatal(err)
	}

	stdio := cli.Stdio{}
	cmd := &MakeCommand{}

	err = cmd.Run(context.Background(), []string{"migration", "test_migration"}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmpDir, "migrations"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if !e.IsDir() {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected migration file to be created")
	}
}

func TestMakeCommand_RunSeed(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	stdio := cli.Stdio{}
	cmd := &MakeCommand{}

	err = cmd.Run(context.Background(), []string{"seed", "demo_data"}, stdio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(tmpDir, "seeds"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Error("expected seed file to be created")
	}
}

func TestMakeCommand_UnknownSubcommand(t *testing.T) {
	stdio := cli.Stdio{}
	cmd := &MakeCommand{}

	err := cmd.Run(context.Background(), []string{"unknown"}, stdio)
	if err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}
