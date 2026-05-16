package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Ensure no env vars are leaking from the environment
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Name != "nexokit" {
		t.Errorf("expected default APP_NAME 'nexokit', got %s", cfg.App.Name)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected default APP_PORT 8080, got %d", cfg.App.Port)
	}
	if cfg.DB.Host != "localhost" {
		t.Errorf("expected default DB_HOST 'localhost', got %s", cfg.DB.Host)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("expected default DB_PORT 5432, got %d", cfg.DB.Port)
	}
	if cfg.Shutdown.Timeout.Seconds() != 30 {
		t.Errorf("expected default shutdown timeout 30s, got %v", cfg.Shutdown.Timeout)
	}
	if cfg.Cache.Driver != "none" {
		t.Errorf("expected default CACHE_DRIVER 'none', got %s", cfg.Cache.Driver)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("expected default LOG_LEVEL 'info', got %s", cfg.Log.Level)
	}
	if cfg.Log.File != "logs/app.log" {
		t.Errorf("expected default LOG_FILE 'logs/app.log', got %s", cfg.Log.File)
	}
	if cfg.Log.MaxSize != 100 {
		t.Errorf("expected default LOG_MAX_SIZE 100, got %d", cfg.Log.MaxSize)
	}
	if cfg.Log.MaxBackups != 3 {
		t.Errorf("expected default LOG_MAX_BACKUPS 3, got %d", cfg.Log.MaxBackups)
	}
	if cfg.Log.MaxAge != 28 {
		t.Errorf("expected default LOG_MAX_AGE 28, got %d", cfg.Log.MaxAge)
	}
	if !cfg.Log.Compress {
		t.Error("expected default LOG_COMPRESS true")
	}
	if cfg.Log.GinFile != "logs/gin.log" {
		t.Errorf("expected default LOG_GIN_FILE 'logs/gin.log', got %s", cfg.Log.GinFile)
	}
	if cfg.Log.ErrorFile != "logs/error.log" {
		t.Errorf("expected default LOG_ERROR_FILE 'logs/error.log', got %s", cfg.Log.ErrorFile)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_NAME", "testapp")
	os.Setenv("APP_PORT", "3000")
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("SHUTDOWN_TIMEOUT_SECONDS", "10")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FILE", "logs/test.log")
	os.Setenv("LOG_MAX_SIZE", "50")
	os.Setenv("LOG_MAX_BACKUPS", "5")
	os.Setenv("LOG_MAX_AGE", "14")
	os.Setenv("LOG_COMPRESS", "false")
	os.Setenv("LOG_GIN_FILE", "logs/test-gin.log")
	os.Setenv("LOG_ERROR_FILE", "logs/test-error.log")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.App.Name != "testapp" {
		t.Errorf("expected APP_NAME 'testapp', got %s", cfg.App.Name)
	}
	if cfg.App.Port != 3000 {
		t.Errorf("expected APP_PORT 3000, got %d", cfg.App.Port)
	}
	if cfg.DB.Host != "db.example.com" {
		t.Errorf("expected DB_HOST 'db.example.com', got %s", cfg.DB.Host)
	}
	if cfg.DB.Port != 5433 {
		t.Errorf("expected DB_PORT 5433, got %d", cfg.DB.Port)
	}
	if cfg.Shutdown.Timeout.Seconds() != 10 {
		t.Errorf("expected shutdown timeout 10s, got %v", cfg.Shutdown.Timeout)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("expected LOG_LEVEL 'debug', got %s", cfg.Log.Level)
	}
	if cfg.Log.File != "logs/test.log" {
		t.Errorf("expected LOG_FILE 'logs/test.log', got %s", cfg.Log.File)
	}
	if cfg.Log.MaxSize != 50 {
		t.Errorf("expected LOG_MAX_SIZE 50, got %d", cfg.Log.MaxSize)
	}
	if cfg.Log.MaxBackups != 5 {
		t.Errorf("expected LOG_MAX_BACKUPS 5, got %d", cfg.Log.MaxBackups)
	}
	if cfg.Log.MaxAge != 14 {
		t.Errorf("expected LOG_MAX_AGE 14, got %d", cfg.Log.MaxAge)
	}
	if cfg.Log.Compress {
		t.Error("expected LOG_COMPRESS false")
	}
	if cfg.Log.GinFile != "logs/test-gin.log" {
		t.Errorf("expected LOG_GIN_FILE 'logs/test-gin.log', got %s", cfg.Log.GinFile)
	}
	if cfg.Log.ErrorFile != "logs/test-error.log" {
		t.Errorf("expected LOG_ERROR_FILE 'logs/test-error.log', got %s", cfg.Log.ErrorFile)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_PORT", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid APP_PORT")
	}
}

func TestLoad_InvalidShutdownTimeout(t *testing.T) {
	os.Clearenv()
	os.Setenv("SHUTDOWN_TIMEOUT_SECONDS", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid SHUTDOWN_TIMEOUT_SECONDS")
	}
}

func TestLoad_AuthDefaults(t *testing.T) {
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Auth.PASETOKey != "" {
		t.Errorf("expected default PASETO_KEY empty, got %s", cfg.Auth.PASETOKey)
	}
	if cfg.Auth.AccessTTLMinutes != 15 {
		t.Errorf("expected default ACCESS_TTL_MINUTES 15, got %d", cfg.Auth.AccessTTLMinutes)
	}
	if cfg.Auth.RefreshTTLDays != 7 {
		t.Errorf("expected default REFRESH_TTL_DAYS 7, got %d", cfg.Auth.RefreshTTLDays)
	}
}

func TestLoad_AuthCustomValues(t *testing.T) {
	os.Clearenv()
	os.Setenv("PASETO_KEY", "super-secret-32-byte-key!!")
	os.Setenv("ACCESS_TTL_MINUTES", "30")
	os.Setenv("REFRESH_TTL_DAYS", "14")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Auth.PASETOKey != "super-secret-32-byte-key!!" {
		t.Errorf("expected PASETO_KEY 'super-secret-32-byte-key!!', got %s", cfg.Auth.PASETOKey)
	}
	if cfg.Auth.AccessTTLMinutes != 30 {
		t.Errorf("expected ACCESS_TTL_MINUTES 30, got %d", cfg.Auth.AccessTTLMinutes)
	}
	if cfg.Auth.RefreshTTLDays != 14 {
		t.Errorf("expected REFRESH_TTL_DAYS 14, got %d", cfg.Auth.RefreshTTLDays)
	}
}

func TestLoad_InvalidAccessTTL(t *testing.T) {
	os.Clearenv()
	os.Setenv("ACCESS_TTL_MINUTES", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid ACCESS_TTL_MINUTES")
	}
}

func TestLoad_InvalidLogMaxSize(t *testing.T) {
	os.Clearenv()
	os.Setenv("LOG_MAX_SIZE", "big")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid LOG_MAX_SIZE")
	}
}
