package commands

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
	"github.com/enviniom/nexokit/internal/cli/root"
	"github.com/enviniom/nexokit/internal/infra/db"
	pwd "github.com/enviniom/nexokit/internal/platform/password"
	"github.com/joho/godotenv"
)

// These tests exercise the public command without a database or bootstrap bypass.
func TestInitSetupApplyRequiresInputsBeforeWrites(t *testing.T) {
	initFixture(t)
	for _, key := range []string{"ROOT_USER_NAME", "ROOT_USER_EMAIL", "ROOT_USER_PASSWORD", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSL_MODE", "DATABASE_URL", "PASETO_KEY"} {
		t.Setenv(key, "")
	}
	before, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	err = (&InitCommand{}).Run(context.Background(), []string{"--name", "App", "--module", "example.com/app", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr})
	if err == nil {
		t.Error("full setup accepted missing root/database inputs")
	}
	after, readErr := os.ReadFile("go.mod")
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(before, after) {
		t.Error("missing setup inputs changed identity")
	}
	if strings.Contains(out.String(), "Applied.") {
		t.Error("missing setup inputs advertised success")
	}
}

func TestInitSetupExistingEnvironmentRequiresExplicitFlagChoice(t *testing.T) {
	initFixture(t)
	const invented = "INVENTED_SETTING=preserved\n"
	if err := os.WriteFile(".env", []byte(invented), 0640); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	var out, stderr bytes.Buffer
	err = (&InitCommand{}).Run(context.Background(), []string{"--name", "App", "--module", "example.com/app", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr})
	if err == nil {
		t.Error("existing environment accepted without explicit keep/rebuild selection")
	}
	after, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Error("missing environment choice changed identity")
	}
	data, err := os.ReadFile(".env")
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(".env")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != invented || info.Mode().Perm() != 0640 {
		t.Error("rejection changed kept environment bytes/mode")
	}
}

type initSetupCountingInput struct{ reads int }

func (r *initSetupCountingInput) Read([]byte) (int, error) { r.reads++; return 0, io.EOF }

func TestInitSetupAdoptedRejectsBeforeReadingInput(t *testing.T) {
	initFixture(t)
	const adopted = "module example.com/adopted\n"
	if err := os.WriteFile("go.mod", []byte(adopted), 0644); err != nil {
		t.Fatal(err)
	}
	input := &initSetupCountingInput{}
	var out, stderr bytes.Buffer
	err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: input, Out: &out, Err: &stderr})
	if err == nil {
		t.Error("adopted module accepted")
	}
	if input.reads != 0 {
		t.Error("adopted module prompted/read input before adoption rejection")
	}
	if strings.Contains(out.String(), "Application display name:") {
		t.Error("adopted module prompted for identity")
	}
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != adopted || strings.Contains(string(data), projectinit.OriginalModule()) {
		t.Error("adopted rejection changed module")
	}
}

