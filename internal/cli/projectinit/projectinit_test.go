package projectinit

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func put(t *testing.T, root, path, text string) {
	t.Helper()
	p := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0640); err != nil {
		t.Fatal(err)
	}
}
func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	put(t, root, "go.mod", "module "+OriginalModule()+"\n\ngo 1.26\n")
	put(t, root, "README.md", "# NexoKit\n\nNexoKit tool attribution.\n")
	// User-provided, nonsecret anchor; never copied from the original example.
	put(t, root, ".env.example", "APP_NAME=nexokit\n# invented fixture comment\nUNRELATED=fixture\n")
	put(t, root, "cmd/nexokit/main.go", "package main\nimport _ \""+OriginalModule()+"/internal/cli\"\n")
	put(t, root, "internal/cli/templates/example.tmpl", "import \""+OriginalModule()+"/internal/shared\"\n")
	put(t, root, "internal/fixture_test.go", "package internal\nconst source = `import \""+OriginalModule()+"/seeds\"`\n")
	return root
}
func get(t *testing.T, root, path string) string {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(root, path))
	if e != nil {
		t.Fatal(e)
	}
	return string(b)
}
func TestPlanPreviewApply(t *testing.T) {
	root := fixture(t)
	put(t, root, ".env", "invented private fixture")
	put(t, root, "odd/task.go", OriginalModule())
	put(t, root, "vendor/x.go", OriginalModule())
	put(t, root, "internal/ignored.bin", OriginalModule())
	put(t, root, "go.sum", "untouched fixture sum")
	before := get(t, root, "go.mod")
	p, e := New(root, "Starter Smoke", "example.com/team/api/v2")
	if e != nil {
		t.Fatal(e)
	}
	var paths []string
	for _, c := range p.Changes {
		paths = append(paths, c.Path)
	}
	want := []string{".env.example", "README.md", "cmd/nexokit/main.go", "go.mod", "internal/cli/templates/example.tmpl", "internal/fixture_test.go"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths %v", paths)
	}
	if get(t, root, "go.mod") != before {
		t.Fatal("preview wrote files")
	}
	if e = p.Apply(); e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(get(t, root, "go.mod"), "module example.com/team/api/v2") {
		t.Fatal("module not rewritten")
	}
	if get(t, root, ".env.example") != "APP_NAME=\"Starter Smoke\"\n# invented fixture comment\nUNRELATED=fixture\n" {
		t.Fatal("example preservation/quoting failed")
	}
	if get(t, root, "README.md") != "# Starter Smoke\n\nNexoKit tool attribution.\n" {
		t.Fatal("attribution changed")
	}
	for _, path := range []string{"cmd/nexokit/main.go", "internal/cli/templates/example.tmpl", "internal/fixture_test.go"} {
		if strings.Contains(get(t, root, path), OriginalModule()) {
			t.Fatalf("residual %s", path)
		}
	}
	for path, text := range map[string]string{".env": "invented private fixture", "odd/task.go": OriginalModule(), "vendor/x.go": OriginalModule(), "internal/ignored.bin": OriginalModule(), "go.sum": "untouched fixture sum"} {
		if get(t, root, path) != text {
			t.Fatalf("excluded %s changed", path)
		}
	}
	info, _ := os.Stat(filepath.Join(root, "go.mod"))
	if info.Mode().Perm() != 0640 {
		t.Fatal("mode changed")
	}
	if _, e = New(root, "Again", "example.com/again"); e == nil {
		t.Fatal("repeat accepted")
	}
	if e = p.Apply(); e == nil {
		t.Fatal("stale plan accepted")
	}
}
func TestInvalidIdentity(t *testing.T) {
	for _, tc := range []struct{ name, module string }{{"", "example.com/a"}, {"x\ny", "example.com/a"}, {"../x", "example.com/a"}, {"x", "../escape"}, {"x", "example.com/a\\b"}, {"x", "example.com/a\n"}, {"x", "example.com/a/../b"}, {"x", OriginalModule()}, {"x", "example.com/a/v1"}} {
		t.Run(tc.name+tc.module, func(t *testing.T) {
			if _, e := New(fixture(t), tc.name, tc.module); e == nil {
				t.Fatal("unsafe identity accepted")
			}
		})
	}
	for _, path := range []string{"go.mod", "README.md", ".env.example", "cmd/nexokit/main.go"} {
		t.Run(path, func(t *testing.T) {
			root := fixture(t)
			put(t, root, path, "partial identity\n")
			if _, e := New(root, "App", "example.com/a"); e == nil {
				t.Fatal("partial identity accepted")
			}
		})
	}
	if _, e := New(t.TempDir(), "App", "example.com/a"); e == nil {
		t.Fatal("invalid root accepted")
	}
}
func TestPartialSourceIdentity(t *testing.T) {
	root := fixture(t)
	put(t, root, "internal/partial.go", "package partial\nimport _ \""+"example.com/renamed"+"/internal/shared\"\n")
	if _, e := New(root, "App", "example.com/a"); e == nil {
		t.Fatal("partial source identity accepted")
	}
}
func TestPreflight(t *testing.T) {
	for _, action := range []string{"changed", "missing", "mode", "symlink"} {
		t.Run(action, func(t *testing.T) {
			root := fixture(t)
			p, e := New(root, "App", "example.com/a")
			if e != nil {
				t.Fatal(e)
			}
			before := get(t, root, "README.md")
			path := filepath.Join(root, "cmd/nexokit/main.go")
			switch action {
			case "changed":
				put(t, root, "cmd/nexokit/main.go", "changed")
			case "missing":
				if e = os.Rename(path, path+".moved"); e != nil {
					t.Fatal(e)
				}
			case "mode":
				if e = os.Chmod(path, 0600); e != nil {
					t.Fatal(e)
				}
			case "symlink":
				if e = os.Rename(path, path+".moved"); e != nil {
					t.Fatal(e)
				}
				if e = os.Symlink(path+".moved", path); e != nil {
					t.Fatal(e)
				}
			}
			if e = p.Apply(); e == nil {
				t.Fatal("unsafe apply accepted")
			}
			if get(t, root, "README.md") != before {
				t.Fatal("partial writes")
			}
		})
	}
}
func TestUnchangedCandidatePreflight(t *testing.T) {
	root := fixture(t)
	put(t, root, "internal/unchanged.go", "package unchanged\n")
	p, e := New(root, "App", "example.com/a")
	if e != nil {
		t.Fatal(e)
	}
	put(t, root, "internal/unchanged.go", "package changed\n")
	before := get(t, root, "go.mod")
	if e = p.Apply(); e == nil {
		t.Fatal("unchanged candidate mutation ignored")
	}
	if get(t, root, "go.mod") != before {
		t.Fatal("preflight failure wrote module")
	}
}
func TestRootSymlink(t *testing.T) {
	root := fixture(t)
	link := filepath.Join(t.TempDir(), "linked-root")
	if e := os.Symlink(root, link); e != nil {
		t.Fatal(e)
	}
	before := get(t, root, "go.mod")
	if _, e := New(link, "App", "example.com/a"); e == nil {
		t.Fatal("symlink root accepted")
	}
	if get(t, root, "go.mod") != before {
		t.Fatal("outside root mutated")
	}
}
func TestApplyRollback(t *testing.T) {
	root := fixture(t)
	p, e := New(root, "App", "example.com/a")
	if e != nil {
		t.Fatal(e)
	}
	before := map[string]string{}
	for _, c := range p.Changes {
		before[c.Path] = get(t, root, c.Path)
	}
	calls := 0
	e = p.apply(func(a, b string) error {
		calls++
		if calls == 3 {
			return errors.New("injected rename failure")
		}
		return os.Rename(a, b)
	})
	if e == nil || !strings.Contains(e.Error(), "injected rename failure") {
		t.Fatalf("missing failure: %v", e)
	}
	for path, text := range before {
		if get(t, root, path) != text {
			t.Fatalf("rollback failed: %s", path)
		}
	}
	matches, _ := filepath.Glob(filepath.Join(root, ".nexokit-init-*"))
	if len(matches) != 0 {
		t.Fatalf("staged files left: %v", matches)
	}
}
func TestApplyRollbackFailurePreservesRecovery(t *testing.T) {
	root := fixture(t)
	outside := t.TempDir()
	put(t, outside, "marker", "outside fixture")
	p, e := New(root, "App", "example.com/a")
	if e != nil {
		t.Fatal(e)
	}
	calls := 0
	e = p.apply(func(a, b string) error {
		calls++
		if calls == 2 {
			first := filepath.Join(root, ".env.example")
			if err := os.Rename(first, first+".moved"); err != nil {
				return err
			}
			if err := os.Symlink(filepath.Join(outside, "marker"), first); err != nil {
				return err
			}
			return errors.New("injected commit failure")
		}
		return os.Rename(a, b)
	})
	if e == nil || !strings.Contains(e.Error(), "ROLLBACK FAILED") || !strings.Contains(e.Error(), "recovery copy") {
		t.Fatalf("missing recovery error: %v", e)
	}
	matches, err := filepath.Glob(filepath.Join(root, ".nexokit-init-*"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("missing recovery copy: %v %v", matches, err)
	}
	if b, err := os.ReadFile(matches[0]); err != nil || !strings.HasPrefix(string(b), "APP_NAME=nexokit\n") {
		t.Fatalf("bad recovery copy: %v", err)
	}
	if get(t, outside, "marker") != "outside fixture" {
		t.Fatal("rollback followed symlink")
	}
}
func TestSymlinkEscape(t *testing.T) {
	for _, path := range []string{"go.mod", "README.md", ".env.example", "internal", "cmd/nexokit/main.go"} {
		t.Run(path, func(t *testing.T) {
			root := fixture(t)
			outside := t.TempDir()
			put(t, outside, "marker.go", OriginalModule())
			target := filepath.Join(root, path)
			if e := os.Rename(target, target+".moved"); e != nil {
				t.Fatal(e)
			}
			destination := filepath.Join(outside, "marker.go")
			if path == "internal" {
				destination = outside
			}
			if e := os.Symlink(destination, target); e != nil {
				t.Fatal(e)
			}
			if _, e := New(root, "App", "example.com/a"); e == nil {
				t.Fatal("symlink accepted")
			}
			if get(t, outside, "marker.go") != OriginalModule() {
				t.Fatal("outside mutation")
			}
		})
	}
}
