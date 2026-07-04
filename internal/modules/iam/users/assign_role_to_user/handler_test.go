package assign_role_to_user

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	resp *core.UserResponse
	err  error
}

func (f fakeHandlerService) ChangeRole(tenant.TenantContext, string, string, core.ChangeUserRoleRequest) (*core.UserResponse, error) {
	return f.resp, f.err
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
			service:    fakeHandlerService{resp: &core.UserResponse{PublicID: "user-1", Name: "Alice", RoleID: 20, RoleName: "Admin"}},
			body:       core.ChangeUserRoleRequest{RoleID: "role-admin"},
			statusCode: http.StatusOK,
		},
		{
			name:        "maps not found to 404 with platform message",
			service:     fakeHandlerService{err: core.ErrNotFound},
			body:        core.ChangeUserRoleRequest{RoleID: "role-admin"},
			statusCode:  http.StatusNotFound,
			wantMessage: "Recurso no encontrado",
		},
		{
			name:        "maps forbidden to 403 with platform message",
			service:     fakeHandlerService{err: core.ErrForbidden},
			body:        core.ChangeUserRoleRequest{RoleID: "role-admin"},
			statusCode:  http.StatusForbidden,
			wantMessage: "Acceso denegado",
		},
		{
			name:        "maps forbidden role assignment to 403 with module message",
			service:     fakeHandlerService{err: core.ErrForbiddenRoleAssignment},
			body:        core.ChangeUserRoleRequest{RoleID: "role-root"},
			statusCode:  http.StatusForbidden,
			wantMessage: "forbidden role assignment",
		},
		{
			name:        "maps forbidden company scope to 403 with slice message",
			service:     fakeHandlerService{err: core.ErrForbiddenCompanyScope},
			body:        core.ChangeUserRoleRequest{RoleID: "role-other"},
			statusCode:  http.StatusForbidden,
			wantMessage: "forbidden company scope",
		},
		{
			name:       "maps unknown errors to 500",
			service:    fakeHandlerService{err: errors.New("db down")},
			body:       core.ChangeUserRoleRequest{RoleID: "role-admin"},
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
			body:       core.ChangeUserRoleRequest{RoleID: ""},
			statusCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
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

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
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
