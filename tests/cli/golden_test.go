package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli/generator"
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

	files := []string{
		"model.go",
		"dto.go",
		"repository.go",
		"service.go",
		"handler.go",
		"routes.go",
		"validation.go",
	}

	update := os.Getenv("UPDATE_GOLDEN") == "1"

	for _, f := range files {
		gotPath := filepath.Join(moduleDir, f)
		got, err := os.ReadFile(gotPath)
		if err != nil {
			t.Fatalf("reading generated %s: %v", f, err)
		}

		goldenPath := filepath.Join(goldenDir, f)
		if update {
			_ = os.MkdirAll(goldenDir, 0755)
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
			t.Errorf("golden mismatch for %s:\ngot:\n%s\nwant:\n%s", f, got, want)
		}
	}

	repository, err := os.ReadFile(filepath.Join(moduleDir, "repository.go"))
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