func TestInitSetupKeptEnvironmentLifecycle(t *testing.T) {
	for _, failure := range []string{"", "prepare", "migrations", "seeds", "root"} {
		t.Run("failure_"+failure, func(t *testing.T) {
			initFixture(t)
			const secret = " NotPrintedRoot9 "
			kept := "APP_ENV=development\nDATABASE_URL=\nDB_HOST=kept.invalid\nDB_PORT=5439\nDB_NAME=kept_db\nDB_USER=kept_user\nDB_PASSWORD=synthetic-kept\nDB_SSL_MODE=disable\nPASETO_KEY=" + strings.Repeat("ab", 32) + "\nUNKNOWN_SETTING=retained\n"
			if err := os.WriteFile(".env", []byte(kept), 0640); err != nil {
				t.Fatal(err)
			}
			t.Setenv("ROOT_USER_NAME", "Root")
			t.Setenv("ROOT_USER_EMAIL", "root@example.test")
			t.Setenv("ROOT_USER_PASSWORD", secret)
			t.Setenv("DB_HOST", "ambient.invalid")
			var calls []string
			old := initSetupPrepare
			t.Cleanup(func() { initSetupPrepare = old })
			initSetupPrepare = func(ctx context.Context, values map[string]string, input root.CreateRootInput) (*preparedInitBootstrap, error) {
				calls = append(calls, "prepare")
				if values["DB_HOST"] != "kept.invalid" || values["UNKNOWN_SETTING"] != "retained" {
					t.Error("effective kept configuration not used")
				}
				if input.Password != secret {
					t.Error("root password not preserved exactly")
				}
				data, err := os.ReadFile("go.mod")
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), projectinit.OriginalModule()) {
					t.Error("identity applied before preparation")
				}
				if failure == "prepare" {
					return nil, errors.New("synthetic preparation failure")
				}
				p := &preparedInitBootstrap{close: func() error { calls = append(calls, "close"); return nil }}
				for _, name := range []string{"migrations", "seeds", "root"} {
					stage := name
					p.stages = append(p.stages, bootstrapStage{name: stage, run: func(context.Context) error {
						calls = append(calls, stage)
						data, err := os.ReadFile("go.mod")
						if err != nil {
							t.Fatal(err)
						}
						if !strings.Contains(string(data), "example.com/app") {
							t.Error("database stage ran before identity apply")
						}
						if failure == stage {
							return errors.New("synthetic secret-bearing driver error: " + secret)
						}
						return nil
					}})
				}
				return p, nil
			}
			var out, stderr bytes.Buffer
			err := (&InitCommand{}).Run(context.Background(), []string{"--name", "App", "--module", "example.com/app", "--env-action", "keep", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr})
			want := []string{"prepare"}
			if failure != "prepare" {
				for _, name := range []string{"migrations", "seeds", "root"} {
					want = append(want, name)
					if failure == name {
						break
					}
				}
				want = append(want, "close")
			}
			if !reflect.DeepEqual(calls, want) {
				t.Errorf("calls %v; want %v", calls, want)
			}
			if (err != nil) != (failure != "") {
				t.Errorf("unexpected failure state; output %s", out.String())
			}
			if strings.Contains(out.String()+stderr.String(), secret) || (err != nil && strings.Contains(err.Error(), secret)) {
				t.Error("root secret leaked")
			}
			if failure != "" && strings.Contains(out.String(), "Setup complete") {
				t.Error("failure advertised completion")
			}
			if failure != "" && failure != "prepare" && !strings.Contains(out.String(), "cannot rerun") {
				t.Error("adopted failure omitted recovery warning")
			}
			data, readErr := os.ReadFile(".env")
			if readErr != nil {
				t.Fatal(readErr)
			}
			info, statErr := os.Stat(".env")
			if statErr != nil {
				t.Fatal(statErr)
			}
			if string(data) != kept || info.Mode().Perm() != 0640 {
				t.Error("keep changed environment")
			}
		})
	}
}

func initSetupProcessSettings(t *testing.T) {
	t.Helper()
	for key, value := range map[string]string{
		"DB_HOST": "selected.invalid", "DB_PORT": "5440", "DB_NAME": "selected_db", "DB_USER": "selected_user", "DB_PASSWORD": "SyntheticDBOnly9", "APP_PORT": "8081", "APP_URL": "http://localhost:8081", "CORS_ALLOWED_ORIGINS": "http://localhost:3001", "ROOT_USER_NAME": "Root", "ROOT_USER_EMAIL": "root@example.test", "ROOT_USER_PASSWORD": " NotPrintedRoot9 ",
	} {
		t.Setenv(key, value)
	}
	t.Setenv("DATABASE_URL", "postgres://ambient:WrongPassword9@ambient.invalid/other")
}

