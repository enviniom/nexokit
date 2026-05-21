package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/middleware"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/gin-gonic/gin"
)

func TestRBACIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		user       *authctx.User
		expected   int
		assertNext bool
	}{
		{name: "user with permission accesses resource", user: &authctx.User{PublicID: "u1", Permissions: []string{"users.index"}, IsActive: true}, expected: http.StatusOK},
		{name: "user without permission gets 403", user: &authctx.User{PublicID: "u2", Permissions: []string{"users.view"}, IsActive: true}, expected: http.StatusForbidden},
		{name: "unauthenticated gets 401", user: nil, expected: http.StatusUnauthorized},
		{name: "root has all permissions", user: &authctx.User{PublicID: "root", IsRoot: true, Permissions: []string{"*"}, IsActive: true}, expected: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tt.user != nil {
					authctx.SetGin(c, tt.user)
				}
				c.Next()
			})
			r.GET("/protected", middleware.RequirePermission("users.index"), func(c *gin.Context) { c.Status(http.StatusOK) })
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			r.ServeHTTP(w, req)
			if w.Code != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, w.Code)
			}
		})
	}
}
