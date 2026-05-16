package commands

import (
	"os"
	"testing"

	"github.com/enviniom/nexokit/internal/config"
)

func TestToDisplayConfig_MasksSecrets(t *testing.T) {
	os.Clearenv()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	display := toDisplayConfig(cfg)

	if display.DB.Password != "***masked***" {
		t.Errorf("expected password masked, got %q", display.DB.Password)
	}
	if display.DB.DatabaseURL != "***masked***" {
		t.Errorf("expected database_url masked, got %q", display.DB.DatabaseURL)
	}
}

func TestToDisplayConfig_PreservesNonSecrets(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_NAME", "testapp")
	os.Setenv("APP_PORT", "9090")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	display := toDisplayConfig(cfg)

	if display.App.Name != "testapp" {
		t.Errorf("expected app name preserved, got %q", display.App.Name)
	}
	if display.App.Port != 9090 {
		t.Errorf("expected app port preserved, got %d", display.App.Port)
	}
}