func TestInitSetupGeneratedEnvironment(t *testing.T) {
	for _, rebuild := range []bool{false, true} {
		t.Run(map[bool]string{false: "new", true: "rebuild"}[rebuild], func(t *testing.T) {
			initFixture(t)
			template := "APP_NAME=nexokit\n# retained comment\nUNKNOWN_SETTING=template-retained\nDB_SSL_MODE=disable\n"
			if err := os.WriteFile(".env.example", []byte(template), 0644); err != nil {
				t.Fatal(err)
			}
			const oldEnv = "# old invented configuration\nUNKNOWN_SETTING=old\n"
			if rebuild {
				if err := os.WriteFile(".env", []byte(oldEnv), 0640); err != nil {
					t.Fatal(err)
				}
			}
			initSetupProcessSettings(t)
			t.Setenv("DB_SSL_MODE", "disable")
			var selected map[string]string
			calls := 0
			previous := initSetupPrepare
			t.Cleanup(func() { initSetupPrepare = previous })
			initSetupPrepare = func(_ context.Context, values map[string]string, input root.CreateRootInput) (*preparedInitBootstrap, error) {
				selected = values
				calls++
				if input.Password != " NotPrintedRoot9 " {
					t.Error("root secret changed")
				}
				if values["DATABASE_URL"] != "" || values["APP_ENV"] != "development" || values["DB_HOST"] != "selected.invalid" || values["DB_SSL_MODE"] != "disable" || values["UNKNOWN_SETTING"] != "template-retained" {
					t.Error("generated effective configuration mismatch")
				}
				if _, exists := values["ROOT_USER_PASSWORD"]; exists {
					t.Error("root secret included in effective environment")
				}
				return &preparedInitBootstrap{stages: []bootstrapStage{{name: "root", run: func(context.Context) error { return nil }}}}, nil
			}
			args := []string{"--name", "App", "--module", "example.com/app", "--apply"}
			if rebuild {
				args = append(args, "--env-action", "rebuild")
			}
			var out, stderr bytes.Buffer
			err := (&InitCommand{}).Run(context.Background(), args, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr})
			if err != nil {
				t.Fatal(err)
			}
			if calls != 1 {
				t.Errorf("prepare calls %d", calls)
			}
			data, err := os.ReadFile(".env")
			if err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(".env")
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0600 || !strings.Contains(string(data), "# retained comment") {
				t.Error("new environment permission/template mismatch")
			}
			if strings.Contains(string(data)+out.String()+stderr.String(), "NotPrintedRoot9") {
				t.Error("root secret leaked")
			}
			if len(selected["PASETO_KEY"]) != 64 {
				t.Error("fresh key missing")
			}
			backups, err := os.ReadDir(".")
			if err != nil {
				t.Fatal(err)
			}
			count := 0
			for _, entry := range backups {
				if !strings.HasPrefix(entry.Name(), ".env.backup-") {
					continue
				}
				count++
				bytes, err := os.ReadFile(entry.Name())
				if err != nil {
					t.Fatal(err)
				}
				info, err := entry.Info()
				if err != nil {
					t.Fatal(err)
				}
				if string(bytes) != oldEnv || info.Mode().Perm() != 0600 || !strings.Contains(out.String(), entry.Name()) {
					t.Error("backup bytes/mode/preview mismatch")
				}
			}
			if (count == 1) != rebuild {
				t.Error("wrong backup artifact count")
			}
		})
	}
}

func TestInitSetupInteractiveCancellationAndChoice(t *testing.T) {
	for _, answer := range []string{"n\n", "\n", "", "y"} {
		t.Run("answer_"+answer, func(t *testing.T) {
			initFixture(t)
			kept := "DB_HOST=kept.invalid\nDB_PORT=5439\nDB_NAME=kept_db\nDB_USER=kept_user\nDB_PASSWORD=synthetic-only\nDB_SSL_MODE=disable\nPASETO_KEY=" + strings.Repeat("ab", 32) + "\n"
			if err := os.WriteFile(".env", []byte(kept), 0640); err != nil {
				t.Fatal(err)
			}
			calls := 0
			old := initSetupPrepare
			t.Cleanup(func() { initSetupPrepare = old })
			initSetupPrepare = func(context.Context, map[string]string, root.CreateRootInput) (*preparedInitBootstrap, error) {
				calls++
				return nil, errors.New("unexpected preparation")
			}
			var out, stderr bytes.Buffer
			input := "My API\nuser-name\n\nRoot\nroot@example.test\n NotPrintedRoot9 \n" + answer
			err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: strings.NewReader(input), Out: &out, Err: &stderr})
			if err != nil {
				t.Fatal(err)
			}
			if calls != 0 {
				t.Error("cancellation prepared bootstrap")
			}
			if strings.Count(out.String(), "Apply this plan?") != 1 || !strings.Contains(out.String(), "KEEP or BACKUP+REBUILD") {
				t.Error("missing choice or repeated final confirmation")
			}
			if strings.Contains(out.String()+stderr.String(), "NotPrintedRoot9") {
				t.Error("root secret echoed")
			}
			data, err := os.ReadFile("go.mod")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), projectinit.OriginalModule()) {
				t.Error("cancellation adopted identity")
			}
			data, err = os.ReadFile(".env")
			if err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(".env")
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != kept || info.Mode().Perm() != 0640 {
				t.Error("cancellation changed kept environment")
			}
		})
	}
}

