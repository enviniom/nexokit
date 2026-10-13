package commands

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
)

func initFixture(t *testing.T) {
	t.Helper()
	root := t.TempDir()
	for path, text := range map[string]string{"go.mod": "module " + projectinit.OriginalModule() + "\n", "README.md": "# NexoKit\n", ".env.example": "APP_NAME=nexokit\n", "cmd/nexokit/main.go": "package main\nimport _ \"" + projectinit.OriginalModule() + "/internal/cli\"\n"} {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(root)
}

func TestInitInteractive(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		apply, fail bool
	}{
		{"yes", "My API\nuser-name\n Y \n", true, false},
		{"no", "My API\nuser-name\nn\n", false, false},
		{"blank", "My API\nuser-name\n\n", false, false},
		{"confirmation EOF", "My API\nuser-name\n", false, false},
		{"unterminated yes", "My API\nuser-name\ny", false, false},
		{"unknown", "My API\nuser-name\nmaybe\n", false, true},
		{"missing name", "\n", false, true},
		{"missing user", "My API\n", false, true},
		{"unsafe name", "My/API\nuser\ny\n", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initFixture(t)
			var out bytes.Buffer
			err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader(tc.input), Out: &out, Err: &out})
			if (err != nil) != tc.fail {
				t.Fatalf("error = %v; output = %s", err, &out)
			}
			b, err := os.ReadFile("go.mod")
			if err != nil {
				t.Fatal(err)
			}
			want := projectinit.OriginalModule()
			if tc.apply {
				want = "github.com/user-name/my-api"
			}
			if string(b) != "module "+want+"\n" {
				t.Fatalf("module: %s", b)
			}
			if !tc.fail {
				for _, text := range []string{"Application display name", "GitHub username", "github.com/user-name/my-api", "Preview", "[y/N]"} {
					if !strings.Contains(out.String(), text) {
						t.Errorf("missing %q in %s", text, &out)
					}
				}
			}
			if tc.apply {
				b, err = os.ReadFile("README.md")
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(b), "My API") {
					t.Fatalf("display name lost: %s", b)
				}
				if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader(tc.input), Out: &out, Err: &out}); err == nil {
					t.Fatal("repeat accepted")
				}
			}
		})
	}
}

type failingInput struct{}

func (failingInput) Read([]byte) (int, error) { return 0, errors.New("input failed") }

type cancelingInput struct{ cancel context.CancelFunc }

func (r cancelingInput) Read(b []byte) (int, error) {
	r.cancel()
	return copy(b, "My API\nuser-name\ny\n"), nil
}

type forbiddenInput struct{}

func (forbiddenInput) Read([]byte) (int, error) { panic("flag command read stdin") }

func TestInitInteractiveFailures(t *testing.T) {
	for _, username := range []string{"-user", "user-", "u--ser", "@user", "user/name", "https://github.com/user", "üser", strings.Repeat("a", 40), "u\tser"} {
		t.Run(username, func(t *testing.T) {
			initFixture(t)
			var out bytes.Buffer
			if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader("My API\n" + username + "\ny\n"), Out: &out, Err: &out}); err == nil {
				t.Fatal("invalid username accepted")
			}
			b, _ := os.ReadFile("go.mod")
			if !strings.Contains(string(b), projectinit.OriginalModule()) {
				t.Fatal("failure wrote")
			}
		})
	}
	initFixture(t)
	var out bytes.Buffer
	if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: failingInput{}, Out: &out, Err: &out}); err == nil {
		t.Fatal("read failure accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := (&InitCommand{}).Run(ctx, nil, cli.Stdio{In: cancelingInput{cancel}, Out: &out, Err: &out}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation during read: %v", err)
	}
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	if err := (&InitCommand{}).Run(ctx, nil, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &out}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation: %v", err)
	}
	for _, args := range [][]string{{"--name", "App"}, {"--dry-run"}, {"--name", "App", "--module", "example.com/a"}} {
		if err := (&InitCommand{}).Run(context.Background(), args, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &out}); (err == nil) != (len(args) == 4) {
			t.Fatalf("flags %v: %v", args, err)
		}
	}
	if err := os.WriteFile("go.mod", []byte("module example.com/other\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader("App\nuser\ny\n"), Out: &out, Err: &out}); err == nil {
		t.Fatal("invalid root accepted")
	}
}

