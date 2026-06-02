package delete_user

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
		service    Service
		statusCode int
	}{
		{name: "returns 204 on success", service: fakeHandlerService{}, statusCode: http.StatusNoContent},
		{name: "maps not found to 404", service: fakeHandlerService{err: core.ErrNotFound}, statusCode: http.StatusNotFound},
		{name: "maps unknown errors to 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "user-1"}}
			c.Request = httptest.NewRequest(http.MethodDelete, "/iam/users/user-1", nil)

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
