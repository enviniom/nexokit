package delete_role

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/modules/iam/core"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct{ err error }

func (f fakeHandlerService) Delete(tenant.TenantContext, string) error { return f.err }

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		statusCode int
	}{
		{name: "success 204", err: nil, statusCode: http.StatusNoContent},
		{name: "not found 404", err: core.ErrNotFound, statusCode: http.StatusNotFound},
		{name: "protected 403", err: core.ErrRoleProtected, statusCode: http.StatusForbidden},
		{name: "assigned users 422", err: core.ErrRoleHasAssignedUsers, statusCode: http.StatusUnprocessableEntity},
		{name: "generic error 500", err: errors.New("db down"), statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(fakeHandlerService{err: tt.err})
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "role-1"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/iam/roles/role-1", nil)

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
