package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterMountsAuthEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	authMiddleware := func(c *gin.Context) {
		c.Next()
	}
	attachPermissions := func(c *gin.Context) {
		c.Next()
	}
	rateLimit := func(c *gin.Context) {
		c.Next()
	}

	container := &Container{}
	Register(r.Group("/api/v1"), container, authMiddleware, attachPermissions, rateLimit, rateLimit)

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "login", method: http.MethodPost, path: "/api/v1/auth/login"},
		{name: "refresh", method: http.MethodPost, path: "/api/v1/auth/refresh"},
		{name: "logout", method: http.MethodPost, path: "/api/v1/auth/logout"},
		{name: "me", method: http.MethodGet, path: "/api/v1/auth/me"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			if w.Code == http.StatusNotFound {
				t.Fatalf("expected route %s %s to be mounted", tt.method, tt.path)
			}
		})
	}
}
