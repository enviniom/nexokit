package iam

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/enviniom/nexokit/internal/platform/password"
	"github.com/enviniom/nexokit/internal/platform/token"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegisterOwnsAuthenticationRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	container := NewContainer(db, nil, password.Manager{}, token.NewManager("test-secret", time.Hour), 24*time.Hour)
	r := gin.New()
	v1 := r.Group("/api/v1")
	pass := func(*gin.Context) {}
	permission := func(string) gin.HandlerFunc { return pass }
	stop := func(status int) gin.HandlerFunc {
		return func(c *gin.Context) { c.AbortWithStatus(status) }
	}
	global := v1.Group("", stop(http.StatusForbidden))
	tenant := v1.Group("", stop(http.StatusForbidden))
	Register(v1, global, tenant, container, permission, permission, stop(http.StatusUnauthorized), stop(http.StatusTooManyRequests), stop(http.StatusServiceUnavailable))

	for _, tc := range []struct {
		name, method, path string
		status             int
	}{
		{"login", http.MethodPost, "/api/v1/auth/login", http.StatusTooManyRequests},
		{"refresh", http.MethodPost, "/api/v1/auth/refresh", http.StatusServiceUnavailable},
		{"logout", http.MethodPost, "/api/v1/auth/logout", http.StatusBadRequest},
		{"session", http.MethodGet, "/api/v1/auth/me", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			found := false
			for _, route := range r.Routes() {
				if route.Method == tc.method && route.Path == tc.path {
					found = true
				}
			}
			if !found {
				t.Fatalf("IAM registration did not mount %s %s", tc.method, tc.path)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.status, w.Body.String())
			}
		})
	}
}
