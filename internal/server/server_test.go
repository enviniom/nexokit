package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/gin-gonic/gin"
	"log/slog"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		App:  config.AppConfig{Env: "test"},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	r := NewRouter(cfg, log, os.Stdout, HealthDeps{}, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !contains(body, `"success":true`) {
		t.Errorf("expected success=true in response body, got: %s", body)
	}
	if !contains(body, "API is healthy") {
		t.Errorf("expected 'API is healthy' in response body, got: %s", body)
	}
}

func TestNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		App:  config.AppConfig{Env: "test"},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	r := NewRouter(cfg, log, os.Stdout, HealthDeps{}, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/unknown", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestServerStartStop(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		App: config.AppConfig{Env: "test", Port: 0},
	}
	r := gin.New()
	srv := New(cfg, r)

	// Start in background
	go func() {
		_ = srv.Start()
	}()

	// Stop immediately
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Stop(ctx); err != nil {
		t.Fatalf("unexpected error stopping server: %v", err)
	}
}

func TestLiveEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		App:  config.AppConfig{Env: "test"},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	r := NewRouter(cfg, log, os.Stdout, HealthDeps{}, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health/live", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !contains(w.Body.String(), `"status":"alive"`) {
		t.Fatalf("expected status alive response, got: %s", w.Body.String())
	}
}

func TestReadyEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		App:  config.AppConfig{Env: "test"},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))

	r := NewRouter(cfg, log, os.Stdout, HealthDeps{}, nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 for nil DB dependency, got %d", w.Code)
	}
	if !contains(w.Body.String(), `"name":"database"`) {
		t.Fatalf("expected database dependency in response, got: %s", w.Body.String())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
