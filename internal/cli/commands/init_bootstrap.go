package commands

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/cli/root"
	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	pwd "github.com/enviniom/nexokit/internal/platform/password"
	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"
	"gorm.io/gorm"
)

// Preparation is read-only. The caller confirms before preparing, applies its
// file plan after preparation, and only then invokes Run.
type preparedInitBootstrap struct {
	stages []bootstrapStage
	close  func() error
	ran    bool
	closed bool
	// Target contains only non-secret fields suitable for a caller's preview.
	Target bootstrapTarget
}
type bootstrapTarget struct {
	Host, Name, User, SSLMode string
	Port                      int
}
type bootstrapStage struct {
	name string
	run  func(context.Context) error
}

func prepareInitBootstrap(ctx context.Context, values map[string]string, input root.CreateRootInput) (*preparedInitBootstrap, error) {
	return prepareBootstrapWithConnection(ctx, values, input, db.ConnectQuiet)
}

func prepareBootstrapWithConnection(ctx context.Context, values map[string]string, input root.CreateRootInput, connect func(context.Context, *config.Config) (*gorm.DB, error)) (*preparedInitBootstrap, error) {
	if ctx.Err() != nil {
		return nil, errors.New("initial setup preparation canceled")
	}
	if err := root.ValidateInput(input); err != nil {
		return nil, errors.New("invalid root credentials")
	}
	cfg, env, target, err := bootstrapConfig(values)
	if err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("go"); err != nil {
		return nil, errors.New("initial setup requires Go")
	}
	if _, err := currentSeedModule(); err != nil {
		return nil, errors.New("initial setup module validation failed")
	}
	funcs, err := discoverSeedFunctions("seeds")
	if err != nil || len(funcs) == 0 {
		return nil, errors.New("initial setup requires seed resources")
	}
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return nil, errors.New("initial setup requires migration resources")
	}
	hasSQL := false
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			hasSQL = true
		}
	}
	if !hasSQL {
		return nil, errors.New("initial setup requires migration resources")
	}
	database, err := connect(ctx, cfg)
	if err != nil {
		return nil, errors.New("initial setup database preflight failed")
	}
	sqlDB, err := database.DB()
	if err != nil {
		_ = db.Close(database)
		return nil, errors.New("initial setup database preflight failed")
	}
	// A provider indexes migration resources but does not apply them here. Its
	// private logger prevents SQL/server error strings from leaking globally.
	provider, err := goose.NewProvider(goose.DialectPostgres, sqlDB, os.DirFS("migrations"), goose.WithSlog(slog.New(slog.NewTextHandler(io.Discard, nil))))
	if err != nil {
		_ = sqlDB.Close()
		return nil, errors.New("initial setup migration validation failed")
	}
	p := &preparedInitBootstrap{Target: target, close: sqlDB.Close}
	p.stages = []bootstrapStage{
		{"migrations", func(ctx context.Context) error { _, err := provider.Up(ctx); return err }},
		{"seeds", func(ctx context.Context) error {
			return runSeedsWithEnvironment(ctx, funcs, cli.Stdio{Out: io.Discard, Err: io.Discard}, env)
		}},
		{"root", func(ctx context.Context) error {
			creator := root.NewCreator(newRootStorage(database.WithContext(ctx)), pwd.Manager{})
			return creator.Create(input)
		}},
	}
	return p, nil
}

func (p *preparedInitBootstrap) Run(ctx context.Context, stdio cli.Stdio) error {
	if p.closed || p.ran {
		return errors.New("initial setup is no longer available")
	}
	p.ran = true
	for _, stage := range p.stages {
		if ctx.Err() != nil {
			return fmt.Errorf("initial setup %s failed", stage.name)
		}
		stdio.Printf("Running %s...\n", stage.name)
		err := stage.run(ctx)
		if stage.name == "root" && errors.Is(err, root.ErrRootAlreadyExists) {
			stdio.Println("Root user already exists. Skipping.")
			continue
		}
		if err != nil {
			return fmt.Errorf("initial setup %s failed", stage.name)
		}
	}
	return nil
}
func (p *preparedInitBootstrap) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true
	// Release closures holding typed credentials as well as the connection.
	p.stages = nil
	if p.close != nil {
		f := p.close
		p.close = nil
		if f() != nil {
			return errors.New("initial setup connection close failed")
		}
	}
	return nil
}

