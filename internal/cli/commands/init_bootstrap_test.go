package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/enviniom/nexokit/internal/cli/projectinit"
	"github.com/enviniom/nexokit/internal/cli/root"
	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	pwd "github.com/enviniom/nexokit/internal/platform/password"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
)

func bootstrapFixture(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	for _, d := range []string{"seeds", "migrations"} {
		if err := os.Mkdir(d, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for path, data := range map[string]string{
		"go.mod":                     "module example.test/adopted\ngo 1.26.3\n",
		"seeds/roles.go":             "package seeds\nfunc RolesSeed() error {return nil}\n",
		"migrations/001_initial.sql": "-- +goose Up\nCREATE TABLE example(id integer);\n-- +goose Down\nDROP TABLE example;\n",
	} {
		if err := os.WriteFile(filepath.Clean(path), []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
func selectedBootstrapValues() map[string]string {
	return map[string]string{"DATABASE_URL": "", "DB_HOST": "selected.invalid", "DB_PORT": "5439", "DB_NAME": "owned", "DB_USER": "fake", "DB_PASSWORD": "invented #$'\\\\ password", "DB_SSL_MODE": "disable"}
}
func TestBootstrapPreparationRejectsBeforeEffects(t *testing.T) {
	for _, scenario := range []string{"invalid-name", "invalid-email", "invalid-password", "invalid-db", "missing-seeds", "missing-migrations", "missing-module", "missing-toolchain", "inherited-service", "preflight"} {
		t.Run(scenario, func(t *testing.T) {
			bootstrapFixture(t)
			values := selectedBootstrapValues()
			input := root.CreateRootInput{Name: "Root", Email: "root@example.test", Password: "Invented123"}
			switch scenario {
			case "invalid-name":
				input.Name = ""
			case "invalid-email":
				input.Email = "injected-secret"
			case "invalid-password":
				input.Password = "injected-secret"
			case "invalid-db":
				values["DB_PORT"] = "injected-secret"
			case "missing-seeds":
				if err := os.WriteFile("seeds/roles.go", []byte("package seeds"), 0600); err != nil {
					t.Fatal(err)
				}
			case "missing-migrations":
				if err := os.Rename("migrations", "not-migrations"); err != nil {
					t.Fatal(err)
				}
			case "inherited-service":
				t.Setenv("PGSERVICE", "invented-service")
			case "missing-toolchain":
				t.Setenv("PATH", "")
			case "missing-module":
				if err := os.Rename("go.mod", "not-go.mod"); err != nil {
					t.Fatal(err)
				}
			}
			calls := 0
			p, err := prepareBootstrapWithConnection(context.Background(), values, input, func(_ context.Context, cfg *config.Config) (*gorm.DB, error) {
				calls++
				if cfg.DB.DatabaseURL != "" || cfg.DB.Host != "selected.invalid" || cfg.DB.Password != values["DB_PASSWORD"] {
					t.Fatal("wrong target")
				}
				return nil, errors.New("injected-secret")
			})
			want := 0
			if scenario == "preflight" {
				want = 1
			}
			if p != nil || err == nil || calls != want {
				t.Fatalf("prepared %v error %v connection calls %d", p, err, calls)
			}
			if strings.Contains(err.Error(), "injected-secret") {
				t.Fatal("error leaked secret")
			}
		})
	}
}
func TestBootstrapConfigurationPinsEveryDBField(t *testing.T) {
	t.Setenv("PGPASSWORD", "wrong-inherited-password")
	t.Setenv("PGSSLMODE", "verify-full")
	values := selectedBootstrapValues()
	t.Setenv("DATABASE_URL", "postgres://wrong:invented@wrong.invalid/wrong?sslmode=disable")
	cfg, env, target, err := bootstrapConfig(values)
	if err != nil {
		t.Fatal(err)
	}
	if env["PGPASSWORD"] != values["DB_PASSWORD"] || env["PGSSLMODE"] != "disable" || env["PGPASSFILE"] != os.DevNull {
		t.Fatal("libpq credential defaults not pinned")
	}
	if cfg.DB.DatabaseURL != "" || target.Host != "selected.invalid" {
		t.Fatal("inherited URL selected")
	}
	for _, key := range []string{"DATABASE_URL", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE"} {
		value, ok := env[key]
		if !ok || value != values[key] {
			t.Fatalf("unpinned %s", key)
		}
	}
	values["DATABASE_URL"] = "postgres://chosen:invented@127.0.0.1:5441/url_owned?sslmode=disable"
	cfg, env, target, err = bootstrapConfig(values)
	if err != nil || cfg.DB.DatabaseURL != values["DATABASE_URL"] || env["DATABASE_URL"] != values["DATABASE_URL"] || target.Host != "127.0.0.1" || target.Port != 5441 || target.Name != "url_owned" || target.User != "chosen" || target.SSLMode != "disable" {
		t.Fatal("URL target not preserved")
	}
}

func TestBootstrapURLWithoutPasswordOrSSLDoesNotInheritDefaults(t *testing.T) {
	for _, key := range []string{"PGSSLCERT", "PGSSLKEY", "PGSSLROOTCERT"} {
		t.Setenv(key, "")
	}
	t.Setenv("PGPORT", "5544")
	t.Setenv("PGPASSWORD", "wrong-inherited-password")
	t.Setenv("PGSSLMODE", "verify-full")
	values := selectedBootstrapValues()
	values["DATABASE_URL"] = "postgres://chosen@127.0.0.1/owned"
	_, env, target, err := bootstrapConfig(values)
	if err != nil || env["DATABASE_URL"] != values["DATABASE_URL"] || env["PGPASSWORD"] != "" || env["PGPASSFILE"] != os.DevNull || env["PGSSLMODE"] != "prefer" || target.SSLMode != "prefer" || target.Port != 5432 || env["PGPORT"] != "5432" {
		t.Fatal("URL inherited defaults or changed")
	}
}

func TestBootstrapPreparationIsReadOnlyAndCloseDisablesRun(t *testing.T) {
	bootstrapFixture(t)
	t.Setenv("DATABASE_URL", "postgres://wrong.invalid/wrong?sslmode=disable")
	values := selectedBootstrapValues()
	input := root.CreateRootInput{Name: "Root", Email: "root@example.test", Password: "Invented123"}
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	pool, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	connects := 0
	p, err := prepareBootstrapWithConnection(context.Background(), values, input, func(_ context.Context, cfg *config.Config) (*gorm.DB, error) {
		connects++
		if cfg.DB.DatabaseURL != "" || cfg.DB.Host != values["DB_HOST"] || cfg.DB.Port != 5439 || cfg.DB.Name != values["DB_NAME"] || cfg.DB.User != values["DB_USER"] || cfg.DB.Password != values["DB_PASSWORD"] || cfg.DB.SSLMode != values["DB_SSL_MODE"] {
			t.Fatal("preflight was redirected")
		}
		return database, nil
	})
	if err != nil || connects != 1 {
		t.Fatalf("prepare: %v calls %d", err, connects)
	}
	// Public storage boundary still cannot find its table: preparation has not
	// run migrations, seeds or root creation, even on a successful connection.
	if _, err := newRootStorage(database).RootExists(); err == nil {
		t.Fatal("preparation created tables")
	}
	if p.Target.Host != "selected.invalid" || p.Target.Name != "owned" || p.Target.User != "fake" || p.Target.Port != 5439 || p.Target.SSLMode != "disable" {
		t.Fatal("preview target")
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if err := pool.Ping(); err == nil {
		t.Fatal("connection still open")
	}
	var out, stderr bytes.Buffer
	if err := p.Run(context.Background(), cli.Stdio{Out: &out, Err: &stderr}); err == nil || out.Len() != 0 || stderr.Len() != 0 {
		t.Fatal("closed setup ran")
	}
	if _, err := os.Stat(".tmp"); !os.IsNotExist(err) {
		t.Fatal("preparation created runner files")
	}
}

func TestPreparedBootstrapCanceledBeforeStage(t *testing.T) {
	calls := 0
	p := &preparedInitBootstrap{stages: []bootstrapStage{{"migrations", func(context.Context) error { calls++; return nil }}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out, stderr bytes.Buffer
	err := p.Run(ctx, cli.Stdio{Out: &out, Err: &stderr})
	if err == nil || err.Error() != "initial setup migrations failed" || calls != 0 || out.Len() != 0 || stderr.Len() != 0 {
		t.Fatal("canceled setup had effects")
	}
}

func TestPreparedBootstrapLifecycle(t *testing.T) {
	calls, closes := 0, 0
	p := &preparedInitBootstrap{stages: []bootstrapStage{{"root", func(context.Context) error { calls++; return root.ErrRootAlreadyExists }}}, close: func() error { closes++; return errors.New("injected-secret") }}
	var out bytes.Buffer
	stdio := cli.Stdio{Out: &out}
	if err := p.Run(context.Background(), stdio); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "already exists") || strings.Contains(out.String(), "created") {
		t.Fatal("misleading output")
	}
	if err := p.Run(context.Background(), stdio); err == nil {
		t.Fatal("rerun succeeded")
	}
	if err := p.Close(); err == nil || strings.Contains(err.Error(), "injected-secret") {
		t.Fatal("unsafe close")
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	if err := p.Run(context.Background(), stdio); err == nil || calls != 1 || closes != 1 {
		t.Fatal("lifecycle effects")
	}
}

// Parent-owned long verification only. This test never selects a default DB,
// starts a container or creates/drops a database. The parent supplies an empty
// owned database inside the authorized disposable container.
func TestBootstrapIsolatedPostgres(t *testing.T) {
	if testing.Short() {
		t.Skip("isolated PostgreSQL verification is parent-owned")
	}
	raw := os.Getenv("NEXOKIT_BOOTSTRAP_TEST_URL")
	if raw == "" {
		t.Skip("NEXOKIT_BOOTSTRAP_TEST_URL not supplied; no PostgreSQL proof")
	}
	uri, err := url.Parse(raw)
	if err != nil || uri.Hostname() != "127.0.0.1" || uri.Port() == "" || uri.Port() == "5432" || !strings.HasPrefix(strings.TrimPrefix(uri.Path, "/"), "nexokit_init_verify_") || uri.Query().Get("sslmode") != "disable" || uri.User == nil || uri.User.Username() != "nexokit_verify" {
		t.Fatal("refusing non-isolated test database")
	}
	source, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// go test starts in internal/cli/commands; copy from the repository root.
	source = filepath.Clean(filepath.Join(source, "..", "..", ".."))
	if !strings.HasPrefix(source, "/tmp/nexokit-initial-setup-verify-01a131b8/") {
		t.Fatal("integration requires an authorized synthetic source copy")
	}
	fixture := t.TempDir()
	for _, name := range []string{"cmd", "internal", "seeds", "migrations", "docs", "scripts", "tests", "go.mod", "go.sum", "README.md", "Makefile"} {
		start := filepath.Join(source, name)
		if _, err := os.Stat(start); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(start, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor" || entry.Name() == "logs" || entry.Name() == "tmp" || entry.Name() == "artifacts" || entry.Name() == "generated" {
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
			dest := filepath.Join(fixture, rel)
			if entry.IsDir() {
				return os.MkdirAll(dest, 0700)
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
			t.Fatal("synthetic source copy failed")
		}
	}
	t.Chdir(fixture)
	// This is invented test data, never the protected repository template.
	template := "APP_NAME=nexokit\nAPP_ENV=development\nDATABASE_URL=\nDB_HOST=127.0.0.1\nDB_PORT=5439\nDB_NAME=synthetic\nDB_USER=fake\nDB_PASSWORD=invented\nDB_SSL_MODE=disable\n"
	if err := os.WriteFile(".env.example", []byte(template), 0600); err != nil {
		t.Fatal(err)
	}
	values := selectedBootstrapValues()
	values["DATABASE_URL"] = ""
	values["DB_HOST"] = uri.Hostname()
	values["DB_PORT"] = uri.Port()
	values["DB_NAME"] = strings.TrimPrefix(uri.Path, "/")
	values["DB_USER"] = uri.User.Username()
	values["DB_PASSWORD"], _ = uri.User.Password()
	values["DB_SSL_MODE"] = "disable"
	values["APP_ENV"] = "development"
	if os.Getenv("NEXOKIT_BOOTSTRAP_TEST_TARGET") == "url" {
		values["DATABASE_URL"] = raw
	}
	if port, err := strconv.Atoi(values["DB_PORT"]); err != nil || port < 1 || port > 65535 {
		t.Fatal("invalid isolated port")
	}
	input := root.CreateRootInput{Name: "Synthetic Root", Email: "root@example.test", Password: "InventedRoot123"}
	t.Setenv("DATABASE_URL", "postgres://wrong:invented@127.0.0.1:1/wrong?sslmode=disable")
	t.Setenv("ROOT_USER_PASSWORD", "must-not-reach-child")
	prepared, err := prepareInitBootstrap(context.Background(), values, input)
	if err != nil {
		t.Fatal(err)
	}
	defer prepared.Close()
	plan, err := projectinit.New(fixture, "Adopted Fixture", "example.test/adopted")
	if err != nil {
		t.Fatal("identity plan failed")
	}
	if err := plan.ConfigureEnvironment(values); err != nil {
		t.Fatal("synthetic environment plan failed")
	}
	if err := plan.Apply(); err != nil {
		t.Fatal("identity application failed")
	}
	var out, stderr bytes.Buffer
	if err := prepared.Run(context.Background(), cli.Stdio{Out: &out, Err: &stderr}); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Running migrations...\nRunning seeds...\nRunning root...\n" || stderr.Len() != 0 {
		t.Fatal("unexpected setup output")
	}
	cfg, _, _, err := bootstrapConfig(values)
	if err != nil {
		t.Fatal(err)
	}
	database, err := db.ConnectQuiet(context.Background(), cfg)
	if err != nil {
		t.Fatal("verification connection failed")
	}
	defer db.Close(database)
	// Verify existence and refusal to replace through the domain API, not raw
	// user-table assertions. Legacy Creator tests protect the no-hash skip rule.
	changed := input
	changed.Password = "DifferentRoot123"
	if err := root.NewCreator(newRootStorage(database), pwd.Manager{}).Create(changed); !errors.Is(err, root.ErrRootAlreadyExists) {
		t.Fatal("root missing or replacement permitted")
	}
	if err := prepared.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestPreparedBootstrapOrderAndFailure(t *testing.T) {
	for _, fail := range []string{"", "migrations", "seeds", "root"} {
		t.Run("failure_"+fail, func(t *testing.T) {
			var calls []string
			p := &preparedInitBootstrap{}
			for _, name := range []string{"migrations", "seeds", "root"} {
				name := name
				p.stages = append(p.stages, bootstrapStage{name, func(context.Context) error {
					calls = append(calls, name)
					if name == fail {
						return errors.New("injected-secret")
					}
					return nil
				}})
			}
			if len(calls) != 0 {
				t.Fatal("preparation ran stages")
			}
			var out, stderr bytes.Buffer
			err := p.Run(context.Background(), cli.Stdio{Out: &out, Err: &stderr})
			want := []string{"migrations", "seeds", "root"}
			if fail == "migrations" {
				want = want[:1]
			}
			if fail == "seeds" {
				want = want[:2]
			}
			if !reflect.DeepEqual(calls, want) {
				t.Fatalf("calls %v want %v", calls, want)
			}
			if (err != nil) != (fail != "") {
				t.Fatalf("error %v", err)
			}
			if err != nil && err.Error() != "initial setup "+fail+" failed" {
				t.Fatalf("unsafe error %v", err)
			}
			var progress strings.Builder
			for _, name := range want {
				fmt.Fprintf(&progress, "Running %s...\n", name)
			}
			if out.String() != progress.String() {
				t.Fatalf("stdout %q want %q", out.String(), progress.String())
			}
			if strings.Contains(out.String()+stderr.String(), "injected-secret") || stderr.Len() != 0 {
				t.Fatal("secret/error output")
			}
		})
	}
}
