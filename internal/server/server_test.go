package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/platform/apperror"
	"github.com/enviniom/nexokit/internal/platform/response"
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

	r := NewRouter(cfg, log, log, os.Stdout, HealthDeps{}, nil)
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

	r := NewRouter(cfg, log, log, os.Stdout, HealthDeps{}, nil)
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

	r := NewRouter(cfg, log, log, os.Stdout, HealthDeps{}, nil)
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

	r := NewRouter(cfg, log, log, os.Stdout, HealthDeps{}, nil)
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

func TestRouterPanicProducesSingleErrorLog(t *testing.T) {
	// Gin is in debug mode, but AppConfig.Env is production. The response must
	// still redact debug details because config.Env is the source of truth.
	gin.SetMode(gin.DebugMode)
	defer gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		App:  config.AppConfig{Env: "production"},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}))
	errorLog := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	r := NewRouter(cfg, log, errorLog, io.Discard, HealthDeps{}, func(v1 *gin.RouterGroup) {
		v1.GET("/panic", func(c *gin.Context) {
			panic("router boom")
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/panic", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
	body := w.Body.String()
	if contains(body, "router boom") {
		t.Errorf("release response leaked panic text: %s", body)
	}
	if contains(body, "\"debug\"") {
		t.Errorf("release response included debug field: %s", body)
	}

	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	var count int
	for _, line := range lines {
		if len(line) > 0 {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one error log line, got %d: %s", count, buf.String())
	}
	if !bytes.Contains(buf.Bytes(), []byte("panic: router boom")) {
		t.Errorf("expected log to contain panic text, got %s", buf.String())
	}
}

func TestRouterHandledAppErrorProducesSingleErrorLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		App:  config.AppConfig{Env: "production"},
		CORS: config.CORSConfig{AllowedOrigins: "*"},
	}
	var buf bytes.Buffer
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelWarn}))
	errorLog := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))

	r := NewRouter(cfg, log, errorLog, io.Discard, HealthDeps{}, func(v1 *gin.RouterGroup) {
		v1.GET("/handled", func(c *gin.Context) {
			baseErr := apperror.Conflict(
				apperror.Code("catalog.product_conflict"),
				"Product conflict",
				errors.New("database constraint root"),
			)
			err := apperror.Wrap(baseErr, "Product cannot be updated", errors.New("repository unique constraint violation"))
			response.HandleError(c, err)
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/handled", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
	body := w.Body.String()
	if !contains(body, `"message":"Product cannot be updated"`) {
		t.Fatalf("expected public message in response, got: %s", body)
	}
	if contains(body, "repository unique constraint violation") {
		t.Errorf("production response leaked internal error: %s", body)
	}
	if contains(body, "\"debug\"") {
		t.Errorf("production response included debug field: %s", body)
	}

	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	var count int
	for _, line := range lines {
		if len(line) > 0 {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one error log line, got %d: %s", count, buf.String())
	}

	for _, want := range []string{
		`"method":"GET"`,
		`"path":"/api/v1/handled"`,
		`"status":409`,
		`"code":"catalog.product_conflict"`,
		`"public_message":"Product cannot be updated"`,
		"database constraint root",
		"repository unique constraint violation",
		`"request_id":`,
	} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("expected log to contain %s, got %s", want, buf.String())
		}
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
