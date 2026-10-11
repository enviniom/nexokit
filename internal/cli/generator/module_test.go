package generator

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple", "products", true},
		{"with underscore", "order_items", true},
		{"with digits", "v2_products", true},
		{"empty", "", false},
		{"uppercase", "Products", false},
		{"hyphen", "order-items", false},
		{"space", "order items", false},
		{"leading digit", "2products", false},
		{"special char", "items@table", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if tt.want && err != nil {
				t.Errorf("ValidateName(%q) unexpected error: %v", tt.input, err)
			}
			if !tt.want && err == nil {
				t.Errorf("ValidateName(%q) expected error, got nil", tt.input)
			}
		})
	}
}

func TestGenerateModule_PathSafety(t *testing.T) {
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)
	moduleDir := filepath.Join("internal", "modules", "testmod")
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(moduleDir, "keep.txt")
	if err := os.WriteFile(marker, []byte("preexisting"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"testmod", " testmod "} {
		if err := GenerateModule(ModuleOptions{Name: name}); !errors.Is(err, ErrModuleExists) {
			t.Fatalf("existing module %q: %v", name, err)
		}
	}
	for _, name := range []string{"../escape", "a/b", "/absolute", ".", ".."} {
		if err := GenerateModule(ModuleOptions{Name: name}); err == nil {
			t.Errorf("unsafe path accepted: %q", name)
		}
	}
	if got, err := os.ReadFile(marker); err != nil || string(got) != "preexisting" {
		t.Fatalf("existing content changed: %q %v", got, err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "internal" {
		t.Fatalf("unexpected writes: %v", entries)
	}
}

func TestGenerateModule_TemplateExecution(t *testing.T) {
	// Use a temporary directory by changing working directory.
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
	if err := os.MkdirAll(filepath.Join(tmpDir, "migrations"), 0755); err != nil {
		t.Fatal(err)
	}

	opts := ModuleOptions{
		Name:      "testmod",
		CRUD:      true,
		Migration: false,
		Tenant:    false,
	}

	if err := GenerateModule(opts); err != nil {
		t.Fatalf("GenerateModule failed: %v", err)
	}

	expectedFiles := []string{
		"core/testmod/model.go",
		"core/testmod/dtos.go",
		"core/contracts.go",
		"core/errors.go",
		"container.go",
		"routes.go",
		"slices/create_testmod/handler.go",
		"slices/create_testmod/repository.go",
		"slices/get_testmod/handler.go",
		"slices/list_testmod/repository.go",
		"slices/update_testmod/repository.go",
		"slices/delete_testmod/repository.go",
	}
	moduleDir := filepath.Join(tmpDir, "internal", "modules", "testmod")
	for _, f := range expectedFiles {
		path := filepath.Join(moduleDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", path)
		}
	}

	// Verify idempotency: running again should fail
	if err := GenerateModule(opts); err == nil {
		t.Error("expected error on duplicate module generation, got nil")
	} else if !errors.Is(err, ErrModuleExists) {
		t.Logf("got expected idempotency error: %v", err)
	}
}

func TestGenerateModule_WithMigration(t *testing.T) {
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
	if err := os.MkdirAll(filepath.Join(tmpDir, "migrations"), 0755); err != nil {
		t.Fatal(err)
	}

	opts := ModuleOptions{
		Name:      "migmod",
		CRUD:      false,
		Migration: true,
		Tenant:    false,
	}

	if err := GenerateModule(opts); err != nil {
		t.Fatalf("GenerateModule failed: %v", err)
	}

	moduleDir := filepath.Join(tmpDir, "internal", "modules", "migmod")
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		t.Fatalf("module dir not created")
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
