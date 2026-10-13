package assign_permissions_to_role

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamrole "github.com/enviniom/nexokit/internal/modules/iam/core/role"
	"github.com/enviniom/nexokit/internal/platform/authctx"
	"github.com/enviniom/nexokit/internal/platform/roles"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/enviniom/nexokit/internal/shared"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	item  *iamrole.RolePermissionAssignmentResponse
	err   error
	calls *int
}

func (f fakeHandlerService) Assign(tenant.TenantContext, string, iamrole.AssignRolePermissionsRequest) (*iamrole.RolePermissionAssignmentResponse, error) {
	if f.calls != nil {
		(*f.calls)++
	}
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestHandlerTenantAdminWithoutGrants(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uint(7)
	repo := &fakeRepository{role: &iamrole.IAMRole{BaseModel: shared.BaseModel{ID: 11, PublicID: "custom"}, Slug: "custom", CompanyID: &id}, selected: map[string]bool{}, normalized: []string{}, ids: []uint{}}
	h := NewHandler(NewService(repo, nil))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "custom"}}
	tenant.SetGin(c, tenant.NewScoped(id, "own"))
	authctx.SetGin(c, &authctx.User{RoleSlug: roles.AdminRoleSlug, CompanyID: &id})
	c.Request = httptest.NewRequest("PATCH", "/roles/custom/permissions", bytes.NewBufferString(`{"permissions":[]}`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Handle(c)
	if w.Code != 200 {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	if repo.replacedRoleID != 11 {
		t.Fatal("assignment was not persisted")
	}
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"permissions":["roles.read"]}`

	tests := []struct {
		name       string
		service    Service
		body       string
		rejected   bool
		statusCode int
	}{
		{name: "bad json", service: fakeHandlerService{}, body: `{`, rejected: true, statusCode: http.StatusBadRequest},
		{name: "missing permissions", service: fakeHandlerService{}, body: `{}`, rejected: true, statusCode: http.StatusUnprocessableEntity},
		{name: "null permissions", service: fakeHandlerService{}, body: `{"permissions":null}`, rejected: true, statusCode: http.StatusUnprocessableEntity},
		{name: "empty permissions allowed", service: fakeHandlerService{item: &iamrole.RolePermissionAssignmentResponse{RoleID: "role-1", Permissions: []string{}, Catalog: []iamrole.RolePermissionGroupResponse{}}}, body: `{"permissions":[]}`, statusCode: http.StatusOK},
		{name: "success", service: fakeHandlerService{item: &iamrole.RolePermissionAssignmentResponse{RoleID: "role-1"}}, statusCode: http.StatusOK},
		{name: "maps domain forbidden to 403", service: fakeHandlerService{err: core.ErrSystemImmutable}, statusCode: http.StatusForbidden},
		{name: "maps invalid slug to 400", service: fakeHandlerService{err: core.ErrInvalidPermissionSlug}, statusCode: http.StatusBadRequest},
		{name: "maps unknown to 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			service := tt.service.(fakeHandlerService)
			service.calls = &calls
			h := NewHandler(service)
			requestBody := body
			if tt.body != "" {
				requestBody = tt.body
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "role-1"}}
			tenant.SetGin(c, tenant.NewRoot())
			c.Request = httptest.NewRequest(http.MethodPatch, "/iam/roles/role-1/permissions", bytes.NewBufferString(requestBody))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}
			var payload map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			if tt.rejected {
				if calls != 0 {
					t.Fatalf("service calls = %d, want 0", calls)
				}
				if payload["success"] != false || payload["message"] == "" || payload["message"] == nil {
					t.Fatalf("invalid rejection envelope: %#v", payload)
				}
				if tt.statusCode == 422 {
					fields, ok := payload["errors"].(map[string]any)
					if !ok || fields["permissions"] == nil {
						t.Fatalf("missing permissions error: %#v", payload)
					}
				}
			}
			if tt.statusCode == 200 {
				if payload["success"] != true || calls != 1 {
					t.Fatalf("success envelope/calls: %#v / %d", payload, calls)
				}
				data, ok := payload["data"].(map[string]any)
				if !ok || data["role_id"] != "role-1" || len(data) != 3 {
					t.Fatalf("assignment shape: %#v", payload)
				}
				for _, field := range []string{"permissions", "catalog"} {
					value, exists := data[field]
					if !exists {
						t.Fatalf("missing %s", field)
					}
					if tt.name == "success" && value != nil {
						t.Errorf("%s = %#v, want null", field, value)
					}
					if tt.name == "empty permissions allowed" {
						if values, ok := value.([]any); !ok || len(values) != 0 {
							t.Errorf("%s = %#v, want []", field, value)
						}
					}
				}
			}
		})
	}
}
