package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakeDB struct{ err error }

func (f fakeDB) PingContext(context.Context) error { return f.err }

type fakeCache struct{ err error }

func (f fakeCache) Get(context.Context, string) ([]byte, error) { return nil, f.err }

func TestLiveHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
	}{
		{name: "basic GET returns alive"},
		{name: "response does not depend on request details"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodGet, "/health/live", nil)
			c.Request = req

			liveHandler(c)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}

			var got LiveResponse
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("expected valid JSON, got error: %v", err)
			}
			if got.Status != "alive" {
				t.Fatalf("expected status=alive, got %q", got.Status)
			}
		})
	}
}

func TestReadyHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		deps       HealthDeps
		wantStatus int
		wantDeps   map[string]string
	}{
		{
			name: "all healthy",
			deps: HealthDeps{DB: fakeDB{}, CacheEnabled: false},
			wantStatus: http.StatusOK,
			wantDeps: map[string]string{"database": "healthy", "cache": "healthy"},
		},
		{
			name: "db fail",
			deps: HealthDeps{DB: fakeDB{err: errors.New("connection refused")}, CacheEnabled: false},
			wantStatus: http.StatusServiceUnavailable,
			wantDeps: map[string]string{"database": "unhealthy", "cache": "healthy"},
		},
		{
			name: "cache fail",
			deps: HealthDeps{DB: fakeDB{}, Cache: fakeCache{err: errors.New("cache unavailable")}, CacheEnabled: true},
			wantStatus: http.StatusServiceUnavailable,
			wantDeps: map[string]string{"database": "healthy", "cache": "unhealthy"},
		},
		{
			name: "cache disabled",
			deps: HealthDeps{DB: fakeDB{}, CacheEnabled: false},
			wantStatus: http.StatusOK,
			wantDeps: map[string]string{"database": "healthy", "cache": "healthy"},
		},
		{
			name: "nil db",
			deps: HealthDeps{CacheEnabled: false},
			wantStatus: http.StatusServiceUnavailable,
			wantDeps: map[string]string{"database": "unhealthy", "cache": "healthy"},
		},
		{
			name: "nil cache when enabled",
			deps: HealthDeps{DB: fakeDB{}, CacheEnabled: true},
			wantStatus: http.StatusServiceUnavailable,
			wantDeps: map[string]string{"database": "healthy", "cache": "unhealthy"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
			c.Request = req

			readyHandler(tt.deps)(c)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			var got ReadyResponse
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("invalid json: %v", err)
			}

			gotMap := map[string]string{}
			for _, dep := range got.Dependencies {
				gotMap[dep.Name] = dep.Status
			}

			for k, v := range tt.wantDeps {
				if gotMap[k] != v {
					t.Fatalf("expected %s status=%s, got=%s", k, v, gotMap[k])
				}
			}
		})
	}
}

func TestHealthHandler(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	c.Request = req

	healthHandler(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	checks := []string{`"success":true`, `"message":"API is healthy"`, `"status":"ok"`}
	for _, check := range checks {
		if !contains(body, check) {
			t.Fatalf("expected response to contain %s, got %s", check, body)
		}
	}
}
