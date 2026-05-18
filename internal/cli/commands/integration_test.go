package commands

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
	"github.com/enviniom/nexokit/internal/modules/roles"
	"github.com/enviniom/nexokit/internal/modules/users"
	"gorm.io/gorm"
)

func dbURLAvailable() bool {
	return os.Getenv("DATABASE_URL") != "" || os.Getenv("DB_HOST") != ""
}

func TestMigrateCommand_Up(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if !dbURLAvailable() {
		t.Skip("skipping: no database environment configured")
	}

	cmd := &MigrateCommand{}
	stdio := cli.Stdio{Out: &bytes.Buffer{}, Err: &bytes.Buffer{}}

	err := cmd.Run(context.Background(), []string{"status"}, stdio)
	if err != nil {
		t.Fatalf("migrate status failed: %v", err)
	}
}

func TestMigrateCommand_CreateAndValidate(t *testing.T) {
	tmpDir := t.TempDir()
	cwd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(cwd)

	_ = os.MkdirAll("migrations", 0755)

	cmd := &MigrateCommand{}
	stdio := cli.Stdio{}

	err := cmd.Run(context.Background(), []string{"create", "integration_test"}, stdio)
	if err != nil {
		t.Fatalf("migrate create failed: %v", err)
	}
}

func TestCreateRootCommand_IdempotentRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if !dbURLAvailable() {
		t.Skip("skipping: no database environment configured")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close(database)

	// Ensure schema exists
	if err := database.AutoMigrate(&roles.Role{}, &users.User{}); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}

	// Seed root role if missing
	var rootRole roles.Role
	if err := database.Where("slug = ?", roles.RootRoleSlug).First(&rootRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rootRole = roles.Role{Name: "root", Slug: "root", IsSystem: true}
			if createErr := database.Create(&rootRole).Error; createErr != nil {
				t.Fatalf("failed to seed root role: %v", createErr)
			}
		} else {
			t.Fatalf("failed to query root role: %v", err)
		}
	}

	cmd := &CreateRootCommand{}
	var out bytes.Buffer
	stdio := cli.Stdio{In: strings.NewReader(""), Out: &out}

	// First run should create root
	err = cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "Password1"}, stdio)
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}

	// Second run should be idempotent
	out.Reset()
	err = cmd.Run(context.Background(), []string{"--name", "Root", "--email", "root@test.com", "--password", "Password1"}, stdio)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if !strings.Contains(out.String(), "already exists") {
		t.Errorf("expected idempotency message, got %q", out.String())
	}
}

func TestMigrateCommand_DBConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	if !dbURLAvailable() {
		t.Skip("skipping: no database environment configured")
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close(database)

	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}
}
