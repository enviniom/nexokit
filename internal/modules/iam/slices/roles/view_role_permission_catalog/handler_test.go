package view_role_permission_catalog

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	data []iamrole.RolePermissionGroupResponse
	err  error
}

func (f fakeHandlerService) View(tenant.TenantContext, string) ([]iamrole.RolePermissionGroupResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.data, nil
}

func TestHandlerMissingTenant(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	defer func() {
		if p := recover(); p != nil {
			t.Errorf("service invoked without context: %v", p)
		}
	}()
	NewHandler(nil).Handle(c)
	if w.Code != 400 {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandlerSafeCatalogContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name string
		data []iamrole.RolePermissionGroupResponse
		want string
	}{
		{name: "nil catalog", want: `null`},
		{name: "empty catalog", data: []iamrole.RolePermissionGroupResponse{}, want: `[]`},
		{name: "nil permissions", data: []iamrole.RolePermissionGroupResponse{{Module: "users"}}, want: `[{"module":"users","permissions":null}]`},
		{name: "empty permissions", data: []iamrole.RolePermissionGroupResponse{{Module: "users", Permissions: []iamrole.RolePermissionResponse{}}}, want: `[{"module":"users","permissions":[]}]`},
		{name: "safe granted permission", data: []iamrole.RolePermissionGroupResponse{{Module: "users", Permissions: []iamrole.RolePermissionResponse{{PublicID: "perm-1", Slug: "users.read", Name: "Read users", Module: "users", Action: "read", Description: "desc", IsSystem: true, DisplayOrder: 3, Granted: true}}}}, want: `[{"module":"users","permissions":[{"public_id":"perm-1","slug":"users.read","name":"Read users","module":"users","action":"read","description":"desc","is_system":true,"display_order":3,"granted":true}]}]`},
		{name: "ungranted retained", data: []iamrole.RolePermissionGroupResponse{{Module: "users", Permissions: []iamrole.RolePermissionResponse{{PublicID: "perm-2", Slug: "users.write", Module: "users", Action: "write"}}}}, want: `[{"module":"users","permissions":[{"public_id":"perm-2","slug":"users.write","name":"","module":"users","action":"write","description":"","is_system":false,"display_order":0,"granted":false}]}]`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "role-1"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/iam/roles/role-1/permissions", nil)
			tenant.SetGin(c, tenant.NewRoot())
			NewHandler(fakeHandlerService{data: tt.data}).Handle(c)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			var payload map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if payload["success"] != true {
				t.Fatalf("success envelope: %#v", payload)
			}
			var want any
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			data, exists := payload["data"]
			if !exists || !reflect.DeepEqual(data, want) {
				t.Fatalf("safe catalog = %#v, want %#v", data, want)
			}
		})
	}
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		service    Service
		statusCode int
	}{
		{name: "returns success payload", service: fakeHandlerService{data: []iamrole.RolePermissionGroupResponse{{Module: "users"}}}, statusCode: http.StatusOK},
		{name: "maps not found to 404", service: fakeHandlerService{err: core.ErrNotFound}, statusCode: http.StatusNotFound},
		{name: "maps unknown errors to 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		for _, scope := range []tenant.TenantContext{tenant.NewRoot(), tenant.NewScoped(7, "selected")} {
			t.Run(tt.name+scope.CompanySlug, func(t *testing.T) {
				h := NewHandler(tt.service)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{{Key: "id", Value: "role-1"}}
				c.Request = httptest.NewRequest(http.MethodGet, "/iam/roles/role-1/permissions", nil)

				tenant.SetGin(c, scope)
				h.Handle(c)

				if w.Code != tt.statusCode {
					t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
				}
			})
		}
	}
}
