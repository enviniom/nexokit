package commands

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/enviniom/nexokit/internal/cli"
	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/infra/db"
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

func TestCreateRootCommand_StorageSafety(t *testing.T) {
	// This test verifies that create-root fails safely when storage is not wired,
	// even when database is available.
	if !dbURLAvailable() {
		t.Skip("skipping: no database environment configured")
	}

	var out bytes.Buffer
	stdio := cli.Stdio{Out: &out}
	cmd := &CreateRootCommand{}

	err := cmd.Run(context.Background(), []string{"--email", "root@test.com", "--password", "Password1"}, stdio)
	if err == nil {
		t.Fatal("expected error because storage is not wired")
	}
	if !bytes.Contains([]byte(err.Error()), []byte("not yet wired")) {
		t.Errorf("expected 'not yet wired' error, got %v", err)
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
