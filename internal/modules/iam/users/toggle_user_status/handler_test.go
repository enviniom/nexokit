package toggle_user_status

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

func (f fakeHandlerService) Toggle(tenant.TenantContext, string, core.UpdateStatusRequest) (*core.UserResponse, error) {
	return f.resp, f.err
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		service    Service
		body       any
		statusCode int
	}{
		{
			name:       "returns success on toggle",
			service:    fakeHandlerService{resp: &core.UserResponse{PublicID: "user-1", Name: "Alice", IsActive: false}},
			body:       core.UpdateStatusRequest{IsActive: false},
			statusCode: http.StatusOK,
		},
		{
			name:       "maps not found to 404",
			service:    fakeHandlerService{err: core.ErrNotFound},
			body:       core.UpdateStatusRequest{IsActive: true},
			statusCode: http.StatusNotFound,
		},
		{
			name:       "maps unknown errors to 500",
			service:    fakeHandlerService{err: errors.New("db down")},
			body:       core.UpdateStatusRequest{IsActive: true},
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "returns 400 on invalid JSON",
			service:    fakeHandlerService{},
			body:       "not-json",
			statusCode: http.StatusBadRequest,
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
			c.Request = httptest.NewRequest(http.MethodPatch, "/iam/users/user-1/toggle-status", bytes.NewReader(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			h.Handle(c)

			if w.Code != tt.statusCode {
				t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}