// Values, not process environment or dotenv, determine the complete DB target.
func bootstrapConfig(values map[string]string) (*config.Config, map[string]string, bootstrapTarget, error) {
	// pgx resolves a present service through external files, even when the
	// service value is empty. Do not let it select credentials outside this plan.
	if os.Getenv("PGSERVICE") != "" {
		return nil, nil, bootstrapTarget{}, errors.New("database service configuration is not supported for initial setup")
	}
	// Seed config.Load parses these selected settings before considering the
	// URL override. Reject malformed values here, before the caller applies files.
	for _, key := range []string{"APP_PORT", "DB_PORT", "SHUTDOWN_TIMEOUT_SECONDS", "DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "DB_CONN_MAX_LIFETIME_SECONDS", "LOG_MAX_SIZE", "LOG_MAX_BACKUPS", "LOG_MAX_AGE", "ACCESS_TTL_MINUTES", "REFRESH_TTL_DAYS", "REDIS_PORT", "REDIS_DB", "REDIS_DIAL_TIMEOUT_SECONDS", "RATE_LIMIT_GLOBAL_RPM", "RATE_LIMIT_LOGIN_RPM", "RATE_LIMIT_REFRESH_RPM", "RATE_LIMIT_WINDOW_SECONDS", "RATE_LIMIT_CLEANUP_INTERVAL_MINUTES"} {
		if value := values[key]; value != "" {
			if _, err := strconv.Atoi(value); err != nil {
				return nil, nil, bootstrapTarget{}, errors.New("invalid configuration")
			}
		}
	}
	cfg := &config.Config{}
	env := map[string]string{}
	for k, v := range values {
		env[k] = v
	}
	for _, k := range []string{"DATABASE_URL", "DB_HOST", "DB_PORT", "DB_NAME", "DB_USER", "DB_PASSWORD", "DB_SSL_MODE"} {
		env[k] = values[k]
	}
	cfg.DB.DatabaseURL = values["DATABASE_URL"]
	cfg.DB.Host = values["DB_HOST"]
	cfg.DB.Name = values["DB_NAME"]
	cfg.DB.User = values["DB_USER"]
	cfg.DB.Password = values["DB_PASSWORD"]
	cfg.DB.SSLMode = values["DB_SSL_MODE"]
	port, err := strconv.Atoi(values["DB_PORT"])
	if cfg.DB.DatabaseURL == "" && (err != nil || port < 1 || port > 65535 || cfg.DB.Host == "" || cfg.DB.Name == "" || cfg.DB.User == "" || cfg.DB.SSLMode == "") {
		return nil, nil, bootstrapTarget{}, errors.New("invalid database configuration")
	}
	cfg.DB.Port = port
	cfg.DB.MaxOpenConns = 5
	cfg.DB.MaxIdleConns = 2
	target := bootstrapTarget{cfg.DB.Host, cfg.DB.Name, cfg.DB.User, cfg.DB.SSLMode, cfg.DB.Port}
	if cfg.DB.DatabaseURL != "" {
		// Require an explicit URL, not keyword settings that could use pgservice.
		if !strings.HasPrefix(cfg.DB.DatabaseURL, "postgres://") && !strings.HasPrefix(cfg.DB.DatabaseURL, "postgresql://") {
			return nil, nil, bootstrapTarget{}, errors.New("invalid database configuration")
		}
		uri, e := url.Parse(cfg.DB.DatabaseURL)
		if e != nil || uri.Hostname() == "" || uri.User == nil || uri.User.Username() == "" || strings.Trim(uri.Path, "/") == "" {
			return nil, nil, bootstrapTarget{}, errors.New("invalid database configuration")
		}
		query := uri.Query()
		for _, key := range []string{"passfile", "service", "servicefile"} {
			if _, present := query[key]; present {
				return nil, nil, bootstrapTarget{}, errors.New("database URL must use explicit credentials")
			}
		}
		query.Set("passfile", os.DevNull)

		if query.Get("sslmode") == "" {
			query.Set("sslmode", "prefer")
		}
		if uri.Port() == "" && query.Get("port") == "" {
			query.Set("port", "5432")
		}
		parsedURI := *uri
		parsedURI.RawQuery = query.Encode()
		parsed, e := pgx.ParseConfig(parsedURI.String())
		if e != nil || parsed.Host == "" || parsed.Database == "" || parsed.User == "" {
			return nil, nil, bootstrapTarget{}, errors.New("invalid database configuration")
		}
		target.Host = parsed.Host
		target.Name = parsed.Database
		target.User = parsed.User
		target.Port = int(parsed.Port)
		target.SSLMode = uri.Query().Get("sslmode")
		if target.SSLMode == "" {
			target.SSLMode = "prefer"
		}
		password, _ := uri.User.Password()
		if entries, ok := uri.Query()["password"]; ok && len(entries) > 0 {
			password = entries[0]
		}
		env["PGPASSWORD"] = password
	}
	// Pin libpq defaults too: URL credentials may intentionally be empty, and
	// inherited pgpass/SSL settings must not make the seed select other values.
	if cfg.DB.DatabaseURL == "" {
		env["PGPASSWORD"] = cfg.DB.Password
	}
	env["PGPASSFILE"] = os.DevNull
	env["PGSSLMODE"] = target.SSLMode
	env["PGPORT"] = strconv.Itoa(target.Port)
	// The URL and all component keys are explicitly present even when empty,
	// preventing godotenv's non-overwrite loading from resurrecting a target.
	return cfg, env, target, nil
}
