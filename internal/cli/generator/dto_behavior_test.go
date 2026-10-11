package generator

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/cli/templates"
)

// This focused regression executes the DTO template independently of layout,
// so a layout failure cannot conceal the normalization defect.
func TestGeneratedDTONormalization(t *testing.T) {
	if testing.Short() {
		t.Skip("generated DTO subprocess integration")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(cwd, "../../.."))
	dir := t.TempDir()
	for _, file := range []string{"go.mod", "go.sum"} {
		copyFixtureFile(t, root, dir, file)
	}
	for _, pkg := range []string{"shared", "platform/validator", "platform/messages", "platform/identity"} {
		rel := filepath.Join("internal", pkg)
		entries, err := os.ReadDir(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
				copyFixtureFile(t, root, dir, filepath.Join(rel, e.Name()))
			}
		}
	}
	target := filepath.Join(dir, "internal/modules/widgets/core/widgets")
	if err := os.MkdirAll(target, 0755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"dto", "model"} {
		tmpl, err := templates.ParseModuleTemplate(name)
		if err != nil {
			t.Fatal(err)
		}
		var out bytes.Buffer
		if err := tmpl.Execute(&out, templates.ModuleData{Name: "widgets", Package: "widgets", Struct: "Widgets", Table: "widgets", Tenant: true}); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(target, name+".go"), out.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeFixtureTest(t, dir, "internal/modules/widgets/core/widgets/normalization_test.go", `package widgets
import "testing"
func TestNormalize(t *testing.T) {
 c:=CreateWidgetsRequest{Name:"  Alpha  "}; u:=UpdateWidgetsRequest{Name:"  Beta  "}
 c.Validate();u.Validate()
 if c.Name!="Alpha" || u.Name!="Beta" {t.Fatalf("normalization did not persist: %q %q",c.Name,u.Name)}
 c.Name="   "; if !c.Validate().HasErrors() {t.Fatal("whitespace-only name accepted")}
}
`, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "-count=1", "-timeout", "60s", "./internal/modules/widgets/core/widgets")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated DTO: %v\n%s", err, output)
	}
}
