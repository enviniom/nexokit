package permissions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/permissions/core"
	"github.com/enviniom/nexokit/internal/modules/permissions/list_permissions"
	"github.com/enviniom/nexokit/internal/modules/permissions/update_permission"
	"github.com/enviniom/nexokit/internal/modules/permissions/view_permission"
	"github.com/gin-gonic/gin"
)

type listSvc struct{}

func (listSvc) ListGrouped() ([]core.PermissionGroupResponse, error) { return []core.PermissionGroupResponse{}, nil }

type viewSvc struct{}

func (viewSvc) GetByPublicID(string) (*core.PermissionResponse, error) { return &core.PermissionResponse{}, nil }

type updateSvc struct{}

func (updateSvc) Update(string, core.UpdatePermissionRequest) (*core.PermissionResponse, error) {
	return &core.PermissionResponse{}, nil
}

func testContainer() *Container {
	return &Container{
		ListHandler:   list_permissions.NewHandler(list_permissions.NewService(listRepo{})),
		ViewHandler:   view_permission.NewHandler(viewSvc{}),
		UpdateHandler: update_permission.NewHandler(updateSvc{}),
	}
}

type listRepo struct{}

func (listRepo) ListAll() ([]core.Permission, error) { return []core.Permission{}, nil }

func TestRegisterAppliesPermissionManageGuards(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct{ method, path string }{{http.MethodGet, "/permissions"}, {http.MethodGet, "/permissions/perm1"}, {http.MethodPut, "/permissions/perm1"}}
	for _, tt := range cases {
		var got string
		r := gin.New()
		Register(r.Group(""), testContainer(), func(slug string) gin.HandlerFunc {
			return func(c *gin.Context) { got = slug; c.Status(http.StatusNoContent); c.Abort() }
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
		if w.Code != http.StatusNoContent || got != "permissions.manage" {
			t.Fatalf("unexpected guard status/slug: %d %s", w.Code, got)
		}
	}
}

func TestUnregisteredEndpointsReturnNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	Register(r.Group(""), testContainer(), func(string) gin.HandlerFunc {
		return func(c *gin.Context) { c.Status(http.StatusNoContent); c.Abort() }
	})

	for _, tt := range []struct{ method, path string }{{http.MethodPost, "/permissions"}, {http.MethodDelete, "/permissions/perm1"}} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	}
}
