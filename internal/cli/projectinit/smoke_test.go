package projectinit

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestInitializedSourceCopy deliberately excludes the protected original example.
// Only the user-supplied APP_NAME line is synthesized in the disposable copy.
func TestInitializedSourceCopy(t *testing.T) {
	if testing.Short() || os.Getenv("NEXOKIT_INIT_SMOKE") == "1" {
		t.Skip("external-command source-copy smoke")
	}
	source, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	mod, err := os.ReadFile(filepath.Join(source, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	currentModule := ""
	for _, line := range strings.Split(string(mod), "\n") {
		fields := strings.Fields(strings.SplitN(line, "//", 2)[0])
		if len(fields) == 0 || fields[0] != "module" {
			continue
		}
		if len(fields) != 2 || currentModule != "" {
			t.Fatal("malformed module directive in source go.mod")
		}
		currentModule = fields[1]
		if strings.HasPrefix(currentModule, "\"") || strings.HasPrefix(currentModule, "`") {
			currentModule, err = strconv.Unquote(currentModule)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if currentModule == "" || strings.ContainsAny(currentModule, " \t\r\n\\\"`") {
		t.Fatal("missing or malformed source module")
	}
	for _, part := range strings.Split(currentModule, "/") {
		if !regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~-]*$`).MatchString(part) || strings.Contains(part, "..") || strings.HasSuffix(part, ".") {
			t.Fatal("malformed source module path")
		}
	}
	if currentModule != OriginalModule() {
		t.Skip("starter adoption smoke requires canonical source")
	}
	target := t.TempDir()
	copyFile := func(path, rel string) error {
		b, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		info, e := os.Stat(path)
		if e != nil {
			return e
		}
		dest := filepath.Join(target, rel)
		if e = os.MkdirAll(filepath.Dir(dest), 0755); e != nil {
			return e
		}
		return os.WriteFile(dest, b, info.Mode().Perm())
	}
	// No blanket archive/copy: only these known project families and root files.
	for _, dir := range []string{"cmd", "internal", "tests", "seeds", "docs", "migrations", "scripts"} {
		base := filepath.Join(source, dir)
		e = filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if os.IsNotExist(err) && path == base {
				return nil
			}
			if err != nil {
				return err
			}
			if excluded(d.Name()) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			rel, e := filepath.Rel(source, path)
			if e != nil {
				return e
			}
			return copyFile(path, rel)
		})
		if e != nil {
			t.Fatal(e)
		}
	}
	for _, file := range []string{"go.mod", "go.sum", "README.md", "Makefile"} {
		if e = copyFile(filepath.Join(source, file), file); e != nil {
			t.Fatal(e)
		}
	}
	put(t, target, ".env.example", "APP_NAME=nexokit\n")
	run := func(wantSuccess bool, args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Dir = target
		cmd.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "NEXOKIT_INIT_SMOKE=1")
		output, e := cmd.CombinedOutput()
		if (e == nil) != wantSuccess {
			t.Fatalf("go %v: %v\n%s", args, e, output)
		}
		t.Logf("go %v: expected success=%v (%d output bytes)", args, wantSuccess, len(output))
		return string(output)
	}
	args := []string{"run", "./cmd/nexokit", "init", "--name", "Starter Smoke", "--module", "example.com/team/starter-smoke"}
	before := get(t, target, "go.mod")
	run(true, args...)
	if get(t, target, "go.mod") != before {
		t.Fatal("CLI preview wrote")
	}
	run(true, append(args, "--apply")...)
	run(true, "run", "./cmd/nexokit", "help")
	output := run(false, args...)
	if !strings.Contains(output, "starter identity") && !strings.Contains(output, "not an uninitialized NexoKit module") {
		t.Fatal("repeat did not fail identity check")
	}
	for _, dir := range []string{"cmd", "internal", "tests", "seeds"} {
		e = filepath.WalkDir(filepath.Join(target, dir), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".go" && ext != ".tmpl" {
				return nil
			}
			b, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			if strings.Contains(string(b), OriginalModule()) {
				t.Errorf("original project reference remains: %s", path)
			}
			return nil
		})
		if e != nil {
			t.Fatal(e)
		}
	}
	run(true, "build", "./...")
	run(true, "test", "./...")
}
