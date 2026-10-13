package commands

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
)

func TestInitFlags(t *testing.T) {
	for _, args := range [][]string{nil, {"--name", "App"}, {"--module", "example.com/a"}, {"--name", "App", "--module", "example.com/a", "--apply", "--dry-run"}, {"--unknown"}, {"--name"}, {"--name", "App", "--module", "example.com/a", "extra"}, {"--name", "App", "--name", "Other", "--module", "example.com/a"}} {
		var out bytes.Buffer
		stdio := cli.Stdio{Out: &out, Err: &out}
		if len(args) > 0 {
			stdio.In = forbiddenInput{}
		}
		if e := (&InitCommand{}).Run(context.Background(), args, stdio); e == nil {
			t.Errorf("accepted %v", args)
		}
	}
}
func TestInitPreviewAndApply(t *testing.T) {
	root := t.TempDir()
	for path, text := range map[string]string{"go.mod": "module " + projectinit.OriginalModule() + "\n", "README.md": "# NexoKit\n", ".env.example": "APP_NAME=nexokit\n", "cmd/nexokit/main.go": "package main\nimport _ \"" + projectinit.OriginalModule() + "/internal/cli\"\n"} {
		full := filepath.Join(root, path)
		if e := os.MkdirAll(filepath.Dir(full), 0755); e != nil {
			t.Fatal(e)
		}
		if e := os.WriteFile(full, []byte(text), 0644); e != nil {
			t.Fatal(e)
		}
	}
	t.Chdir(root)
	args := []string{"--name", "App", "--module", "example.com/a"}
	for _, extra := range [][]string{nil, {"--dry-run"}, {"--apply"}} {
		var out bytes.Buffer
		if e := (&InitCommand{}).Run(context.Background(), append(append([]string{}, args...), extra...), cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &out}); e != nil {
			t.Fatal(e)
		}
		b, e := os.ReadFile("go.mod")
		if e != nil {
			t.Fatal(e)
		}
		if len(extra) > 0 && extra[0] == "--apply" {
			if !strings.Contains(string(b), "example.com/a") {
				t.Fatal("apply did not write")
			}
		} else {
			if !strings.Contains(string(b), projectinit.OriginalModule()) || !strings.Contains(out.String(), "Preview") {
				t.Fatal("preview wrote or missing output")
			}
		}
	}
	var out bytes.Buffer
	if e := (&InitCommand{}).Run(context.Background(), args, cli.Stdio{Out: &out, Err: &out}); e == nil {
		t.Fatal("repeat accepted")
	}
}
