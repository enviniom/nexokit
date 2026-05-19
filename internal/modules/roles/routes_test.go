package roles

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterAppliesRolePermissionGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{name: "list roles", method: http.MethodGet, path: "/roles", expected: "roles.index"},
		{name: "view role", method: http.MethodGet, path: "/roles/role1", expected: "roles.view"},
		{name: "view role permissions", method: http.MethodGet, path: "/roles/role1/permissions", expected: "roles.view"},
		{name: "assign role permissions", method: http.MethodPut, path: "/roles/role1/permissions", expected: "roles.assign_permissions"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			r := gin.New()
			Register(r.Group(""), NewHandler(&fakeService{}), func(slug string) gin.HandlerFunc {
				return func(c *gin.Context) {
					got = slug
					c.Status(http.StatusNoContent)
					c.Abort()
				}
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))

			if w.Code != http.StatusNoContent {
				t.Fatalf("expected route guard to abort with 204, got %d", w.Code)
			}
			if got != tt.expected {
				t.Fatalf("expected guard %q, got %q", tt.expected, got)
			}
		})
	}
}
