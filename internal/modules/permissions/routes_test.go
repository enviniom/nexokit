package permissions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterAppliesPermissionManageGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Registered endpoints should trigger the permission guard
	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "list permissions", method: http.MethodGet, path: "/permissions"},
		{name: "view permission", method: http.MethodGet, path: "/permissions/perm1"},
		{name: "update permission", method: http.MethodPut, path: "/permissions/perm1"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			r := gin.New()
			Register(r.Group(""), NewHandler(NewService(&fakeRepository{})), func(slug string) gin.HandlerFunc {
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
			if got != "permissions.manage" {
				t.Fatalf("expected permissions.manage guard, got %q", got)
			}
		})
	}
}

func TestUnregisteredEndpointsReturnNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "create permission is unregistered", method: http.MethodPost, path: "/permissions"},
		{name: "delete permission is unregistered", method: http.MethodDelete, path: "/permissions/perm1"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			Register(r.Group(""), NewHandler(NewService(&fakeRepository{})), func(slug string) gin.HandlerFunc {
				return func(c *gin.Context) {
					c.Status(http.StatusNoContent)
					c.Abort()
				}
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))

			if w.Code != http.StatusNotFound {
				t.Fatalf("expected unregistered route to return 404, got %d", w.Code)
			}
		})
	}
}