func TestInitSetupInvalidKeptKeyRejectsBeforePreparation(t *testing.T) {
	initFixture(t)
	if err := os.WriteFile(".env", []byte("PASETO_KEY=invalid\n"), 0600); err != nil {
		t.Fatal(err)
	}
	initSetupProcessSettings(t)
	calls := 0
	old := initSetupPrepare
	t.Cleanup(func() { initSetupPrepare = old })
	initSetupPrepare = func(context.Context, map[string]string, root.CreateRootInput) (*preparedInitBootstrap, error) {
		calls++
		return nil, errors.New("unexpected")
	}
	var out bytes.Buffer
	err := (&InitCommand{}).Run(context.Background(), []string{"--name", "App", "--module", "example.com/app", "--env-action", "keep", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &out})
	if err == nil || calls != 0 {
		t.Error("invalid kept key passed validation")
	}
	data, readErr := os.ReadFile("go.mod")
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(data), projectinit.OriginalModule()) {
		t.Error("invalid key adopted identity")
	}
}

// TestInitPublicIsolatedPostgres is deliberately excluded from the writer's
// focused selection. The parent supplies a fresh authorized disposable database
// for each entrypoint/target case; this harness never provisions a service.
func TestInitPublicIsolatedPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("real PostgreSQL/public process verification is parent-owned")
	}
	raw := os.Getenv("NEXOKIT_BOOTSTRAP_TEST_URL")
	if raw == "" {
		t.Skip("NEXOKIT_BOOTSTRAP_TEST_URL absent; no public PostgreSQL/Make proof")
	}
	uri, err := url.Parse(raw)
	if err != nil || uri.Hostname() != "127.0.0.1" || uri.Port() == "" || uri.Port() == "5432" || !strings.HasPrefix(strings.TrimPrefix(uri.Path, "/"), "nexokit_init_verify_") || uri.Query().Get("sslmode") != "disable" || uri.User == nil || uri.User.Username() != "nexokit_verify" {
		t.Fatal("refusing non-isolated database")
	}
	source, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal("source root unavailable")
	}
	if !strings.HasPrefix(source, "/tmp/nexokit-initial-setup-verify-01a131b8/") {
		t.Fatal("public probe requires authorized synthetic source")
	}
	target := t.TempDir()
	for _, name := range []string{"cmd", "internal", "tests", "seeds", "docs", "migrations", "scripts", "go.mod", "go.sum", "README.md", "Makefile"} {
		start := filepath.Join(source, name)
		err := filepath.WalkDir(start, func(path string, entry fs.DirEntry, walkErr error) error {
			if os.IsNotExist(walkErr) && path == start {
				return nil
			}
			if walkErr != nil {
				return walkErr
			}
			if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor" || entry.Name() == "generated" || entry.Name() == "log" || entry.Name() == "logs" || entry.Name() == "tmp" || entry.Name() == "artifacts" {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return nil
			}
			rel, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			dest := filepath.Join(target, rel)
			if entry.IsDir() {
				return os.MkdirAll(dest, 0700)
			}
			if !entry.Type().IsRegular() {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(dest), 0700); err != nil {
				return err
			}
			return os.WriteFile(dest, data, 0600)
		})
		if err != nil {
			t.Fatal("synthetic whitelist copy failed")
		}
	}
	if err := os.WriteFile(filepath.Join(target, ".env.example"), []byte("APP_NAME=nexokit\n# invented public probe\nDB_SSL_MODE=disable\nUNKNOWN_SETTING=retained\n"), 0600); err != nil {
		t.Fatal("template creation failed")
	}
	password, _ := uri.User.Password()
	values := map[string]string{"APP_PORT": "8080", "APP_URL": "http://localhost:8080", "CORS_ALLOWED_ORIGINS": "http://localhost:3000", "DB_HOST": uri.Hostname(), "DB_PORT": uri.Port(), "DB_NAME": strings.TrimPrefix(uri.Path, "/"), "DB_USER": uri.User.Username(), "DB_PASSWORD": password, "DB_SSL_MODE": "disable", "DATABASE_URL": "", "PASETO_KEY": strings.Repeat("ab", 32)}
	for _, value := range values {
		if !safeLocalValue(value) {
			t.Fatal("unsafe synthetic input")
		}
	}
	input := root.CreateRootInput{Name: "Synthetic Root", Email: "root@example.test", Password: " PublicRootSecret9 "}
	kept := os.Getenv("NEXOKIT_BOOTSTRAP_TEST_TARGET") == "url"
	if kept {
		values["DATABASE_URL"] = raw
		var dotenv strings.Builder
		for _, key := range []string{"APP_PORT", "APP_URL", "CORS_ALLOWED_ORIGINS", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE", "DATABASE_URL", "PASETO_KEY"} {
			// godotenv marshaling handles special credentials without an argv secret.
			line, err := godotenv.Marshal(map[string]string{key: values[key]})
			if err != nil {
				t.Fatal("dotenv encoding failed")
			}
			dotenv.WriteString(line + "\n")
		}
		if err := os.WriteFile(filepath.Join(target, ".env"), []byte(dotenv.String()), 0640); err != nil {
			t.Fatal("kept fixture creation failed")
		}
	}
	entrypoint := os.Getenv("NEXOKIT_INIT_TEST_ENTRYPOINT")
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	var command *exec.Cmd
	if entrypoint == "make" {
		command = exec.CommandContext(ctx, "make", "init")
		wizard := "Public Fixture\nfixture-user\n"
		if kept {
			wizard += "keep\n"
		} else {
			for _, key := range []string{"DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE", "APP_PORT", "APP_URL", "CORS_ALLOWED_ORIGINS"} {
				wizard += values[key] + "\n"
			}
		}
		command.Stdin = strings.NewReader(wizard + input.Name + "\n" + input.Email + "\n" + input.Password + "\ny\n")
	} else if entrypoint == "" || entrypoint == "cli" {
		args := []string{"run", "./cmd/nexokit", "init", "--name", "Public Fixture", "--module", "example.test/public-fixture", "--apply"}
		if kept {
			args = append(args, "--env-action", "keep")
		}
		command = exec.CommandContext(ctx, "go", args...)
		command.Stdin = strings.NewReader("")
	} else {
		t.Fatal("unsupported test entrypoint")
	}
	command.Dir = target
	for _, entry := range os.Environ() {
		key := strings.SplitN(entry, "=", 2)[0]
		if strings.HasPrefix(key, "DB_") || strings.HasPrefix(key, "PG") || strings.HasPrefix(key, "ROOT_") || strings.HasPrefix(key, "NEXOKIT_") || key == "DATABASE_URL" || key == "GOFLAGS" {
			continue
		}
		command.Env = append(command.Env, entry)
	}
	for key, value := range values {
		if key != "PASETO_KEY" {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "ROOT_USER_NAME="+input.Name, "ROOT_USER_EMAIL="+input.Email, "ROOT_USER_PASSWORD="+input.Password, "GOWORK=off", "GOPROXY=off", "GOSUMDB=off")
	var out, stderr bytes.Buffer
	command.Stdout = &out
	command.Stderr = &stderr
	runErr := command.Run()
	if strings.Contains(out.String()+stderr.String(), input.Password) || (password != "" && strings.Contains(out.String()+stderr.String(), password)) {
		t.Fatal("public process leaked a secret")
	}
	if runErr != nil {
		t.Fatal("public setup process failed; driver/output details intentionally withheld")
	}
	text := out.String()
	previous := -1
	for _, stage := range []string{"Running migrations...", "Running seeds...", "Running root...", "Setup complete"} {
		index := strings.Index(text, stage)
		if index <= previous {
			t.Fatal("public phases missing or out of order")
		}
		previous = index
	}
	t.Chdir(target)
	effective, err := godotenv.Read(".env")
	if err != nil {
		t.Fatal("effective dotenv unavailable")
	}
	if strings.Contains(effective["ROOT_USER_PASSWORD"], "PublicRootSecret9") {
		t.Fatal("root secret stored in dotenv")
	}
	cfg, _, _, err := bootstrapConfig(effective)
	if err != nil {
		t.Fatal("effective target validation failed")
	}
	database, err := db.ConnectQuiet(context.Background(), cfg)
	if err != nil {
		t.Fatal("root verification connection failed")
	}
	defer db.Close(database)
	if err := root.NewCreator(newRootStorage(database), pwd.Manager{}).Create(input); !errors.Is(err, root.ErrRootAlreadyExists) {
		t.Fatal("public init did not create root")
	}
	var rejected bytes.Buffer
	if err := (&InitCommand{}).Run(context.Background(), nil, cli.Stdio{In: forbiddenInput{}, Out: &rejected, Err: &rejected}); err == nil {
		t.Fatal("public adoption rerunnable")
	}
	t.Log("Public setup process, ordered phases, stored root and one-time guard verified against isolated PostgreSQL")
}

