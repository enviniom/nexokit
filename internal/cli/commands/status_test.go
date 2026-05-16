package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCountMigrationFiles(t *testing.T) {
	dir := t.TempDir()

	_ = os.WriteFile(filepath.Join(dir, "001_init.sql"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(dir, "002_add.sql"), []byte(""), 0644)
	_ = os.WriteFile(filepath.Join(dir, "README.md"), []byte(""), 0644)
	_ = os.Mkdir(filepath.Join(dir, "subdir"), 0755)

	count := countMigrationFiles(dir)
	if count != 2 {
		t.Errorf("expected 2 migration files, got %d", count)
	}
}

func TestCountMigrationFiles_MissingDir(t *testing.T) {
	count := countMigrationFiles("/nonexistent/path")
	if count != 0 {
		t.Errorf("expected 0 for missing dir, got %d", count)
	}
}

func TestCountMigrationFiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	count := countMigrationFiles(dir)
	if count != 0 {
		t.Errorf("expected 0 for empty dir, got %d", count)
	}
}
