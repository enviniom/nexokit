package view_user

import (
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
	item *core.UserResponse
	err  error
}

func (f fakeHandlerService) GetByPublicID(tenant.TenantContext, string) (*core.UserResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.item, nil
}

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		service     Service
		statusCode  int
		wantMessage string
	}{
		{name: "returns success payload", service: fakeHandlerService{item: &core.UserResponse{PublicID: "user-1", Name: "Alice"}}, statusCode: http.StatusOK},
		{name: "maps not found to 404 with platform message", service: fakeHandlerService{err: core.ErrNotFound}, statusCode: http.StatusNotFound, wantMessage: "Recurso no encontrado"},
		{name: "maps unknown errors to 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(tt.service)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "id", Value: "user-1"}}
			c.Request = httptest.NewRequest(http.MethodGet, "/iam/users/user-1", nil)

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
