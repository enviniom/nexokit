package config

import (
	"os"
	"testing"
)

func TestIsLocal(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "local")
	cfg, _ := Load()
	if !cfg.IsLocal() {
		t.Error("expected IsLocal() to be true")
	}
	if cfg.IsProduction() {
		t.Error("expected IsProduction() to be false")
	}
}

func TestIsProduction(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "production")
	cfg, _ := Load()
	if !cfg.IsProduction() {
		t.Error("expected IsProduction() to be true")
	}
	if cfg.IsLocal() {
		t.Error("expected IsLocal() to be false")
	}
}

func TestIsTest(t *testing.T) {
	os.Clearenv()
	os.Setenv("APP_ENV", "test")
	cfg, _ := Load()
	if !cfg.IsTest() {
		t.Error("expected IsTest() to be true")
	}
}