func TestInitSetupKeptEnvironmentChangedDuringPreparation(t *testing.T) {
	initFixture(t)
	kept := "DB_HOST=kept.invalid\nDB_PORT=5439\nDB_NAME=kept_db\nDB_USER=kept_user\nDB_SSL_MODE=disable\nPASETO_KEY=" + strings.Repeat("ab", 32) + "\n"
	if err := os.WriteFile(".env", []byte(kept), 0640); err != nil {
		t.Fatal(err)
	}
	initSetupProcessSettings(t)
	stages, closed := 0, 0
	old := initSetupPrepare
	t.Cleanup(func() { initSetupPrepare = old })
	initSetupPrepare = func(context.Context, map[string]string, root.CreateRootInput) (*preparedInitBootstrap, error) {
		if err := os.WriteFile(".env", []byte(kept+"# concurrent change\n"), 0600); err != nil {
			t.Fatal(err)
		}
		return &preparedInitBootstrap{close: func() error { closed++; return nil }, stages: []bootstrapStage{{name: "migrations", run: func(context.Context) error { stages++; return nil }}}}, nil
	}
	var out bytes.Buffer
	err := (&InitCommand{}).Run(context.Background(), []string{"--name", "App", "--module", "example.com/app", "--env-action", "keep", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &out})
	if err == nil || stages != 0 || closed != 1 {
		t.Error("changed kept snapshot did not stop before writes and close")
	}
	data, readErr := os.ReadFile("go.mod")
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !strings.Contains(string(data), projectinit.OriginalModule()) {
		t.Error("changed kept snapshot adopted identity")
	}
}

