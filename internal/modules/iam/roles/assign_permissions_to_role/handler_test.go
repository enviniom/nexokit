package assign_permissions_to_role

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	item *core.RolePermissionAssignmentResponse
	err  error
}

func (f fakeHandlerService) Assign(tenant.TenantContext, string, core.AssignRolePermissionsRequest, []string) (*core.RolePermissionAssignmentResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	body := `{"permissions":["roles.read"]}`

	tests := []struct {
		name       string
		service    Service
		statusCode int
	}{
		{name: "success", service: fakeHandlerService{item: &core.RolePermissionAssignmentResponse{RoleID: "role-1"}}, statusCode: http.StatusOK},
		{name: "maps domain forbidden to 403", service: fakeHandlerService{err: core.ErrSystemImmutable}, statusCode: http.StatusForbidden},
		{name: "maps invalid slug to 400", service: fakeHandlerService{err: core.ErrInvalidPermissionSlug}, statusCode: http.StatusBadRequest},
		{name: "maps unknown to 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "role-1"}}
			tenant.SetGin(c, tenant.NewRoot())
			c.Request = httptest.NewRequest(http.MethodPatch, "/iam/roles/role-1/permissions", bytes.NewBufferString(body))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