func TestRepositorySlug(t *testing.T) {
	for _, name := range []string{"My API", "..My__API--", "My ._- API"} {
		slug, err := repositorySlug(name)
		if err != nil || slug != "my-api" {
			t.Fatalf("%q: %q, %v", name, slug, err)
		}
	}
	for _, name := range []string{"", "._-", "é", "a/b", "a\tb"} {
		if _, err := repositorySlug(name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
}

// Capture before command tests that deliberately clear the process environment.
var interactiveSmokeNested = os.Getenv("NEXOKIT_INIT_SMOKE") == "1"
var interactiveSmokeEnv = os.Environ()
var interactiveSmokeGo, interactiveSmokeGoError = exec.LookPath("go")

// Copy only known source families, never the original environment files.
func TestInteractiveSourceCopy(t *testing.T) {
	if testing.Short() || interactiveSmokeNested {
		t.Skip("external-command source-copy smoke")
	}
	if interactiveSmokeGoError != nil {
		t.Fatal(interactiveSmokeGoError)
	}
	source, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	copyFile := func(path, rel string) error {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		return os.WriteFile(dest, b, 0644)
	}
	for _, dir := range []string{"cmd", "internal", "tests", "seeds", "docs", "migrations", "scripts"} {
		err := filepath.WalkDir(filepath.Join(source, dir), func(path string, d fs.DirEntry, err error) error {
			if os.IsNotExist(err) {
				return nil
			}
			if err != nil {
				return err
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "logs" || name == "tmp" || name == "temp" || name == "bin" || strings.HasSuffix(name, ".log") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 || d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			return copyFile(path, rel)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []string{"go.mod", "go.sum", "README.md", "Makefile"} {
		if err := copyFile(filepath.Join(source, file), file); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(target, ".env.example"), []byte("APP_NAME=nexokit\n"), 0644); err != nil {
		t.Fatal(err)
	}
	run := func(input string, success bool, args ...string) string {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, interactiveSmokeGo, args...)
		cmd.Dir = target
		cmd.Env = append(append([]string{}, interactiveSmokeEnv...), "GOWORK=off", "GOPROXY=off", "GOSUMDB=off", "NEXOKIT_INIT_SMOKE=1")
		cmd.Stdin = strings.NewReader(input)
		output, err := cmd.CombinedOutput()
		if (err == nil) != success {
			t.Fatalf("go %v: %v\n%s", args, err, output)
		}
		t.Logf("go %v: success=%v", args, success)
		return string(output)
	}
	identity := map[string][]byte{}
	for _, path := range []string{"go.mod", "README.md", ".env.example", "cmd/nexokit/main.go"} {
		b, err := os.ReadFile(filepath.Join(target, path))
		if err != nil {
			t.Fatal(err)
		}
		identity[path] = b
	}
	for _, answer := range []string{"n\n", "\n"} {
		run("My API\nuser-name\n"+answer, true, "run", "./cmd/nexokit", "init")
		for path, original := range identity {
			after, err := os.ReadFile(filepath.Join(target, path))
			if err != nil || !bytes.Equal(original, after) {
				t.Fatalf("cancel changed %s", path)
			}
		}
	}
	run("My API\nuser-name\ny\n", true, "run", "./cmd/nexokit", "init")
	after, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil || !strings.Contains(string(after), "module github.com/user-name/my-api\n") {
		t.Fatal("module not applied")
	}
	readme, err := os.ReadFile(filepath.Join(target, "README.md"))
	if err != nil || !strings.HasPrefix(string(readme), "# My API\n") {
		t.Fatal("name not applied")
	}
	run("", true, "run", "./cmd/nexokit", "help")
	run("My API\nuser-name\ny\n", false, "run", "./cmd/nexokit", "init")
	run("", true, "build", "./...")
	output := run("", true, "test", "-v", "./...")
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "SKIP") || strings.Contains(line, "Redis") {
			t.Log(line)
		}
	}
}