func TestInitSetupExecuteExitAndRedactedPreparationFailure(t *testing.T) {
	initFixture(t)
	initSetupProcessSettings(t)
	t.Setenv("DB_SSL_MODE", "disable")
	old := initSetupPrepare
	t.Cleanup(func() { initSetupPrepare = old })
	calls := 0
	initSetupPrepare = func(context.Context, map[string]string, root.CreateRootInput) (*preparedInitBootstrap, error) {
		calls++
		return nil, errors.New("driver details PublicRootSecret9")
	}
	var out, stderr bytes.Buffer
	exit := cli.Execute([]string{"init", "--name", "App", "--module", "example.com/app", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr}, []cli.Command{&InitCommand{}})
	if exit != 1 || calls != 1 || !strings.Contains(stderr.String(), "initial setup preparation failed") {
		t.Error("public exit/stderr contract incorrect")
	}
	if strings.Contains(out.String()+stderr.String(), "PublicRootSecret9") {
		t.Error("preparation details leaked")
	}
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), projectinit.OriginalModule()) {
		t.Error("failed preparation adopted identity")
	}
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		t.Error("failed preparation published environment")
	}
}

func TestInitSetupFileFailureStopsStagesAndCloses(t *testing.T) {
	initFixture(t)
	initSetupProcessSettings(t)
	t.Setenv("DB_SSL_MODE", "disable")
	old := initSetupPrepare
	t.Cleanup(func() { initSetupPrepare = old })
	calls := []string{}
	initSetupPrepare = func(context.Context, map[string]string, root.CreateRootInput) (*preparedInitBootstrap, error) {
		calls = append(calls, "prepare")
		if err := os.WriteFile("README.md", []byte("# Competing edit\n"), 0644); err != nil {
			t.Fatal(err)
		}
		return &preparedInitBootstrap{close: func() error { calls = append(calls, "close"); return nil }, stages: []bootstrapStage{{name: "migrations", run: func(context.Context) error { calls = append(calls, "migrations"); return nil }}}}, nil
	}
	var out, stderr bytes.Buffer
	exit := cli.Execute([]string{"init", "--name", "App", "--module", "example.com/app", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr}, []cli.Command{&InitCommand{}})
	if exit != 1 || !reflect.DeepEqual(calls, []string{"prepare", "close"}) {
		t.Error("file failure ran database stage or failed to close")
	}
	if !strings.Contains(stderr.String(), "candidate changed since preview: README.md") {
		t.Error("file failure hid the safe publication diagnostic needed for recovery")
	}
	if strings.Contains(out.String()+stderr.String(), "NotPrintedRoot9") {
		t.Error("file failure leaked root credentials")
	}
	data, err := os.ReadFile("go.mod")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), projectinit.OriginalModule()) {
		t.Error("file failure retained adopted identity")
	}
	data, err = os.ReadFile("README.md")
	if err != nil || string(data) != "# Competing edit\n" {
		t.Error("file failure overwrote competitor")
	}
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		t.Error("file failure published environment")
	}
}

