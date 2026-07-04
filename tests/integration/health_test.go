package integration_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/config"
	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/server"
)

func TestHealthIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cfg := &config.Config{App: config.AppConfig{Env: "test"}, CORS: config.CORSConfig{AllowedOrigins: "*"}}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := server.NewRouter(cfg, log, log, io.Discard, server.HealthDeps{}, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp response.APIResponse[map[string]string]
	mustDecode(t, w.Body.Bytes(), &resp)
	if resp.Data["status"] != "ok" {
		t.Fatalf("expected health status ok, got %+v", resp.Data)
	}
}
