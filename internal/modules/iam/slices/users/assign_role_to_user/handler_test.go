package assign_role_to_user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	iamuser "github.com/enviniom/nexokit/internal/modules/iam/core/user"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	resp  *iamuser.UserResponse
	err   error
	calls *int
}

func (f fakeHandlerService) ChangeRole(tenant.TenantContext, string, string, iamuser.ChangeUserRoleRequest) (*iamuser.UserResponse, error) {
	if f.calls != nil {
		(*f.calls)++
	}
	return f.resp, f.err
}

func TestHandlerMissingTenant(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPatch, "/", bytes.NewBufferString(`{"role_id":"role-admin"}`))
	c.Request.Header.Set("Content-Type", "application/json")
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

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		service     Service
		body        any
		statusCode  int
		wantMessage string
	}{
		{
			name:       "returns success on role assignment",
			service:    fakeHandlerService{resp: &iamuser.UserResponse{PublicID: "user-1", Name: "Alice", RoleID: 20, RoleName: "Admin"}},
			body:       iamuser.ChangeUserRoleRequest{RoleID: "role-admin"},
			statusCode: http.StatusOK,
		},
		{
			name:        "maps not found to 404 with platform message",
			service:     fakeHandlerService{err: core.ErrNotFound},
			body:        iamuser.ChangeUserRoleRequest{RoleID: "role-admin"},
			statusCode:  http.StatusNotFound,
			wantMessage: "Recurso no encontrado",
		},
		{
			name:        "maps forbidden to 403 with platform message",
			service:     fakeHandlerService{err: core.ErrForbidden},
			body:        iamuser.ChangeUserRoleRequest{RoleID: "role-admin"},
			statusCode:  http.StatusForbidden,
			wantMessage: "Acceso denegado",
		},
		{
			name:        "maps forbidden role assignment to 403 with module message",
			service:     fakeHandlerService{err: core.ErrForbiddenRoleAssignment},
			body:        iamuser.ChangeUserRoleRequest{RoleID: "role-root"},
			statusCode:  http.StatusForbidden,
			wantMessage: "forbidden role assignment",
		},
		{
			name:        "maps forbidden company scope to 403 with slice message",
			service:     fakeHandlerService{err: core.ErrForbiddenCompanyScope},
			body:        iamuser.ChangeUserRoleRequest{RoleID: "role-other"},
			statusCode:  http.StatusForbidden,
			wantMessage: "forbidden company scope",
		},
		{
			name:       "maps unknown errors to 500",
			service:    fakeHandlerService{err: errors.New("db down")},
			body:       iamuser.ChangeUserRoleRequest{RoleID: "role-admin"},
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "returns 400 on invalid JSON",
			service:    fakeHandlerService{},
			body:       "not-json",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "returns 422 on missing role_id",
			service:    fakeHandlerService{},
			body:       iamuser.ChangeUserRoleRequest{RoleID: ""},
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		for _, scope := range []tenant.TenantContext{tenant.NewRoot(), tenant.NewScoped(7, "selected")} {
			t.Run(tt.name+scope.CompanySlug, func(t *testing.T) {
				calls := 0
				service := tt.service.(fakeHandlerService)
				service.calls = &calls
				h := NewHandler(service)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Params = gin.Params{{Key: "id", Value: "user-1"}}

				var bodyBytes []byte
				switch v := tt.body.(type) {
				case string:
					bodyBytes = []byte(v)
				default:
					bodyBytes, _ = json.Marshal(v)
				}
				c.Request = httptest.NewRequest(http.MethodPatch, "/iam/users/user-1/role", bytes.NewReader(bodyBytes))
				c.Request.Header.Set("Content-Type", "application/json")

				tenant.SetGin(c, scope)
				h.Handle(c)

				if w.Code != tt.statusCode {
					t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
				}

				if tt.name == "returns 400 on invalid JSON" || tt.name == "returns 422 on missing role_id" {
					if calls != 0 {
						t.Fatalf("service calls = %d, want 0", calls)
					}
					var envelope map[string]any
					if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
						t.Fatal(err)
					}
					if envelope["success"] != false || envelope["message"] == "" || envelope["message"] == nil {
						t.Fatalf("invalid rejection envelope: %#v", envelope)
					}
					if tt.statusCode == 422 {
						fields, ok := envelope["errors"].(map[string]any)
						if !ok || fields["role_id"] == nil {
							t.Fatalf("missing role_id error: %#v", envelope)
						}
					}
				}
				if tt.wantMessage == "" {
					return
				}
				var payload map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Fatalf("decode response body: %v", err)
				}
				if got := payload["message"]; got != tt.wantMessage {
					t.Errorf("message = %q, want %q", got, tt.wantMessage)
				}
			})
		}
	}
}
