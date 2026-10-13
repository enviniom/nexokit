package cli_test

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

// This subprocess stands in for Go: no server, initializer or database is run.
func TestMakeHelper(t *testing.T) {
	if os.Getenv("NEXOKIT_MAKE_HELPER") != "1" {
		return
	}
	args := strings.Split(os.Getenv("NEXOKIT_EXPECT_ARGS"), " ")
	separator := 0
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i + 1
			break
		}
	}
	if separator == 0 || !reflect.DeepEqual(os.Args[separator:], args) {
		t.Fatal("command arguments changed")
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil || string(input) != "synthetic input\n" {
		t.Fatal("stdin not inherited")
	}
	for _, key := range []string{"DB_PASSWORD", "DB_HOST", "APP_PORT", "CACHE_DRIVER"} {
		_, set := os.LookupEnv(key)
		if set {
			t.Fatalf("unexpected pre-load export: %s", key)
		}
	}
	if err := godotenv.Load(); err != nil {
		t.Fatal("synthetic dotenv did not load")
	}
	if os.Getenv("DB_PASSWORD") != `invented space "quote" # $dollar $(shell touch evaluated)` {
		t.Fatal("dotenv value changed")
	}
	if os.Getenv("DB_HOST") != "fixture-host" || os.Getenv("APP_PORT") != "8181" {
		t.Fatal("dotenv blocked by empty exports")
	}
	if os.Getenv("APP_NAME") != "external-name" {
		t.Fatal("external environment override lost")
	}
	if value, set := os.LookupEnv("APP_URL"); !set || value != "" {
		t.Fatal("explicit external empty override lost")
	}
}

func TestMakeDelegatesDotenv(t *testing.T) {
	if testing.Short() {
		t.Skip("external make subprocess")
	}
	makePath, err := exec.LookPath("make")
	if err != nil {
		t.Skip("make unavailable")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("Unix shell unavailable")
	}
	source, err := os.ReadFile("../../Makefile")
	if err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ goal, args string }{
		{"", "run ./cmd/nexokit serve"}, {"dev", "run ./cmd/nexokit serve"},
		{"init", "run ./cmd/nexokit init"}, {"migrate-up", "run ./cmd/nexokit migrate up"},
		{"migrate-down", "run ./cmd/nexokit migrate down"}, {"migrate-status", "run ./cmd/nexokit migrate status"},
		{"migrate-reset", "run ./cmd/nexokit migrate reset"},
	} {
		name := tc.goal
		if name == "" {
			name = "default-dev"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			for path, contents := range map[string]string{
				"Makefile": string(source),
				".env":     "APP_NAME=dotenv-name\nAPP_PORT=8181\nDB_HOST=fixture-host\nDB_PASSWORD='invented space \"quote\" # $dollar $(shell touch evaluated)'\n",
				"go":       "#!/bin/sh\nexec \"$NEXOKIT_TEST_BINARY\" -test.run='^TestMakeHelper$' -- \"$@\"\n",
			} {
				if err := os.WriteFile(filepath.Join(dir, path), []byte(contents), 0700); err != nil {
					t.Fatal(err)
				}
			}
			args := []string{"--no-print-directory", "-s"}
			if tc.goal != "" {
				args = append(args, tc.goal)
			}
			cmd := exec.Command(makePath, args...)
			cmd.Dir = dir
			cmd.Env = []string{"PATH=" + dir + ":" + os.Getenv("PATH"), "NEXOKIT_TEST_BINARY=" + binary, "NEXOKIT_MAKE_HELPER=1", "NEXOKIT_EXPECT_ARGS=" + tc.args, "APP_NAME=external-name", "APP_URL="}
			cmd.Stdin = strings.NewReader("synthetic input\n")
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("wrapper failed: %v\n%s", err, output)
			}
			if _, err := os.Stat(filepath.Join(dir, "evaluated")); !os.IsNotExist(err) {
				t.Fatal("dotenv executed as Make syntax")
			}
		})
	}
}
