package list_selectable_roles

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/enviniom/nexokit/internal/platform/response"
	"github.com/enviniom/nexokit/internal/platform/tenant"
	"github.com/gin-gonic/gin"
)

type fakeHandlerService struct {
	items []response.SelectResponse
	err   error
}

func (f fakeHandlerService) List(_ tenant.TenantContext) ([]response.SelectResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
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

func TestHandlerHandle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		service    Service
		statusCode int
		assertBody func(t *testing.T, payload map[string]any)
	}{
		{
			name:       "success returns select payload",
			service:    fakeHandlerService{items: []response.SelectResponse{{ID: "role-1", Name: "Manager", Meta: map[string]any{"slug": "manager"}}}},
			statusCode: http.StatusOK,
			assertBody: func(t *testing.T, payload map[string]any) {
				if payload["success"] != true {
					t.Fatalf("expected success true")
				}
				data, ok := payload["data"].([]any)
				if !ok || len(data) != 1 {
					t.Fatalf("expected one selectable role")
				}
			},
		},
		{name: "service error returns 500", service: fakeHandlerService{err: errors.New("db down")}, statusCode: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		for _, scope := range []tenant.TenantContext{tenant.NewRoot(), tenant.NewScoped(7, "selected")} {
			t.Run(tt.name+scope.CompanySlug, func(t *testing.T) {
				h := NewHandler(tt.service)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/roles/select", nil)

				tenant.SetGin(c, scope)
				h.Handle(c)

				if w.Code != tt.statusCode {
					t.Fatalf("expected status %d, got %d", tt.statusCode, w.Code)
				}

				var payload map[string]any
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if tt.assertBody != nil {
					tt.assertBody(t, payload)
				}
			})
		}
	}
}
