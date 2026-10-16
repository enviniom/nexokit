package commands

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
	"github.com/joho/godotenv"
)

func initSetupProcessSettings(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{"DB_HOST": "selected.invalid", "DB_PORT": "5440", "DB_NAME": "selected_db", "DB_USER": "selected_user", "DB_PASSWORD": "SyntheticDBOnly9", "DB_SSL_MODE": "disable", "APP_PORT": "8081", "APP_URL": "http://localhost:8081", "CORS_ALLOWED_ORIGINS": "http://localhost:3001"} {
		t.Setenv(key, value)
	}
}

const initSetupKept = "DB_HOST=selected.invalid\nDB_PORT=5440\nDB_NAME=owned\nDB_USER=fake\nDB_PASSWORD=SyntheticDBOnly9\nDB_SSL_MODE=disable\nPASETO_KEY=abababababababababababababababababababababababababababababababab\n"

type initSetupCountingInput struct{ reads int }

func (r *initSetupCountingInput) Read([]byte) (int, error) {
	r.reads++
	return 0, io.EOF
}

func TestInitSetupAdoptedRejectsBeforeReadingInput(t *testing.T) {
	initFixture(t)
	if err := os.WriteFile("go.mod", []byte("module example.com/adopted\n"), 0644); err != nil {
		t.Fatal(err)
	}
	input := &initSetupCountingInput{}
	var out, stderr bytes.Buffer
	err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: input, Out: &out, Err: &stderr})
	if err == nil || input.reads != 0 || out.Len() != 0 || stderr.Len() != 0 {
		t.Fatal("adopted init read input or emitted prompts")
	}
}

func TestInitSetupInvalidSettingsAreWriteless(t *testing.T) {
	for _, field := range []string{"DB_PORT", "DB_SSL_MODE", "PASETO_KEY", "ACCESS_TTL_MINUTES"} {
		t.Run(field, func(t *testing.T) {
			initFixture(t)
			values, err := godotenv.Unmarshal(initSetupKept)
			if err != nil {
				t.Fatal(err)
			}
			values[field] = "invalid-secret"
			data, err := godotenv.Marshal(values)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(".env", []byte(data), 0640); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile("go.mod")
			if err != nil {
				t.Fatal(err)
			}
			var out, stderr bytes.Buffer
			err = (&InitCommand{}).Run(context.Background(), []string{"--name", "App", "--module", "example.com/app", "--apply", "--env-action", "keep"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr})
			if err == nil {
				t.Fatal("invalid settings accepted")
			}
			after, readErr := os.ReadFile("go.mod")
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatal("invalid settings changed identity")
			}
			kept, readErr := os.ReadFile(".env")
			if readErr != nil || string(kept) != data {
				t.Fatal("invalid settings changed environment")
			}
			if strings.Contains(out.String()+stderr.String()+err.Error(), "invalid-secret") {
				t.Fatal("secret leaked")
			}
		})
	}
}

func TestInitSetupMissingSettingsAndEnvironmentChoice(t *testing.T) {
	for _, scenario := range []string{"missing-settings", "missing-choice", "keep-absent"} {
		t.Run(scenario, func(t *testing.T) {
			initFixture(t)
			initSetupProcessSettings(t)
			args := []string{"--name", "App", "--module", "example.com/app", "--apply"}
			if scenario == "missing-settings" {
				t.Setenv("DB_HOST", "")
			}
			if scenario == "keep-absent" {
				args = append(args, "--env-action", "keep")
			}
			if scenario == "missing-choice" {
				if err := os.WriteFile(".env", []byte(initSetupKept), 0640); err != nil {
					t.Fatal(err)
				}
			}
			before, err := os.ReadFile("go.mod")
			if err != nil {
				t.Fatal(err)
			}
			var out, stderr bytes.Buffer
			err = (&InitCommand{}).Run(context.Background(), args, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr})
			if err == nil || out.Len() != 0 || stderr.Len() != 0 {
				t.Fatal("missing settings/choice accepted or advertised success")
			}
			after, readErr := os.ReadFile("go.mod")
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatal("rejection changed identity")
			}
			if scenario == "missing-choice" {
				kept, readErr := os.ReadFile(".env")
				if readErr != nil || string(kept) != initSetupKept {
					t.Fatal("rejection changed environment")
				}
			} else if _, statErr := os.Stat(".env"); !os.IsNotExist(statErr) {
				t.Fatal("rejection wrote environment")
			}
		})
	}
}

func TestInitSetupGeneratedEnvironment(t *testing.T) {
	for _, rebuild := range []bool{false, true} {
		name := "new"
		if rebuild {
			name = "rebuild"
		}
		t.Run(name, func(t *testing.T) {
			initFixture(t)
			initSetupProcessSettings(t)
			t.Setenv("ROOT_USER_EMAIL", "invalid-root-must-be-ignored")
			t.Setenv("ROOT_USER_PASSWORD", "do-not-store-root")
			args := []string{"--name", "App", "--module", "example.com/app", "--apply"}
			if rebuild {
				if err := os.WriteFile(".env", []byte(initSetupKept), 0640); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--env-action", "rebuild")
			}
			var out, stderr bytes.Buffer
			if err := (&InitCommand{}).Run(context.Background(), args, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr}); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(".env")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data)+out.String()+stderr.String(), "do-not-store-root") {
				t.Fatal("root input leaked")
			}
			values, err := godotenv.Unmarshal(string(data))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := validateInitSetup(values); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "make dev") || stderr.Len() != 0 {
				t.Fatal("missing manual workflow")
			}
			info, err := os.Stat(".env")
			if err != nil || info.Mode().Perm() != 0600 {
				t.Fatal("unsafe generated environment mode")
			}
			entries, err := os.ReadDir(".")
			if err != nil {
				t.Fatal(err)
			}
			backups := 0
			for _, entry := range entries {
				if !strings.HasPrefix(entry.Name(), ".env.backup-") {
					continue
				}
				backups++
				data, err := os.ReadFile(entry.Name())
				if err != nil {
					t.Fatal(err)
				}
				info, err := entry.Info()
				if err != nil || string(data) != initSetupKept || info.Mode().Perm() != 0600 || !strings.Contains(out.String(), entry.Name()) {
					t.Fatal("backup bytes, protection or preview changed")
				}
			}
			if (backups == 1) != rebuild || backups > 1 {
				t.Fatal("unexpected backup count")
			}
		})
	}
}

