package view_role_permission_catalog

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	data []core.RolePermissionGroupResponse
	err  error
}

func (f fakeHandlerService) View(tenant.TenantContext, string) ([]core.RolePermissionGroupResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.data, nil
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		service    Service
		statusCode int
	}{
		{name: "returns success payload", service: fakeHandlerService{data: []core.RolePermissionGroupResponse{{Module: "users"}}}, statusCode: http.StatusOK},
		{name: "maps not found to 404", service: fakeHandlerService{err: core.ErrNotFound}, statusCode: http.StatusNotFound},
		{name: "maps unknown errors to 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "role-1"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/iam/roles/role-1/permissions", nil)

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
