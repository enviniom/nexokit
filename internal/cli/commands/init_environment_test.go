package commands

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
	"github.com/joho/godotenv"
)

func TestInitCommandCreatesLocalEnvironment(t *testing.T) {
	initFixture(t)
	var out bytes.Buffer
	input := "My API\nuser-name\n\n\n\n\n  invented password  \n\n\n\n\ny\n"
	if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader(input), Out: &out, Err: &out}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(".env")
	if err != nil {
		t.Fatal(err)
	}
	env, err := godotenv.Unmarshal(string(b))
	if err != nil {
		t.Fatal(err)
	}
	if env["DB_PASSWORD"] != "  invented password  " || env["DB_HOST"] != "localhost" || env["APP_ENV"] != "development" || len(env["PASETO_KEY"]) != 64 {
		t.Fatal("incorrect local configuration")
	}
	if strings.Contains(out.String(), "invented password") || strings.Contains(out.String(), env["PASETO_KEY"]) {
		t.Fatal("secret disclosed")
	}
	info, err := os.Stat(".env")
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("unsafe permissions")
	}
}

func TestLocalEnvironmentValidationAndEntropy(t *testing.T) {
	for _, tc := range []struct {
		name        string
		fields      []string
		entropyFail bool
	}{
		{"port", []string{"localhost", "0"}, false},
		{"host credentials", []string{"user:invented@host"}, false},
		{"ssl", []string{"", "", "", "", "invalid"}, false},
		{"app port", []string{"", "", "", "", "", "65536"}, false},
		{"url credentials", []string{"", "", "", "", "", "", "http://user:invented@localhost"}, false},
		{"cors wildcard", []string{"", "", "", "", "", "", "", "*"}, false},
		{"cors malformed", []string{"", "", "", "", "", "", "", "http://localhost:3000,http://user:invented@localhost"}, false},
		{"control", []string{"bad\x00host"}, false},
		{"incomplete", nil, false},
		{"entropy", []string{"", "", "", "", "", "", "", ""}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			initFixture(t)
			var out bytes.Buffer
			index := 0
			read := func(string) (string, error) {
				if index >= len(tc.fields) {
					return "", io.EOF
				}
				v := tc.fields[index]
				index++
				return v, nil
			}
			var entropy io.Reader = strings.NewReader(strings.Repeat("a", 32))
			if tc.entropyFail {
				entropy = failingInput{}
			}
			_, err := collectLocalEnvironment(cli.Stdio{Out: &out}, "app", read, func() (string, error) { return "invented secret", nil }, entropy)
			if err == nil {
				t.Fatal("invalid configuration accepted")
			}
			if strings.Contains(err.Error()+out.String(), "invented") {
				t.Fatal("secret disclosed")
			}
			if _, err := os.Stat(".env"); !os.IsNotExist(err) {
				t.Fatal("configuration failure wrote")
			}
		})
	}
}

func TestInitCommandCancelAndEntropyFailure(t *testing.T) {
	for _, answer := range []string{"n\n", "\n", "", "y"} {
		t.Run("cancel "+answer, func(t *testing.T) {
			initFixture(t)
			var out bytes.Buffer
			input := "App\nuser\n" + strings.Repeat("\n", 9) + answer
			if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader(input), Out: &out}); err != nil {
				t.Fatal(err)
			}
			b, _ := os.ReadFile("go.mod")
			if !strings.Contains(string(b), projectinit.OriginalModule()) {
				t.Fatal("cancellation changed identity")
			}
			if _, err := os.Stat(".env"); !os.IsNotExist(err) {
				t.Fatal("cancellation wrote environment")
			}
		})
	}
	initFixture(t)
	original := initEntropy
	initEntropy = failingInput{}
	t.Cleanup(func() { initEntropy = original })
	var out bytes.Buffer
	err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader("App\nuser\n" + strings.Repeat("\n", 9) + "y\n"), Out: &out})
	if err == nil || strings.Contains(err.Error(), "input failed") {
		t.Fatal("entropy failure not safely handled")
	}
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		t.Fatal("entropy failure wrote")
	}
}

func TestCustomLocalEnvironmentFreshKeys(t *testing.T) {
	var previous string
	for attempt := 0; attempt < 2; attempt++ {
		fields := []string{"127.0.0.1", "5544", "custom_db", "custom_user", "verify-full", "9000", "http://localhost:9000", "http://localhost:3000,https://example.com"}
		index := 0
		var out bytes.Buffer
		values, err := collectLocalEnvironment(cli.Stdio{Out: &out}, "app", func(string) (string, error) { v := fields[index]; index++; return v, nil }, func() (string, error) { return " invented '$#\\ password ", nil }, rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		if values["DB_PORT"] != "5544" || values["APP_PORT"] != "9000" || values["DB_SSL_MODE"] != "verify-full" {
			t.Fatal("custom settings lost")
		}
		key := values["PASETO_KEY"]
		decoded, err := hex.DecodeString(key)
		if err != nil || len(decoded) != 32 || key == previous {
			t.Fatal("invalid/reused development key")
		}
		if strings.Contains(out.String(), key) || strings.Contains(out.String(), values["DB_PASSWORD"]) {
			t.Fatal("secret disclosed")
		}
		previous = key
	}
}

func TestInitCommandKeptEnvironmentSkipsWizardEntropy(t *testing.T) {
	initFixture(t)
	kept := "DB_HOST=localhost\nDB_PORT=5432\nDB_NAME=synthetic\nDB_USER=synthetic\nDB_PASSWORD=synthetic-only\nDB_SSL_MODE=disable\nPASETO_KEY=" + strings.Repeat("ab", 32) + "\n"
	if err := os.WriteFile(".env", []byte(kept), 0640); err != nil {
		t.Fatal(err)
	}
	previous := initEntropy
	initEntropy = failingInput{}
	t.Cleanup(func() { initEntropy = previous })
	var out bytes.Buffer
	if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader("App\nuser\nkeep\nn\n"), Out: &out}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Keeping existing") || strings.Contains(out.String(), "DB_PASSWORD") || strings.Contains(out.String(), "synthetic-only") {
		t.Fatal("kept environment wizard/secret handling incorrect")
	}
	data, err := os.ReadFile(".env")
	if err != nil || string(data) != kept {
		t.Fatal("kept environment bytes changed")
	}
	info, err := os.Stat(".env")
	if err != nil || info.Mode().Perm() != 0640 {
		t.Fatal("kept environment mode changed")
	}
}
