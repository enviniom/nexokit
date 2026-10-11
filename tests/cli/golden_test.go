package cli_test

import (
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli/generator"
	"github.com/enviniom/nexokit/internal/cli/templates"
)

func TestGolden_ModuleWithAllFlags(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(cwd)

	_ = os.MkdirAll(filepath.Join(tmpDir, "internal", "modules"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "migrations"), 0755)

	opts := generator.ModuleOptions{
		Name:      "goldenmod",
		CRUD:      true,
		Migration: true,
		Tenant:    true,
	}
	if err := generator.GenerateModule(opts); err != nil {
		t.Fatalf("GenerateModule failed: %v", err)
	}

	goldenDir := filepath.Join(cwd, "testdata", "golden", "goldenmod")
	moduleDir := filepath.Join(tmpDir, "internal", "modules", "goldenmod")

	plan := templates.ModuleFiles(templates.ModuleData{Name: opts.Name, Package: opts.Name, CRUD: opts.CRUD, Tenant: opts.Tenant})

	update := os.Getenv("UPDATE_GOLDEN") == "1"

	// Keep retired snapshots as inert markers rather than deleting tracked files.
	// The negative assertion also protects against accidentally restoring flat output.
	retired := []string{"model.go", "dto.go", "repository.go", "service.go", "handler.go", "validation.go"}
	marker := []byte("package goldenmod\n\n// Retired flat scaffold snapshot. Current generated code lives in core/ and slices/.\n")
	for _, name := range retired {
		if _, err := os.Stat(filepath.Join(moduleDir, name)); !os.IsNotExist(err) {
			t.Fatalf("retired root implementation %s generated: %v", name, err)
		}
		path := filepath.Join(goldenDir, name)
		if update {
			if err := os.WriteFile(path, marker, 0644); err != nil {
				t.Fatal(err)
			}
		} else {
			got, err := os.ReadFile(path)
			if err != nil || string(got) != string(marker) {
				t.Fatalf("retired snapshot %s not updated: %v", name, err)
			}
		}
	}

	for _, file := range plan {
		f := file.Path
		gotPath := filepath.Join(moduleDir, f)
		got, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatalf("reading generated %s: %v", f, err)
		}

		goldenPath := filepath.Join(goldenDir, f)
		if update {
			if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(goldenPath, got, 0644); err != nil {
				t.Fatalf("writing golden %s: %v", f, err)
			}
			continue
		}

		want, err := os.ReadFile(goldenPath)
		if err != nil {
			t.Fatalf("reading golden %s: %v", f, err)
		}

		if string(got) != string(want) {
			formattedGot, err := format.Source(got)
			if err != nil {
				t.Fatalf("format generated %s: %v", f, err)
			}
			formattedWant, err := format.Source(want)
			if err != nil {
				t.Fatalf("format golden %s: %v", f, err)
			}
			if string(formattedGot) != string(formattedWant) {
				t.Errorf("golden mismatch for %s:\ngot:\n%s\nwant:\n%s", f, formattedGot, formattedWant)
			}
		}
	}

	repository, err := os.ReadFile(filepath.Join(moduleDir, "slices/list_goldenmod/repository.go"))
	if err != nil {
		t.Fatalf("reading generated repository.go: %v", err)
	}
	repositoryText := string(repository)
	for _, want := range []string{
		`"github.com/enviniom/nexokit/internal/platform/tenant"`,
		"tc tenant.TenantContext",
		"tenant.ApplyTenantScope(q, tc)",
	} {
		if !strings.Contains(repositoryText, want) {
			t.Errorf("generated tenant repository missing %q", want)
		}
	}
	if strings.Contains(repositoryText, `ctx.Value("company_id")`) {
		t.Error("generated tenant repository must not read company_id from context values")
	}

	// Verify migration was created
	entries, err := os.ReadDir(filepath.Join(tmpDir, "migrations"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") && strings.Contains(e.Name(), "goldenmod") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected migration file for goldenmod")
	}
}