func TestInitSetupKeptURLPreviewUsesExactTarget(t *testing.T) {
	initFixture(t)
	const selectedURL = "postgres://selected_user:HiddenDBPassword9@kept.invalid:5448/selected_db?sslmode=disable"
	kept := "DATABASE_URL=" + selectedURL + "\nPASETO_KEY=" + strings.Repeat("ab", 32) + "\n"
	if err := os.WriteFile(".env", []byte(kept), 0600); err != nil {
		t.Fatal(err)
	}
	initSetupProcessSettings(t)
	old := initSetupPrepare
	t.Cleanup(func() { initSetupPrepare = old })
	initSetupPrepare = func(_ context.Context, values map[string]string, _ root.CreateRootInput) (*preparedInitBootstrap, error) {
		if values["DATABASE_URL"] != selectedURL {
			t.Error("kept URL replaced by ambient settings")
		}
		return &preparedInitBootstrap{}, nil
	}
	var out, stderr bytes.Buffer
	exit := cli.Execute([]string{"init", "--name", "App", "--module", "example.com/app", "--apply", "--env-action", "keep"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr}, []cli.Command{&InitCommand{}})
	if exit != 0 || stderr.Len() != 0 {
		t.Error("valid kept URL setup failed")
	}
	if !strings.Contains(out.String(), "host=kept.invalid port=5448 database=selected_db user=selected_user sslmode=disable") {
		t.Error("preview target not actual kept URL target")
	}
	if strings.Contains(out.String()+stderr.String(), "HiddenDBPassword9") {
		t.Error("URL credentials leaked")
	}
}

func TestInitSetupInvalidRootRejectsBeforePreparation(t *testing.T) {
	for key, value := range map[string]string{"ROOT_USER_NAME": "", "ROOT_USER_EMAIL": "invalid", "ROOT_USER_PASSWORD": "weak"} {
		t.Run(key, func(t *testing.T) {
			initFixture(t)
			initSetupProcessSettings(t)
			t.Setenv("DB_SSL_MODE", "disable")
			t.Setenv(key, value)
			calls := 0
			old := initSetupPrepare
			t.Cleanup(func() { initSetupPrepare = old })
			initSetupPrepare = func(context.Context, map[string]string, root.CreateRootInput) (*preparedInitBootstrap, error) {
				calls++
				return nil, errors.New("unexpected")
			}
			var out, stderr bytes.Buffer
			exit := cli.Execute([]string{"init", "--name", "App", "--module", "example.com/app", "--apply"}, cli.Stdio{In: forbiddenInput{}, Out: &out, Err: &stderr}, []cli.Command{&InitCommand{}})
			if exit != 1 || calls != 0 {
				t.Error("invalid root input passed pre-write validation")
			}
			data, err := os.ReadFile("go.mod")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), projectinit.OriginalModule()) {
				t.Error("invalid root input adopted identity")
			}
			if _, err := os.Stat(".env"); !os.IsNotExist(err) {
				t.Error("invalid root input wrote environment")
			}
		})
	}
}