func TestInitSetupInteractiveCancellationAndChoice(t *testing.T) {
	for _, answer := range []string{"\n", "n\n", "no\n", "", "y", "invalid\n"} {
		t.Run("answer_"+answer, func(t *testing.T) {
			initFixture(t)
			if err := os.WriteFile(".env", []byte(initSetupKept), 0640); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile("go.mod")
			if err != nil {
				t.Fatal(err)
			}
			var out, stderr bytes.Buffer
			err = (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader("App\nuser\nkeep\n" + answer), Out: &out, Err: &stderr})
			if (err != nil) != (answer == "invalid\n") {
				t.Fatalf("unexpected error %v", err)
			}
			after, readErr := os.ReadFile("go.mod")
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatal("cancellation wrote identity")
			}
			data, readErr := os.ReadFile(".env")
			if readErr != nil || string(data) != initSetupKept {
				t.Fatal("cancellation changed environment")
			}
			if strings.Count(out.String(), "Apply this plan?") != 1 || stderr.Len() != 0 {
				t.Fatal("unexpected prompts or stderr")
			}
		})
	}
}

func TestInitSetupFileFailurePreservesDiagnostic(t *testing.T) {
	initFixture(t)
	if err := os.WriteFile(".env", []byte(initSetupKept), 0640); err != nil {
		t.Fatal(err)
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	plan, err := projectinit.New(dir, "App", "example.com/app")
	if err != nil {
		t.Fatal(err)
	}
	selected, err := plan.SelectEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("README.md", []byte("# Competing edit\n"), 0644); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	err = executeInitSetup(context.Background(), cli.Stdio{Out: &out, Err: &stderr}, plan, selected)
	if err == nil || errors.Unwrap(err) == nil || !strings.Contains(err.Error(), "candidate changed since preview: README.md") || out.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("unsafe or missing publication diagnostic: %v", err)
	}
	data, err := os.ReadFile("README.md")
	if err != nil || string(data) != "# Competing edit\n" {
		t.Fatal("publication overwrote competing edit")
	}
	data, err = os.ReadFile("go.mod")
	if err != nil || !strings.Contains(string(data), projectinit.OriginalModule()) {
		t.Fatal("failed publication adopted identity")
	}
}

func TestInitSetupKeptURLPreviewUsesExactTarget(t *testing.T) {
	initFixture(t)
	const kept = "DATABASE_URL=postgres://selected_user:HiddenDBPassword9@kept.invalid:5448/selected_db?sslmode=disable\nPASETO_KEY=abababababababababababababababababababababababababababababababab\n"
	if err := os.WriteFile(".env", []byte(kept), 0640); err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	exit := cli.Execute([]string{"init", "--name", "App", "--module", "example.com/app", "--apply", "--env-action", "keep"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr}, []cli.Command{&InitCommand{}})
	if exit != 0 || stderr.Len() != 0 || !strings.Contains(out.String(), "host=kept.invalid port=5448 database=selected_db user=selected_user sslmode=disable") {
		t.Fatal("selected URL target preview incorrect")
	}
	if strings.Contains(out.String()+stderr.String(), "HiddenDBPassword9") {
		t.Fatal("URL credential leaked")
	}
}

func TestInitSetupExecuteCanceledAndSnapshotChanged(t *testing.T) {
	for _, scenario := range []string{"canceled", "changed"} {
		t.Run(scenario, func(t *testing.T) {
			initFixture(t)
			if err := os.WriteFile(".env", []byte(initSetupKept), 0640); err != nil {
				t.Fatal(err)
			}
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			plan, err := projectinit.New(dir, "App", "example.com/app")
			if err != nil {
				t.Fatal(err)
			}
			selected, err := plan.SelectEnvironment()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if scenario == "canceled" {
				cancel()
			} else if err := os.WriteFile(".env", []byte(initSetupKept+"# changed\n"), 0640); err != nil {
				t.Fatal(err)
			}
			before, err := os.ReadFile("go.mod")
			if err != nil {
				t.Fatal(err)
			}
			var out, stderr bytes.Buffer
			err = executeInitSetup(ctx, cli.Stdio{Out: &out, Err: &stderr}, plan, selected)
			if err == nil || out.Len() != 0 || stderr.Len() != 0 {
				t.Fatal("invalid execution published or printed success")
			}
			after, readErr := os.ReadFile("go.mod")
			if readErr != nil || !bytes.Equal(before, after) {
				t.Fatal("execution changed identity")
			}
			if scenario == "canceled" && !strings.Contains(err.Error(), "canceled") {
				t.Fatal("inaccurate cancellation")
			}
		})
	}
}
